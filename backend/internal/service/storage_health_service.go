package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"path"
	"strings"
	"time"

	"omepic/backend/internal/model"
	"omepic/backend/internal/repository"
	"omepic/backend/internal/storage"
)

const (
	StorageHealthDefaultInterval = 5 * time.Minute
	storageHealthTimeout         = 10 * time.Second
	storageHealthProbeSize       = 1024
)

type StorageHealthService struct {
	repo    *repository.Repository
	storage interface {
		ForKey(string) (storage.ResolvedProvider, error)
	}
	logger *slog.Logger
}

func NewStorageHealthService(repo *repository.Repository, storageManager interface {
	ForKey(string) (storage.ResolvedProvider, error)
}, logger *slog.Logger) *StorageHealthService {
	if logger == nil {
		logger = slog.Default()
	}
	return &StorageHealthService{repo: repo, storage: storageManager, logger: logger}
}

func (s *StorageHealthService) List(ctx context.Context) ([]model.StorageHealthCheck, error) {
	configs, err := s.repo.ListStorageConfigs(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: storage config list failed", ErrDependencyUnavailable)
	}
	checks, err := s.repo.ListStorageHealthChecks(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: storage health list failed", ErrDependencyUnavailable)
	}
	byKey := make(map[string]model.StorageHealthCheck, len(checks))
	for _, check := range checks {
		byKey[check.StorageKey] = check
	}
	result := make([]model.StorageHealthCheck, 0, len(configs))
	for _, cfg := range configs {
		if check, ok := byKey[cfg.StorageKey]; ok {
			result = append(result, check)
			continue
		}
		result = append(result, model.StorageHealthCheck{
			StorageKey:   cfg.StorageKey,
			Status:       model.StorageHealthUnavailable,
			ErrorMessage: "not checked yet",
		})
	}
	return result, nil
}

func (s *StorageHealthService) Check(ctx context.Context, storageKey string) (model.StorageHealthCheck, error) {
	storageKey = strings.TrimSpace(storageKey)
	if storageKey == "" {
		return model.StorageHealthCheck{}, ErrInvalidInput
	}
	if s.storage == nil {
		return model.StorageHealthCheck{}, fmt.Errorf("%w: storage manager unavailable", ErrDependencyUnavailable)
	}
	resolved, err := s.storage.ForKey(storageKey)
	if err != nil {
		return model.StorageHealthCheck{}, fmt.Errorf("%w: storage config not found", ErrNotFound)
	}
	return s.checkResolved(ctx, resolved.Config.StorageKey, resolved.Provider)
}

func (s *StorageHealthService) CheckAll(ctx context.Context) ([]model.StorageHealthCheck, error) {
	configs, err := s.repo.ListStorageConfigs(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: storage config list failed", ErrDependencyUnavailable)
	}
	checks := make([]model.StorageHealthCheck, 0, len(configs))
	for _, cfg := range configs {
		check, err := s.Check(ctx, cfg.StorageKey)
		if err != nil {
			if err == ErrInvalidInput || err == ErrNotFound {
				return nil, err
			}
			s.logger.Warn("storage health check failed", "storage_key", cfg.StorageKey, "error", err.Error())
		}
		if check.StorageKey != "" {
			checks = append(checks, check)
		}
	}
	return checks, nil
}

func (s *StorageHealthService) History(ctx context.Context, storageKey string, since time.Time) ([]model.StorageHealthCheck, error) {
	storageKey = strings.TrimSpace(storageKey)
	if storageKey == "" {
		return nil, ErrInvalidInput
	}
	if since.IsZero() {
		since = time.Now().UTC().Add(-24 * time.Hour)
	}
	checks, err := s.repo.ListStorageHealthHistory(ctx, storageKey, since.UTC())
	if err != nil {
		return nil, fmt.Errorf("%w: storage health history failed", ErrDependencyUnavailable)
	}
	return checks, nil
}

func (s *StorageHealthService) StartHeartbeat(ctx context.Context, interval time.Duration) func() {
	if interval <= 0 {
		interval = StorageHealthDefaultInterval
	}
	heartbeatCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-heartbeatCtx.Done():
				return
			case <-ticker.C:
				func() {
					defer func() {
						if recovered := recover(); recovered != nil {
							s.logger.Error("storage health heartbeat panic", "panic", recovered)
						}
					}()
					if _, err := s.CheckAll(heartbeatCtx); err != nil {
						s.logger.Warn("storage health heartbeat failed", "error", err.Error())
					}
				}()
			}
		}
	}()
	return func() {
		cancel()
		<-done
	}
}

func (s *StorageHealthService) checkResolved(ctx context.Context, storageKey string, provider storage.Provider) (model.StorageHealthCheck, error) {
	probeCtx, cancel := context.WithTimeout(ctx, storageHealthTimeout)
	defer cancel()

	payload := make([]byte, storageHealthProbeSize)
	if _, err := rand.Read(payload); err != nil {
		return model.StorageHealthCheck{}, fmt.Errorf("%w: probe payload generation failed", ErrDependencyUnavailable)
	}
	objectKey := path.Join(".omepic-health", fmt.Sprintf("probe-%d.bin", time.Now().UTC().UnixNano()))
	started := time.Now()
	probeErr := executeStorageProbe(probeCtx, provider, objectKey, payload)
	latency := time.Since(started).Milliseconds()

	check := model.StorageHealthCheck{
		StorageKey: storageKey,
		Status:     model.StorageHealthHealthy,
		LatencyMS:  latency,
	}
	if probeErr != nil {
		check.Status = model.StorageHealthUnavailable
		check.ErrorMessage = safeStorageHealthError(probeErr)
	}
	stored, err := s.repo.UpsertStorageHealthCheck(ctx, check)
	if err != nil {
		return model.StorageHealthCheck{}, fmt.Errorf("%w: storage health save failed", ErrDependencyUnavailable)
	}
	return stored, nil
}

func executeStorageProbe(ctx context.Context, provider storage.Provider, objectKey string, payload []byte) error {
	if _, err := provider.Save(ctx, objectKey, payload, "application/octet-stream"); err != nil {
		return fmt.Errorf("probe write failed: %w", err)
	}
	defer func() { _ = provider.Delete(context.Background(), objectKey) }()
	opened, err := provider.Open(ctx, objectKey)
	if err != nil {
		return fmt.Errorf("probe read failed: %w", err)
	}
	readBack, err := io.ReadAll(opened.Reader)
	closeErr := opened.Reader.Close()
	if err != nil {
		return fmt.Errorf("probe read failed: %w", err)
	}
	if closeErr != nil {
		return fmt.Errorf("probe read close failed: %w", closeErr)
	}
	if !bytes.Equal(readBack, payload) {
		return fmt.Errorf("probe read mismatch")
	}
	if err := provider.Delete(ctx, objectKey); err != nil {
		return fmt.Errorf("probe cleanup failed: %w", err)
	}
	return nil
}

func safeStorageHealthError(err error) string {
	message := strings.TrimSpace(err.Error())
	if len(message) > 500 {
		message = message[:500]
	}
	return message
}
