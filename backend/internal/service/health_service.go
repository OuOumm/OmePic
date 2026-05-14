package service

import (
	"context"
	"fmt"

	"omepic/backend/internal/cache"
	"omepic/backend/internal/repository"
)

type HealthStatus struct {
	Status string `json:"status"`
}

type HealthService struct {
	repo  *repository.Repository
	cache cache.ImageCache
}

func NewHealthService(repo *repository.Repository, imageCache cache.ImageCache) *HealthService {
	return &HealthService{repo: repo, cache: imageCache}
}

func (s *HealthService) Check(ctx context.Context) (HealthStatus, error) {
	if err := s.repo.Ping(ctx); err != nil {
		return HealthStatus{}, fmt.Errorf("%w: sqlite unavailable", ErrDependencyUnavailable)
	}
	if err := s.cache.Ping(ctx); err != nil {
		return HealthStatus{}, fmt.Errorf("%w: redis unavailable", ErrDependencyUnavailable)
	}
	return HealthStatus{Status: "ok"}, nil
}
