package service

import (
	"context"
	"fmt"

	"omepic/backend/internal/cache"
	"omepic/backend/internal/model"
	"omepic/backend/internal/repository"
)

// md5MappingFlow owns the domain semantics for scoped original-byte MD5 mappings.
// Callers express storage-scoped lookup/repair intent and do not compose Redis keys.
type md5MappingRepository interface {
	FindByUID(ctx context.Context, uid string) (*model.ImageRecord, error)
	FindByMD5AndStorageKey(ctx context.Context, md5Hash string, storageKey string) (*model.ImageRecord, error)
	CountByMD5AndStorageKey(ctx context.Context, md5Hash string, storageKey string) (int64, error)
}

type md5MappingFlow struct {
	repo    md5MappingRepository
	cache   cache.MD5MappingCache
	preheat cache.MD5MappingPreheatCache
}

func (s *ImageService) md5Mappings() md5MappingFlow {
	return md5MappingFlow{repo: s.repo, cache: s.md5Cache, preheat: s.md5Preheat}
}

func md5MappingKeyForRecord(record model.ImageRecord) model.MD5MappingKey {
	return model.NewMD5MappingKey(record.StorageKey, record.MD5Hash)
}

func recordMatchesMD5Mapping(record *model.ImageRecord, key model.MD5MappingKey) bool {
	if record == nil {
		return false
	}
	return md5MappingKeyForRecord(*record) == key
}

func (m md5MappingFlow) FindReusableObject(ctx context.Context, key model.MD5MappingKey) (*model.ImageRecord, error) {
	cachedUID, err := m.cache.GetMD5(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("%w: redis md5 lookup failed", ErrDependencyUnavailable)
	}
	if cachedUID != "" {
		record, err := m.repo.FindByUID(ctx, cachedUID)
		switch {
		case err == nil && recordMatchesMD5Mapping(record, key):
			return record, nil
		case err == nil:
		case repository.IsNotFound(err):
		default:
			return nil, fmt.Errorf("%w: sqlite uid lookup failed", ErrDependencyUnavailable)
		}
	}

	record, err := m.repo.FindByMD5AndStorageKey(ctx, key.MD5Hash, key.StorageKey)
	if err != nil {
		if repository.IsNotFound(err) {
			if cachedUID != "" {
				if err := m.cache.DeleteMD5(ctx, key); err != nil {
					return nil, fmt.Errorf("%w: redis md5 stale delete failed", ErrDependencyUnavailable)
				}
			}
			return nil, nil
		}
		return nil, fmt.Errorf("%w: sqlite md5 lookup failed", ErrDependencyUnavailable)
	}
	if cachedUID != record.UID {
		if err := m.cache.SetMD5(ctx, key, record.UID); err != nil {
			return nil, fmt.Errorf("%w: redis md5 repair failed", ErrDependencyUnavailable)
		}
	}
	return record, nil
}

func (m md5MappingFlow) RememberNewPhysical(ctx context.Context, key model.MD5MappingKey, uid string) error {
	if err := m.cache.SetMD5IfAbsent(ctx, key, uid); err != nil {
		return fmt.Errorf("%w: redis md5 write failed", ErrDependencyUnavailable)
	}
	return nil
}

func (m md5MappingFlow) BackfillFromRecord(ctx context.Context, record model.ImageRecord) error {
	if err := m.cache.SetMD5IfAbsent(ctx, md5MappingKeyForRecord(record), record.UID); err != nil {
		return fmt.Errorf("%w: redis md5 repopulate failed", ErrDependencyUnavailable)
	}
	return nil
}

func (m md5MappingFlow) RepairAfterDelete(ctx context.Context, key model.MD5MappingKey, deletedUID string) error {
	md5Count, err := m.repo.CountByMD5AndStorageKey(ctx, key.MD5Hash, key.StorageKey)
	if err != nil {
		return fmt.Errorf("%w: md5 count failed", ErrDependencyUnavailable)
	}
	if md5Count == 0 {
		if err := m.cache.DeleteMD5(ctx, key); err != nil {
			return fmt.Errorf("%w: redis md5 delete failed", ErrDependencyUnavailable)
		}
		return nil
	}

	cachedUID, err := m.cache.GetMD5(ctx, key)
	if err != nil {
		return fmt.Errorf("%w: redis md5 lookup failed", ErrDependencyUnavailable)
	}
	if cachedUID != "" && cachedUID != deletedUID {
		record, err := m.repo.FindByUID(ctx, cachedUID)
		switch {
		case err == nil && recordMatchesMD5Mapping(record, key):
			return nil
		case err == nil:
		case repository.IsNotFound(err):
		default:
			return fmt.Errorf("%w: sqlite uid lookup failed", ErrDependencyUnavailable)
		}
	}

	replacement, err := m.repo.FindByMD5AndStorageKey(ctx, key.MD5Hash, key.StorageKey)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("%w: sqlite md5 lookup failed", ErrDependencyUnavailable)
	}
	if err := m.cache.SetMD5(ctx, key, replacement.UID); err != nil {
		return fmt.Errorf("%w: redis md5 repair failed", ErrDependencyUnavailable)
	}
	return nil
}

func (m md5MappingFlow) Preheat(ctx context.Context, records []model.ImageRecord) error {
	mappings := firstMD5MappingsByStorage(records)
	if err := m.preheat.SetMD5Mappings(ctx, mappings); err != nil {
		return fmt.Errorf("%w: redis md5 preheat failed", ErrDependencyUnavailable)
	}
	return nil
}

func firstMD5MappingsByStorage(records []model.ImageRecord) []model.MD5Mapping {
	seen := make(map[string]struct{}, len(records))
	mappings := make([]model.MD5Mapping, 0, len(records))
	for _, record := range records {
		key := md5MappingKeyForRecord(record)
		seenKey := key.MutexScope()
		if _, ok := seen[seenKey]; ok {
			continue
		}
		seen[seenKey] = struct{}{}
		mappings = append(mappings, model.MD5Mapping{Key: key, UID: record.UID})
	}
	return mappings
}
