package service

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"omepic/backend/internal/config"
	"omepic/backend/internal/model"
	"omepic/backend/internal/repository"
)

func TestMD5MappingFlowFindReusableObjectRepairsStaleCache(t *testing.T) {
	ctx := context.Background()
	flow, repo, cacheStore := newMD5MappingFlowTestHarness(t)
	key := model.NewMD5MappingKey("local-primary", "shared-md5")
	fallback := md5MappingFlowTestRecord("uid-sqlite", key.StorageKey, key.MD5Hash)
	insertMD5MappingFlowRecord(t, repo, fallback)

	tests := []struct {
		name      string
		seedCache func()
	}{
		{
			name: "missing UID falls back to SQLite and repairs cache",
			seedCache: func() {
				cacheStore.setCachedMD5(key, "missing-uid")
			},
		},
		{
			name: "cross-storage UID falls back to SQLite and repairs cache",
			seedCache: func() {
				insertMD5MappingFlowRecord(t, repo, md5MappingFlowTestRecord("uid-cross-storage", "local-secondary", key.MD5Hash))
				cacheStore.setCachedMD5(key, "uid-cross-storage")
			},
		},
		{
			name: "same-storage different-MD5 UID falls back to SQLite and repairs cache",
			seedCache: func() {
				insertMD5MappingFlowRecord(t, repo, md5MappingFlowTestRecord("uid-different-md5", key.StorageKey, "other-md5"))
				cacheStore.setCachedMD5(key, "uid-different-md5")
			},
		},
		{
			name:      "empty cache uses SQLite fallback and backfills cache",
			seedCache: func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cacheStore.clearMD5Mappings()
			tt.seedCache()

			record, err := flow.FindReusableObject(ctx, key)
			if err != nil {
				t.Fatalf("FindReusableObject returned error: %v", err)
			}
			if record == nil || record.UID != fallback.UID {
				t.Fatalf("expected SQLite fallback record %q, got %+v", fallback.UID, record)
			}
			if got, ok := cacheStore.cachedMD5(key); !ok || got != fallback.UID {
				t.Fatalf("expected cache repaired to %q, got %q ok=%v", fallback.UID, got, ok)
			}
		})
	}
}

func TestMD5MappingFlowFindReusableObjectDeletesStaleCacheWhenNoSQLiteFallback(t *testing.T) {
	ctx := context.Background()
	flow, _, cacheStore := newMD5MappingFlowTestHarness(t)
	key := model.NewMD5MappingKey("local-primary", "missing-md5")
	cacheStore.setCachedMD5(key, "missing-uid")

	record, err := flow.FindReusableObject(ctx, key)
	if err != nil {
		t.Fatalf("FindReusableObject returned error: %v", err)
	}
	if record != nil {
		t.Fatalf("expected no reusable record, got %+v", record)
	}
	if got, ok := cacheStore.cachedMD5(key); ok {
		t.Fatalf("expected stale md5 cache to be deleted, got %q", got)
	}
}

func TestMD5MappingFlowFindReusableObjectTrustsMatchingCachedUID(t *testing.T) {
	ctx := context.Background()
	flow, repo, cacheStore := newMD5MappingFlowTestHarness(t)
	key := model.NewMD5MappingKey("local-primary", "shared-md5")
	cached := md5MappingFlowTestRecord("uid-cached", key.StorageKey, key.MD5Hash)
	fallback := md5MappingFlowTestRecord("uid-sqlite-first", key.StorageKey, key.MD5Hash)
	insertMD5MappingFlowRecord(t, repo, cached)
	insertMD5MappingFlowRecord(t, repo, fallback)
	cacheStore.setCachedMD5(key, cached.UID)
	beforeSets, _ := cacheStore.stats()

	record, err := flow.FindReusableObject(ctx, key)
	if err != nil {
		t.Fatalf("FindReusableObject returned error: %v", err)
	}
	if record == nil || record.UID != cached.UID {
		t.Fatalf("expected matching cached uid %q to be reused, got %+v", cached.UID, record)
	}
	afterSets, _ := cacheStore.stats()
	if afterSets != beforeSets {
		t.Fatalf("expected no md5 repair write for a matching cached uid, before=%d after=%d", beforeSets, afterSets)
	}
}

