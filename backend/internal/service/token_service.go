package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"omepic/backend/internal/model"
	"omepic/backend/internal/repository"
)

type TokenService struct {
	repo *repository.Repository
}

type TokenListResult struct {
	Items []model.TokenGovernanceEntry `json:"items"`
}

type TokenDisableInput struct {
	Reason string `json:"reason"`
}

func NewTokenService(repo *repository.Repository) *TokenService {
	return &TokenService{repo: repo}
}

func TokenHash(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

func TokenPreview(token string) string {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return ""
	}
	runes := []rune(trimmed)
	if len(runes) <= 4 {
		return strings.Repeat("*", len(runes))
	}
	return strings.Repeat("*", len(runes)-4) + string(runes[len(runes)-4:])
}

func (s *TokenService) EnsureTokenAllowed(ctx context.Context, token string) error {
	if s == nil || s.repo == nil {
		return nil
	}
	control, err := s.repo.FindTokenControl(ctx, TokenHash(token))
	if err != nil {
		if repository.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("%w: token control lookup failed", ErrDependencyUnavailable)
	}
	if control.Disabled {
		return WithUserMessage(ErrForbidden, "token is disabled")
	}
	return nil
}

func (s *TokenService) RecordUpload(ctx context.Context, token string, size int64, ipAddress string, usedAt time.Time) error {
	if s == nil || s.repo == nil {
		return nil
	}
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return nil
	}
	if usedAt.IsZero() {
		usedAt = time.Now().UTC()
	}
	if size < 0 {
		size = 0
	}
	if err := s.repo.UpsertTokenUsage(ctx, model.TokenUsage{
		TokenHash:    TokenHash(trimmed),
		TokenPreview: TokenPreview(trimmed),
		TotalBytes:   size,
		LastIP:       strings.TrimSpace(ipAddress),
		LastUsedAt:   usedAt,
	}); err != nil {
		return fmt.Errorf("%w: token usage save failed", ErrDependencyUnavailable)
	}
	return nil
}

func (s *TokenService) List(ctx context.Context) (TokenListResult, error) {
	items, err := s.repo.ListTokenGovernance(ctx)
	if err != nil {
		return TokenListResult{}, fmt.Errorf("%w: token list query failed", ErrDependencyUnavailable)
	}
	return TokenListResult{Items: items}, nil
}

func (s *TokenService) Disable(ctx context.Context, tokenHash string, reason string) error {
	return s.setDisabled(ctx, tokenHash, true, reason)
}

func (s *TokenService) Enable(ctx context.Context, tokenHash string) error {
	return s.setDisabled(ctx, tokenHash, false, "")
}

func (s *TokenService) setDisabled(ctx context.Context, tokenHash string, disabled bool, reason string) error {
	hash := strings.ToLower(strings.TrimSpace(tokenHash))
	if len(hash) != 64 || !isLowerHex(hash) {
		return WithUserMessage(ErrInvalidInput, "token_hash must be a SHA-256 hex value")
	}
	preview := ""
	if existing, err := s.repo.FindTokenControl(ctx, hash); err == nil {
		preview = existing.TokenPreview
	} else if !repository.IsNotFound(err) {
		return fmt.Errorf("%w: token control lookup failed", ErrDependencyUnavailable)
	}
	if preview == "" {
		if entries, err := s.repo.ListTokenGovernance(ctx); err == nil {
			for _, entry := range entries {
				if entry.TokenHash == hash {
					preview = entry.TokenPreview
					break
				}
			}
		} else {
			return fmt.Errorf("%w: token list query failed", ErrDependencyUnavailable)
		}
	}
	if err := s.repo.SetTokenDisabled(ctx, hash, preview, strings.TrimSpace(reason), disabled); err != nil {
		return fmt.Errorf("%w: token control save failed", ErrDependencyUnavailable)
	}
	return nil
}

func isLowerHex(value string) bool {
	for _, r := range value {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}
