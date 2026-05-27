package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"omepic/backend/internal/config"
	"omepic/backend/internal/model"
	"omepic/backend/internal/repository"
	"omepic/backend/internal/storage"
)

func TestStorageHealthCheckRecordsHealthyAndCleansProbe(t *testing.T) {
	repo := newStorageHealthTestRepo(t)
	provider := newMemoryHealthProvider()
	manager := &fakeStorageHealthManager{providers: map[string]storage.Provider{"local-default": provider}}
	service := NewStorageHealthService(repo, manager, slog.New(slog.NewTextHandler(io.Discard, nil)))

	check, err := service.Check(context.Background(), "local-default")
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if check.Status != model.StorageHealthHealthy || check.ConsecutiveFailures != 0 || check.LatencyMS < 0 || check.UpdatedAt.IsZero() {
		t.Fatalf("unexpected healthy check: %+v", check)
	}
	if provider.objectCount() != 0 {
		t.Fatalf("expected probe object to be cleaned up, got %d object(s)", provider.objectCount())
	}
}

func TestStorageHealthCheckRecordsFailureAndIncrementsConsecutiveFailures(t *testing.T) {
	repo := newStorageHealthTestRepo(t)
	provider := newMemoryHealthProvider()
	provider.saveErr = errors.New("write unavailable")
	manager := &fakeStorageHealthManager{providers: map[string]storage.Provider{"local-default": provider}}
	service := NewStorageHealthService(repo, manager, nil)

	first, err := service.Check(context.Background(), "local-default")
	if err != nil {
		t.Fatalf("first Check returned error: %v", err)
	}
	second, err := service.Check(context.Background(), "local-default")
	if err != nil {
		t.Fatalf("second Check returned error: %v", err)
	}
	if first.Status != model.StorageHealthUnavailable || first.ConsecutiveFailures != 1 || first.ErrorMessage == "" {
		t.Fatalf("unexpected first failure: %+v", first)
	}
	if second.ConsecutiveFailures != 2 {
		t.Fatalf("expected consecutive failures to increment to 2, got %+v", second)
	}

	provider.saveErr = nil
	healthy, err := service.Check(context.Background(), "local-default")
	if err != nil {
		t.Fatalf("healthy Check returned error: %v", err)
	}
	if healthy.Status != model.StorageHealthHealthy || healthy.ConsecutiveFailures != 0 || healthy.ErrorMessage != "" {
		t.Fatalf("expected healthy status to reset failures, got %+v", healthy)
	}
}

func TestStorageHealthHeartbeatCanStartAndStop(t *testing.T) {
	repo := newStorageHealthTestRepo(t)
	provider := newMemoryHealthProvider()
	manager := &fakeStorageHealthManager{providers: map[string]storage.Provider{"local-default": provider}}
	service := NewStorageHealthService(repo, manager, nil)

	stop := service.StartHeartbeat(context.Background(), 10*time.Millisecond)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		checks, err := service.List(context.Background())
		if err != nil {
			t.Fatalf("List returned error: %v", err)
		}
		if len(checks) == 1 && checks[0].UpdatedAt.IsZero() == false {
			stop()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	stop()
	t.Fatalf("heartbeat did not record a health check before deadline")
}

func newStorageHealthTestRepo(t *testing.T) *repository.Repository {
	t.Helper()
	repo, err := repository.New(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("repository.New returned error: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	ctx := context.Background()
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}
	_, err = repo.InitializeStorageCatalog(ctx, config.RuntimeStorageConfig{
		StorageKey:       "local-default",
		Name:             "Default Local Storage",
		IsDefault:        true,
		Backend:          config.StorageBackendLocal,
		LocalStoragePath: filepath.Join(t.TempDir(), "images"),
	})
	if err != nil {
		t.Fatalf("InitializeStorageCatalog returned error: %v", err)
	}
	return repo
}

type fakeStorageHealthManager struct {
	providers map[string]storage.Provider
}

func (m *fakeStorageHealthManager) ForKey(key string) (storage.ResolvedProvider, error) {
	provider, ok := m.providers[key]
	if !ok {
		return storage.ResolvedProvider{}, errors.New("missing provider")
	}
	return storage.ResolvedProvider{
		Config:   config.RuntimeStorageConfig{StorageKey: key, Backend: provider.Name()},
		Provider: provider,
	}, nil
}

type memoryHealthProvider struct {
	mu      sync.Mutex
	objects map[string][]byte
	saveErr error
	openErr error
}

func newMemoryHealthProvider() *memoryHealthProvider {
	return &memoryHealthProvider{objects: make(map[string][]byte)}
}

func (p *memoryHealthProvider) Name() string { return config.StorageBackendLocal }

func (p *memoryHealthProvider) Save(ctx context.Context, objectKey string, data []byte, contentType string) (string, error) {
	return p.SaveStream(ctx, objectKey, bytes.NewReader(data), int64(len(data)), contentType)
}

func (p *memoryHealthProvider) SaveStream(_ context.Context, objectKey string, reader io.Reader, _ int64, _ string) (string, error) {
	if p.saveErr != nil {
		return "", p.saveErr
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.objects[objectKey] = data
	return objectKey, nil
}

func (p *memoryHealthProvider) Open(_ context.Context, objectKey string) (storage.OpenResult, error) {
	if p.openErr != nil {
		return storage.OpenResult{}, p.openErr
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	data, ok := p.objects[objectKey]
	if !ok {
		return storage.OpenResult{}, errors.New("not found")
	}
	return storage.OpenResult{Reader: io.NopCloser(bytes.NewReader(data)), Size: int64(len(data)), ModTime: time.Now()}, nil
}

func (p *memoryHealthProvider) Delete(_ context.Context, objectKey string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.objects, objectKey)
	return nil
}

func (p *memoryHealthProvider) objectCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.objects)
}