func TestMD5MappingFlowFindReusableObjectErrorPaths(t *testing.T) {
	ctx := context.Background()
	key := model.NewMD5MappingKey("local-primary", "shared-md5")
	fallback := md5MappingFlowTestRecord("uid-sqlite", key.StorageKey, key.MD5Hash)

	tests := []struct {
		name      string
		flow      md5MappingFlow
		cache     *fakeMD5MappingCache
		wantPiece string
	}{
		{
			name: "cache md5 lookup failure is propagated",
			flow: md5MappingFlow{
				repo:  newFakeMD5MappingRepo(),
				cache: configuredMD5MappingCache(func(c *fakeMD5MappingCache) { c.getMD5Err = errors.New("redis get failed") }),
			},
			wantPiece: "redis md5 lookup failed",
		},
		{
			name: "cached uid sqlite lookup failure is propagated",
			flow: md5MappingFlow{
				repo: &fakeMD5MappingRepo{
					findByUIDErr: errors.New("sqlite uid failed"),
				},
				cache: configuredMD5MappingCache(func(c *fakeMD5MappingCache) { c.setCachedMD5(key, "uid-cached") }),
			},
			wantPiece: "sqlite uid lookup failed",
		},
		{
			name: "sqlite md5 fallback failure is propagated",
			flow: md5MappingFlow{
				repo: &fakeMD5MappingRepo{
					findByMD5Err: errors.New("sqlite md5 failed"),
				},
				cache: newFakeMD5MappingCache(),
			},
			wantPiece: "sqlite md5 lookup failed",
		},
		{
			name: "cache repair set failure is propagated",
			flow: md5MappingFlow{
				repo:  fakeMD5MappingRepoWithRecords(fallback),
				cache: configuredMD5MappingCache(func(c *fakeMD5MappingCache) { c.setMD5Err = errors.New("redis set failed") }),
			},
			wantPiece: "redis md5 repair failed",
		},
		{
			name: "stale cache delete failure is propagated",
			flow: md5MappingFlow{
				repo: newFakeMD5MappingRepo(),
				cache: configuredMD5MappingCache(func(c *fakeMD5MappingCache) {
					c.setCachedMD5(key, "missing-uid")
					c.deleteMD5Err = errors.New("redis delete failed")
				}),
			},
			wantPiece: "redis md5 stale delete failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record, err := tt.flow.FindReusableObject(ctx, key)
			assertDependencyErrorContains(t, err, tt.wantPiece)
			if record != nil {
				t.Fatalf("expected no reusable record on error, got %+v", record)
			}
		})
	}
}

func TestMD5MappingFlowRememberNewPhysicalWritesFirstPhysicalUIDMapping(t *testing.T) {
	ctx := context.Background()
	flow, _, cacheStore := newMD5MappingFlowTestHarness(t)
	key := model.NewMD5MappingKey("local-primary", "shared-md5")

	if err := flow.RememberNewPhysical(ctx, key, "uid-first-physical"); err != nil {
		t.Fatalf("RememberNewPhysical returned error: %v", err)
	}

	if got, ok := cacheStore.cachedMD5(key); !ok || got != "uid-first-physical" {
		t.Fatalf("expected first physical uid mapping %q, got %q ok=%v", "uid-first-physical", got, ok)
	}
	md5Sets, _ := cacheStore.stats()
	if md5Sets != 1 {
		t.Fatalf("expected one md5 SetNX write, got %d", md5Sets)
	}
}

