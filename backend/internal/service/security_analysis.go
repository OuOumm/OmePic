package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"omepic/backend/internal/model"
	"omepic/backend/internal/repository"
)

const maxAbuseRange = 90 * 24 * time.Hour

type securityAnalysis struct {
	repo *repository.Repository
}

func newSecurityAnalysis(repo *repository.Repository) securityAnalysis {
	return securityAnalysis{repo: repo}
}

func (a securityAnalysis) CreateIPBan(ctx context.Context, input AdminIPBanCreateInput) (AdminIPBanCreateResult, error) {
	uid := strings.TrimSpace(input.UID)
	ipAddress := strings.TrimSpace(input.IPAddress)
	if uid != "" {
		record, err := a.repo.FindByUID(ctx, uid)
		if err != nil {
			if repository.IsNotFound(err) {
				return AdminIPBanCreateResult{}, ErrNotFound
			}
			return AdminIPBanCreateResult{}, fmt.Errorf("%w: image lookup failed", ErrDependencyUnavailable)
		}
		ipAddress = strings.TrimSpace(record.IPAddress)
	}
	if ipAddress == "" {
		return AdminIPBanCreateResult{}, WithUserMessage(ErrInvalidInput, "uid or ip_address is required")
	}

	ban, err := a.activeBanByIP(ctx, ipAddress)
	if err == nil {
		return a.banCreateResult(ctx, ban)
	}
	if err != nil && !repository.IsNotFound(err) {
		return AdminIPBanCreateResult{}, fmt.Errorf("%w: ip ban lookup failed", ErrDependencyUnavailable)
	}

	ban, err = a.repo.CreateIPBan(ctx, model.IPBan{
		IPHash:          ipHash(ipAddress),
		IPAddress:       ipAddress,
		IPAddressMasked: maskIPAddress(ipAddress),
		Reason:          a.defaultReason(input.Reason, uid, ipAddress),
		ExpiresAt:       a.expiresAt(input.DurationHours),
	})
	if err != nil {
		return AdminIPBanCreateResult{}, fmt.Errorf("%w: ip ban create failed", ErrDependencyUnavailable)
	}
	return a.banCreateResult(ctx, ban)
}

func (a securityAnalysis) Overview(ctx context.Context, input AdminAbuseOverviewInput) (model.AbuseOverview, error) {
	from, to, err := a.normalizeRange(input.From, input.To)
	if err != nil {
		return model.AbuseOverview{}, err
	}
	uploadCount, uploadSize, err := a.repo.AbuseOverviewTotals(ctx, from, to)
	if err != nil {
		return model.AbuseOverview{}, fmt.Errorf("%w: abuse totals query failed", ErrDependencyUnavailable)
	}
	activeBanCount, err := a.repo.CountActiveIPBans(ctx)
	if err != nil {
		return model.AbuseOverview{}, fmt.Errorf("%w: active ip bans query failed", ErrDependencyUnavailable)
	}
	activeBans, err := a.repo.ActiveIPBansByHash(ctx)
	if err != nil {
		return model.AbuseOverview{}, fmt.Errorf("%w: active ip bans query failed", ErrDependencyUnavailable)
	}
	ipAggregates, err := a.repo.TopAbuseIPAggregates(ctx, from, to, 10)
	if err != nil {
		return model.AbuseOverview{}, fmt.Errorf("%w: abuse ip rank query failed", ErrDependencyUnavailable)
	}
	tokenAggregates, err := a.repo.TopAbuseTokenAggregates(ctx, from, to, 10)
	if err != nil {
		return model.AbuseOverview{}, fmt.Errorf("%w: abuse token rank query failed", ErrDependencyUnavailable)
	}
	return model.AbuseOverview{
		From:             from,
		To:               to,
		UploadCount:      uploadCount,
		UploadSize:       uploadSize,
		ActiveIPBanCount: activeBanCount,
		TopIPs:           a.annotateIPRank(ipAggregates, activeBans),
		TopTokens:        a.tokenRank(tokenAggregates),
	}, nil
}

