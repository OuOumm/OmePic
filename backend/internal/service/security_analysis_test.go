package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"omepic/backend/internal/model"
)

func TestCreateIPBanByUIDBuildsSummaryAndDefaultReason(t *testing.T) {
	ctx := context.Background()
	adminService, repo := newAdminServiceTestHarness(t)

	createdAt := time.Now().UTC().Add(-time.Hour)
	if err := repo.InsertImage(ctx, modelImageRecordWithIP("uid-ban", "local-default", "local", "203.0.113.9", 15, createdAt)); err != nil {
		t.Fatalf("InsertImage returned error: %v", err)
	}
	if err := repo.InsertImage(ctx, modelImageRecordWithIP("uid-same-ip", "local-default", "local", "203.0.113.9", 25, createdAt)); err != nil {
		t.Fatalf("InsertImage returned error: %v", err)
	}

	result, err := adminService.CreateIPBan(ctx, AdminIPBanCreateInput{UID: "uid-ban"})
	if err != nil {
		t.Fatalf("CreateIPBan returned error: %v", err)
	}
	if result.Ban.IPAddress != "203.0.113.9" {
		t.Fatalf("expected ban IP from UID image, got %q", result.Ban.IPAddress)
	}
	if !strings.Contains(result.Ban.Reason, "uid-ban") {
		t.Fatalf("expected default UID reason, got %q", result.Ban.Reason)
	}
	if result.AffectedImageCount != 2 || result.AffectedTotalSize != 40 {
		t.Fatalf("unexpected affected summary: count=%d size=%d", result.AffectedImageCount, result.AffectedTotalSize)
	}
	if result.Ban.IPHash == "" || result.Ban.IPAddressMasked == "" || result.Ban.IPAddressMasked == result.Ban.IPAddress {
		t.Fatalf("expected security module to fill hash and masked IP, got %+v", result.Ban)
	}
}

func TestCreateIPBanReturnsExistingActiveBanWithoutDuplicate(t *testing.T) {
	ctx := context.Background()
	adminService, repo := newAdminServiceTestHarness(t)

	createdAt := time.Now().UTC().Add(-time.Hour)
	if err := repo.InsertImage(ctx, modelImageRecordWithIP("uid-one", "local-default", "local", "198.51.100.10", 100, createdAt)); err != nil {
		t.Fatalf("InsertImage returned error: %v", err)
	}

	first, err := adminService.CreateIPBan(ctx, AdminIPBanCreateInput{IPAddress: "198.51.100.10", Reason: "first reason"})
	if err != nil {
		t.Fatalf("first CreateIPBan returned error: %v", err)
	}
	second, err := adminService.CreateIPBan(ctx, AdminIPBanCreateInput{IPAddress: "198.51.100.10", Reason: "second reason"})
	if err != nil {
		t.Fatalf("second CreateIPBan returned error: %v", err)
	}
	if second.Ban.ID != first.Ban.ID || second.Ban.Reason != "first reason" {
		t.Fatalf("expected existing active ban to be reused, first=%+v second=%+v", first.Ban, second.Ban)
	}
	bans, err := adminService.IPBans(ctx)
	if err != nil {
		t.Fatalf("IPBans returned error: %v", err)
	}
	if len(bans) != 1 {
		t.Fatalf("expected one stored ban, got %d", len(bans))
	}
	if second.AffectedImageCount != 1 || second.AffectedTotalSize != 100 {
		t.Fatalf("unexpected existing-ban summary: count=%d size=%d", second.AffectedImageCount, second.AffectedTotalSize)
	}
}