func TestMD5MappingFlowRememberNewPhysicalPropagatesSetNXFailure(t *testing.T) {
	ctx := context.Background()
	key := model.NewMD5MappingKey("local-primary", "shared-md5")
	cacheStore := configuredMD5MappingCache(func(c *fakeMD5MappingCache) {
		c.setMD5IfAbsentErr = errors.New("redis setnx failed")
	})
	flow := md5MappingFlow{repo: newFakeMD5MappingRepo(), cache: cacheStore}

	err := flow.RememberNewPhysical(ctx, key, "uid-first-physical")
	assertDependencyErrorContains(t, err, "redis md5 write failed")
	if got, ok := cacheStore.cachedMD5(key); ok {
		t.Fatalf("expected failed SetNX not to write mapping, got %q", got)
	}
}

func TestMD5MappingFlowBackfillFromRecordWritesMapping(t *testing.T) {
	ctx := context.Background()
	flow, _, cacheStore := newMD5MappingFlowTestHarness(t)
	record := md5MappingFlowTestRecord("uid-backfill", "local-primary", "shared-md5")
	key := md5MappingKeyForRecord(record)

	if err := flow.BackfillFromRecord(ctx, record); err != nil {
		t.Fatalf("BackfillFromRecord returned error: %v", err)
	}

	if got, ok := cacheStore.cachedMD5(key); !ok || got != record.UID {
		t.Fatalf("expected backfilled uid mapping %q, got %q ok=%v", record.UID, got, ok)
	}
	md5Sets, _ := cacheStore.stats()
	if md5Sets != 1 {
		t.Fatalf("expected one md5 SetNX write, got %d", md5Sets)
	}
}

func TestMD5MappingFlowBackfillFromRecordPropagatesSetNXFailure(t *testing.T) {
	ctx := context.Background()
	record := md5MappingFlowTestRecord("uid-backfill", "local-primary", "shared-md5")
	key := md5MappingKeyForRecord(record)
	cacheStore := configuredMD5MappingCache(func(c *fakeMD5MappingCache) {
		c.setMD5IfAbsentErr = errors.New("redis setnx failed")
	})
	flow := md5MappingFlow{repo: newFakeMD5MappingRepo(), cache: cacheStore}

	err := flow.BackfillFromRecord(ctx, record)
	assertDependencyErrorContains(t, err, "redis md5 repopulate failed")
	if got, ok := cacheStore.cachedMD5(key); ok {
		t.Fatalf("expected failed SetNX not to write mapping, got %q", got)
	}
}

func TestMD5MappingFlowRepairAfterDeleteErrorPaths(t *testing.T) {
	ctx := context.Background()
	key := model.NewMD5MappingKey("local-primary", "shared-md5")
	survivor := md5MappingFlowTestRecord("uid-survivor", key.StorageKey, key.MD5Hash)

	tests := []struct {
		name      string
		flow      md5MappingFlow
		wantPiece string
	}{
		{
			name: "sqlite md5 count failure is propagated",
			flow: md5MappingFlow{
				repo: &fakeMD5MappingRepo{
					countByMD5Err: errors.New("sqlite count failed"),
				},
				cache: newFakeMD5MappingCache(),
			},
			wantPiece: "md5 count failed",
		},
		{
			name: "cache delete failure when no survivor is propagated",
			flow: md5MappingFlow{
				repo:  newFakeMD5MappingRepo(),
				cache: configuredMD5MappingCache(func(c *fakeMD5MappingCache) { c.deleteMD5Err = errors.New("redis delete failed") }),
			},
			wantPiece: "redis md5 delete failed",
		},
		{
			name: "cache md5 lookup failure is propagated",
			flow: md5MappingFlow{
				repo:  fakeMD5MappingRepoWithRecords(survivor),
				cache: configuredMD5MappingCache(func(c *fakeMD5MappingCache) { c.getMD5Err = errors.New("redis get failed") }),
			},
			wantPiece: "redis md5 lookup failed",
		},
		{
			name: "cached survivor sqlite lookup failure is propagated",
			flow: md5MappingFlow{
				repo: &fakeMD5MappingRepo{
					records:      []model.ImageRecord{survivor},
					findByUIDErr: errors.New("sqlite uid failed"),
				},
				cache: configuredMD5MappingCache(func(c *fakeMD5MappingCache) { c.setCachedMD5(key, "uid-cached") }),
			},
			wantPiece: "sqlite uid lookup failed",
		},
		{
			name: "sqlite survivor lookup failure is propagated",
			flow: md5MappingFlow{
				repo: &fakeMD5MappingRepo{
					records:      []model.ImageRecord{survivor},
					findByMD5Err: errors.New("sqlite md5 failed"),
				},
				cache: newFakeMD5MappingCache(),
			},
			wantPiece: "sqlite md5 lookup failed",
		},
		{
			name: "cache repair set failure is propagated",
			flow: md5MappingFlow{
				repo:  fakeMD5MappingRepoWithRecords(survivor),
				cache: configuredMD5MappingCache(func(c *fakeMD5MappingCache) { c.setMD5Err = errors.New("redis set failed") }),
			},
			wantPiece: "redis md5 repair failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.flow.RepairAfterDelete(ctx, key, "uid-deleted")
			assertDependencyErrorContains(t, err, tt.wantPiece)
		})
	}
}