func (a securityAnalysis) IPDetail(ctx context.Context, ipAddress string) (model.AbuseIPDetail, error) {
	trimmed := strings.TrimSpace(ipAddress)
	if trimmed == "" {
		return model.AbuseIPDetail{}, ErrInvalidInput
	}
	summary, err := a.repo.ImageSummaryByIP(ctx, trimmed)
	if err != nil {
		return model.AbuseIPDetail{}, fmt.Errorf("%w: abuse ip detail query failed", ErrDependencyUnavailable)
	}
	detail := model.AbuseIPDetail{
		IPAddress:       trimmed,
		IPAddressMasked: maskIPAddress(trimmed),
		UploadCount:     summary.Count,
		TotalSize:       summary.TotalSize,
	}
	ban, err := a.activeBanByIP(ctx, trimmed)
	if err == nil {
		detail.IsBanned = true
		detail.Ban = &ban
		return detail, nil
	}
	if repository.IsNotFound(err) {
		return detail, nil
	}
	return model.AbuseIPDetail{}, fmt.Errorf("%w: abuse ip detail query failed", ErrDependencyUnavailable)
}

func (a securityAnalysis) activeBanByIP(ctx context.Context, ipAddress string) (model.IPBan, error) {
	return a.repo.FindActiveIPBanByHash(ctx, ipHash(ipAddress))
}

func (a securityAnalysis) banCreateResult(ctx context.Context, ban model.IPBan) (AdminIPBanCreateResult, error) {
	summary, err := a.repo.ImageSummaryByIP(ctx, ban.IPAddress)
	if err != nil {
		return AdminIPBanCreateResult{}, fmt.Errorf("%w: ip image summary failed", ErrDependencyUnavailable)
	}
	return AdminIPBanCreateResult{Ban: ban, AffectedImageCount: summary.Count, AffectedTotalSize: summary.TotalSize}, nil
}

func (a securityAnalysis) defaultReason(reason string, uid string, ipAddress string) string {
	trimmed := strings.TrimSpace(reason)
	if trimmed != "" {
		return trimmed
	}
	if uid != "" {
		return "Abusive upload from image " + uid
	}
	return "Abusive upload from IP " + maskIPAddress(ipAddress)
}

func (a securityAnalysis) expiresAt(durationHours int) *time.Time {
	if durationHours <= 0 {
		return nil
	}
	expires := time.Now().UTC().Add(time.Duration(durationHours) * time.Hour)
	return &expires
}

func (a securityAnalysis) normalizeRange(from time.Time, to time.Time) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	if to.IsZero() {
		to = now
	}
	if from.IsZero() {
		from = to.Add(-24 * time.Hour)
	}
	from = from.UTC()
	to = to.UTC()
	if !from.Before(to) {
		return time.Time{}, time.Time{}, ErrInvalidInput
	}
	if to.Sub(from) > maxAbuseRange {
		return time.Time{}, time.Time{}, ErrInvalidInput
	}
	return from, to, nil
}

func (a securityAnalysis) annotateIPRank(aggregates []model.AbuseIPAggregate, activeBans map[string]model.IPBan) []model.AbuseIPRankItem {
	items := make([]model.AbuseIPRankItem, 0, len(aggregates))
	for _, aggregate := range aggregates {
		item := model.AbuseIPRankItem{
			IPAddress:       aggregate.IPAddress,
			IPAddressMasked: maskIPAddress(aggregate.IPAddress),
			UploadCount:     aggregate.UploadCount,
			TotalSize:       aggregate.TotalSize,
			LatestUploadAt:  aggregate.LatestUploadAt,
		}
		if ban, exists := activeBans[ipHash(aggregate.IPAddress)]; exists {
			item.IsBanned = true
			item.BanID = ban.ID
		}
		items = append(items, item)
	}
	return items
}

func (a securityAnalysis) tokenRank(aggregates []model.AbuseTokenAggregate) []model.AbuseTokenRankItem {
	items := make([]model.AbuseTokenRankItem, 0, len(aggregates))
	for _, aggregate := range aggregates {
		items = append(items, model.AbuseTokenRankItem{
			Token:          aggregate.Token,
			TokenPreview:   previewToken(aggregate.Token, 8),
			UploadCount:    aggregate.UploadCount,
			TotalSize:      aggregate.TotalSize,
			LatestUploadAt: aggregate.LatestUploadAt,
		})
	}
	return items
}

func previewToken(value string, max int) string {
	trimmed := strings.TrimSpace(value)
	if max < 1 || len(trimmed) <= max {
		return trimmed
	}
	return trimmed[:max] + "..."
}