func TestAbuseOverviewAnnotatesActiveBanButNotExpiredBan(t *testing.T) {
	ctx := context.Background()
	adminService, repo := newAdminServiceTestHarness(t)

	now := time.Now().UTC()
	if err := repo.InsertImage(ctx, modelImageRecordWithIP("uid-active", "local-default", "local", "192.0.2.10", 10, now.Add(-2*time.Hour))); err != nil {
		t.Fatalf("InsertImage active returned error: %v", err)
	}
	if err := repo.InsertImage(ctx, modelImageRecordWithIP("uid-expired", "local-default", "local", "192.0.2.20", 20, now.Add(-time.Hour))); err != nil {
		t.Fatalf("InsertImage expired returned error: %v", err)
	}

	active, err := adminService.CreateIPBan(ctx, AdminIPBanCreateInput{IPAddress: "192.0.2.10"})
	if err != nil {
		t.Fatalf("CreateIPBan active returned error: %v", err)
	}
	expiredAt := now.Add(-time.Minute)
	if _, err := repo.CreateIPBan(ctx, model.IPBan{
		IPHash:          ipHash("192.0.2.20"),
		IPAddress:       "192.0.2.20",
		IPAddressMasked: maskIPAddress("192.0.2.20"),
		Reason:          "expired",
		ExpiresAt:       &expiredAt,
	}); err != nil {
		t.Fatalf("CreateIPBan expired fixture returned error: %v", err)
	}

	overview, err := adminService.AbuseOverview(ctx, AdminAbuseOverviewInput{From: now.Add(-24 * time.Hour), To: now.Add(time.Hour)})
	if err != nil {
		t.Fatalf("AbuseOverview returned error: %v", err)
	}
	if overview.ActiveIPBanCount != 1 {
		t.Fatalf("expected one active ban, got %d", overview.ActiveIPBanCount)
	}
	seenActive := false
	seenExpired := false
	for _, item := range overview.TopIPs {
		switch item.IPAddress {
		case "192.0.2.10":
			seenActive = true
			if !item.IsBanned || item.BanID != active.Ban.ID {
				t.Fatalf("expected active IP annotation with ban id %d, got %+v", active.Ban.ID, item)
			}
		case "192.0.2.20":
			seenExpired = true
			if item.IsBanned || item.BanID != 0 {
				t.Fatalf("expected expired ban not to annotate, got %+v", item)
			}
		}
	}
	if !seenActive || !seenExpired {
		t.Fatalf("expected both IPs in overview, got %+v", overview.TopIPs)
	}
}

func TestAbuseIPDetailIgnoresExpiredBanAndUsesMaskedIP(t *testing.T) {
	ctx := context.Background()
	adminService, repo := newAdminServiceTestHarness(t)

	now := time.Now().UTC()
	if err := repo.InsertImage(ctx, modelImageRecordWithIP("uid-detail", "local-default", "local", "198.51.100.55", 33, now)); err != nil {
		t.Fatalf("InsertImage returned error: %v", err)
	}
	expiredAt := now.Add(-time.Hour)
	if _, err := repo.CreateIPBan(ctx, model.IPBan{
		IPHash:          ipHash("198.51.100.55"),
		IPAddress:       "198.51.100.55",
		IPAddressMasked: maskIPAddress("198.51.100.55"),
		Reason:          "expired",
		ExpiresAt:       &expiredAt,
	}); err != nil {
		t.Fatalf("CreateIPBan fixture returned error: %v", err)
	}

	detail, err := adminService.AbuseIPDetail(ctx, " 198.51.100.55 ")
	if err != nil {
		t.Fatalf("AbuseIPDetail returned error: %v", err)
	}
	if detail.IsBanned || detail.Ban != nil {
		t.Fatalf("expected expired ban not to be shown in detail, got %+v", detail)
	}
	if detail.UploadCount != 1 || detail.TotalSize != 33 {
		t.Fatalf("unexpected detail summary: %+v", detail)
	}
	if detail.IPAddressMasked == "" || detail.IPAddressMasked == detail.IPAddress {
		t.Fatalf("expected masked IP in detail, got %+v", detail)
	}
}

func TestAbuseOverviewRejectsInvalidTimeWindows(t *testing.T) {
	ctx := context.Background()
	adminService, _ := newAdminServiceTestHarness(t)

	now := time.Now().UTC()
	cases := []AdminAbuseOverviewInput{
		{From: now, To: now},
		{From: now, To: now.Add(-time.Minute)},
		{From: now.Add(-91 * 24 * time.Hour), To: now},
	}
	for _, tc := range cases {
		_, err := adminService.AbuseOverview(ctx, tc)
		if err == nil || !containsError(err, ErrInvalidInput) {
			t.Fatalf("expected ErrInvalidInput for window %+v, got %v", tc, err)
		}
	}
}

func TestAbuseIPDetailRejectsEmptyIP(t *testing.T) {
	ctx := context.Background()
	adminService, _ := newAdminServiceTestHarness(t)

	_, err := adminService.AbuseIPDetail(ctx, "   ")
	if err == nil || !containsError(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput for empty IP detail, got %v", err)
	}
}

func modelImageRecordWithIP(uid string, storageKey string, backend string, ipAddress string, size int64, createdAt time.Time) model.ImageRecord {
	record := modelImageRecord(uid, storageKey, backend)
	record.IPAddress = ipAddress
	record.Size = size
	record.CreatedAt = createdAt
	return record
}