func TestMD5MappingFlowRepairAfterDeleteStaleCacheMatrix(t *testing.T) {
	ctx := context.Background()
	flow, repo, cacheStore := newMD5MappingFlowTestHarness(t)
	key := model.NewMD5MappingKey("local-primary", "shared-md5")

	tests := []struct {
		name        string
		seedRecords func()
		seedCache   func()
		deletedUID  string
		wantUID     string
		wantPresent bool
	}{
		{
			name: "missing cached UID repoints to surviving same-storage record",
			seedRecords: func() {
				insertMD5MappingFlowRecord(t, repo, md5MappingFlowTestRecord("uid-survivor", key.StorageKey, key.MD5Hash))
			},
			seedCache: func() {
				cacheStore.setCachedMD5(key, "missing-uid")
			},
			deletedUID:  "uid-deleted",
			wantUID:     "uid-survivor",
			wantPresent: true,
		},
		{
			name: "cross-storage cached UID repoints within same storage only",
			seedRecords: func() {
				insertMD5MappingFlowRecord(t, repo, md5MappingFlowTestRecord("uid-cross-storage", "local-secondary", key.MD5Hash))
				insertMD5MappingFlowRecord(t, repo, md5MappingFlowTestRecord("uid-survivor", key.StorageKey, key.MD5Hash))
			},
			seedCache: func() {
				cacheStore.setCachedMD5(key, "uid-cross-storage")
			},
			deletedUID:  "uid-deleted",
			wantUID:     "uid-survivor",
			wantPresent: true,
		},
		{
			name: "same-storage different-MD5 cached UID repoints to same MD5 survivor",
			seedRecords: func() {
				insertMD5MappingFlowRecord(t, repo, md5MappingFlowTestRecord("uid-different-md5", key.StorageKey, "other-md5"))
				insertMD5MappingFlowRecord(t, repo, md5MappingFlowTestRecord("uid-survivor", key.StorageKey, key.MD5Hash))
			},
			seedCache: func() {
				cacheStore.setCachedMD5(key, "uid-different-md5")
			},
			deletedUID:  "uid-deleted",
			wantUID:     "uid-survivor",
			wantPresent: true,
		},
		{
			name: "valid different cached survivor is kept",
			seedRecords: func() {
				insertMD5MappingFlowRecord(t, repo, md5MappingFlowTestRecord("uid-cached-survivor", key.StorageKey, key.MD5Hash))
				insertMD5MappingFlowRecord(t, repo, md5MappingFlowTestRecord("uid-sqlite-first", key.StorageKey, key.MD5Hash))
			},
			seedCache: func() {
				cacheStore.setCachedMD5(key, "uid-cached-survivor")
			},
			deletedUID:  "uid-deleted",
			wantUID:     "uid-cached-survivor",
			wantPresent: true,
		},
		{
			name:        "no same-storage survivor clears stale mapping",
			seedRecords: func() {},
			seedCache: func() {
				cacheStore.setCachedMD5(key, "missing-uid")
			},
			deletedUID:  "uid-deleted",
			wantPresent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetMD5MappingFlowRecords(t, repo)
			cacheStore.clearMD5Mappings()
			tt.seedRecords()
			tt.seedCache()

			if err := flow.RepairAfterDelete(ctx, key, tt.deletedUID); err != nil {
				t.Fatalf("RepairAfterDelete returned error: %v", err)
			}
			got, ok := cacheStore.cachedMD5(key)
			if ok != tt.wantPresent || got != tt.wantUID {
				t.Fatalf("expected cache present=%v uid=%q, got present=%v uid=%q", tt.wantPresent, tt.wantUID, ok, got)
			}
		})
	}
}

func TestMD5MappingFlowPreheatKeepsMappingsScopedByStorage(t *testing.T) {
	ctx := context.Background()
	flow, _, cacheStore := newMD5MappingFlowTestHarness(t)
	records := []model.ImageRecord{
		md5MappingFlowTestRecord("uid-primary-first", "local-primary", "shared-md5"),
		md5MappingFlowTestRecord("uid-primary-second", "local-primary", "shared-md5"),
		md5MappingFlowTestRecord("uid-secondary-first", "local-secondary", "shared-md5"),
	}

	if err := flow.Preheat(ctx, records); err != nil {
		t.Fatalf("Preheat returned error: %v", err)
	}

	primaryKey := model.NewMD5MappingKey("local-primary", "shared-md5")
	secondaryKey := model.NewMD5MappingKey("local-secondary", "shared-md5")
	if got, _ := cacheStore.cachedMD5(primaryKey); got != "uid-primary-first" {
		t.Fatalf("expected primary mapping to first primary uid, got %q", got)
	}
	if got, _ := cacheStore.cachedMD5(secondaryKey); got != "uid-secondary-first" {
		t.Fatalf("expected secondary mapping to first secondary uid, got %q", got)
	}
	md5Sets, md5BatchSets := cacheStore.stats()
	if md5Sets != 2 || md5BatchSets != 1 {
		t.Fatalf("expected two scoped md5 writes in one batch, got writes=%d batches=%d", md5Sets, md5BatchSets)
	}
}

func TestMD5MappingFlowPreheatPropagatesBatchWriteFailure(t *testing.T) {
	ctx := context.Background()
	cacheStore := configuredMD5MappingCache(func(c *fakeMD5MappingCache) {
		c.setMD5MappingsErr = errors.New("redis batch set failed")
	})
	flow := md5MappingFlow{repo: newFakeMD5MappingRepo(), cache: cacheStore, preheat: cacheStore}
	records := []model.ImageRecord{
		md5MappingFlowTestRecord("uid-primary-first", "local-primary", "shared-md5"),
		md5MappingFlowTestRecord("uid-primary-second", "local-primary", "shared-md5"),
		md5MappingFlowTestRecord("uid-secondary-first", "local-secondary", "shared-md5"),
	}

	err := flow.Preheat(ctx, records)
	assertDependencyErrorContains(t, err, "redis md5 preheat failed")
	primaryKey := model.NewMD5MappingKey("local-primary", "shared-md5")
	if got, ok := cacheStore.cachedMD5(primaryKey); ok {
		t.Fatalf("expected failed batch preheat not to write mappings, got %q", got)
	}
}

func newMD5MappingFlowTestHarness(t *testing.T) (md5MappingFlow, *repository.Repository, *fakeMD5MappingCache) {
	t.Helper()

	repo, err := repository.New(filepath.Join(t.TempDir(), "test.sqlite"))
	if err != nil {
		t.Fatalf("repository.New returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = repo.Close()
	})
	if err := repo.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}
	cacheStore := newFakeMD5MappingCache()
	return md5MappingFlow{repo: repo, cache: cacheStore, preheat: cacheStore}, repo, cacheStore
}

func md5MappingFlowTestRecord(uid string, storageKey string, md5Hash string) model.ImageRecord {
	return model.ImageRecord{
		UID:            uid,
		Token:          "token-" + uid,
		StorageKey:     storageKey,
		StorageBackend: config.StorageBackendLocal,
		FilePath:       "2026/05/" + uid + publicImageExtension,
		MIMEType:       publicImageMIMEType,
		Size:           1,
		MD5Hash:        md5Hash,
		IPAddress:      "127.0.0.1",
		CreatedAt:      time.Now().UTC(),
	}
}

func insertMD5MappingFlowRecord(t *testing.T, repo *repository.Repository, record model.ImageRecord) {
	t.Helper()
	if err := repo.InsertImage(context.Background(), record); err != nil {
		t.Fatalf("InsertImage(%q) returned error: %v", record.UID, err)
	}
}

func resetMD5MappingFlowRecords(t *testing.T, repo *repository.Repository) {
	t.Helper()
	records, err := repo.ListAllImages(context.Background())
	if err != nil {
		t.Fatalf("ListAllImages returned error: %v", err)
	}
	for _, record := range records {
		if err := repo.DeleteByUID(context.Background(), record.UID); err != nil {
			t.Fatalf("DeleteByUID(%q) returned error: %v", record.UID, err)
		}
	}
}

func configuredMD5MappingCache(configure func(*fakeMD5MappingCache)) *fakeMD5MappingCache {
	cacheStore := newFakeMD5MappingCache()
	if configure != nil {
		configure(cacheStore)
	}
	return cacheStore
}

func assertDependencyErrorContains(t *testing.T, err error, wantPiece string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q", wantPiece)
	}
	if !errors.Is(err, ErrDependencyUnavailable) {
		t.Fatalf("expected ErrDependencyUnavailable, got %v", err)
	}
	if !strings.Contains(err.Error(), wantPiece) {
		t.Fatalf("expected error %q to contain %q", err.Error(), wantPiece)
	}
}

type fakeMD5MappingRepo struct {
	records       []model.ImageRecord
	findByUIDErr  error
	findByMD5Err  error
	countByMD5Err error
}

func newFakeMD5MappingRepo() *fakeMD5MappingRepo {
	return &fakeMD5MappingRepo{}
}

func fakeMD5MappingRepoWithRecords(records ...model.ImageRecord) *fakeMD5MappingRepo {
	return &fakeMD5MappingRepo{records: records}
}

func (r *fakeMD5MappingRepo) FindByUID(_ context.Context, uid string) (*model.ImageRecord, error) {
	if r.findByUIDErr != nil {
		return nil, r.findByUIDErr
	}
	for _, record := range r.records {
		if record.UID == uid {
			copy := record
			return &copy, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *fakeMD5MappingRepo) FindByMD5AndStorageKey(_ context.Context, md5Hash string, storageKey string) (*model.ImageRecord, error) {
	if r.findByMD5Err != nil {
		return nil, r.findByMD5Err
	}
	for _, record := range r.records {
		if record.MD5Hash == md5Hash && record.StorageKey == storageKey {
			copy := record
			return &copy, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (r *fakeMD5MappingRepo) CountByMD5AndStorageKey(_ context.Context, md5Hash string, storageKey string) (int64, error) {
	if r.countByMD5Err != nil {
		return 0, r.countByMD5Err
	}
	var count int64
	for _, record := range r.records {
		if record.MD5Hash == md5Hash && record.StorageKey == storageKey {
			count++
		}
	}
	return count, nil
}
