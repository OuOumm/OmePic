"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import toast from "react-hot-toast";
import { AlertCircle, Ban, Check, Clock, HardDrive, Loader2, RefreshCw, ShieldOff, Trash2, Upload } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/Card";
import { Input } from "@/components/ui/Input";
import { Label } from "@/components/ui/Label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/Table";
import {
  adminCreateIPBan,
  adminDeleteIPBan,
  adminDeleteIPBanImages,
  adminGetAbuseOverview,
  adminGetIPBans,
  adminGetSystemSettings,
  adminUpdateSystemSettings,
} from "@/lib/api";
import { t } from "@/lib/i18n";
import { formatBytes } from "@/lib/utils";
import { DEFAULT_RUNTIME_SETTINGS, normalizeRuntimeSettings } from "./runtime-settings";
import { AdminPageHeader } from "./AdminPageHeader";
import { useAdminSessionStore } from "@/stores/admin-session-store";
import { useUiPreferencesStore } from "@/stores/ui-preferences-store";
import type { AdminAbuseIPRankItem, AdminAbuseOverview, AdminIPBan, RuntimeSettings } from "@/types";

const defaultRangeHours = 24;

export function SecurityAbusePageClient() {
  const token = useAdminSessionStore((state) => state.token);
  const language = useUiPreferencesStore((state) => state.language);
  const [overview, setOverview] = useState<AdminAbuseOverview | null>(null);
  const [bans, setBans] = useState<AdminIPBan[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [fromInput, setFromInput] = useState(() => toDateTimeLocal(new Date(Date.now() - defaultRangeHours * 60 * 60 * 1000)));
  const [toInput, setToInput] = useState(() => toDateTimeLocal(new Date()));
  const [visibleIPs, setVisibleIPs] = useState<Set<string>>(new Set());
  const [busyKey, setBusyKey] = useState<string | null>(null);
  const [runtimeForm, setRuntimeForm] = useState<RuntimeSettings>(DEFAULT_RUNTIME_SETTINGS);
  const [runtimeSaving, setRuntimeSaving] = useState(false);

  const lang = language;

  const loadData = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    setError("");
    try {
      const from = fromInput ? new Date(fromInput).toISOString() : undefined;
      const to = toInput ? new Date(toInput).toISOString() : undefined;
      const [nextOverview, nextBans, settings] = await Promise.all([
        adminGetAbuseOverview(token, from, to),
        adminGetIPBans(token),
        adminGetSystemSettings(token),
      ]);
      setOverview(normalizeOverview(nextOverview));
      setBans(Array.isArray(nextBans) ? nextBans : []);
      setRuntimeForm(normalizeRuntimeSettings(settings.runtime));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load security data");
    } finally {
      setLoading(false);
    }
  }, [token, fromInput, toInput]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const activeBanIds = useMemo(() => new Set((Array.isArray(bans) ? bans : []).filter(isActiveBan).map((ban) => ban.id)), [bans]);

  const updateRuntimeField = <K extends keyof RuntimeSettings>(field: K, value: RuntimeSettings[K]) => {
    setRuntimeForm((prev) => ({ ...prev, [field]: value }));
  };

  const handleSaveRateLimits = async () => {
    if (!token) return;
    setRuntimeSaving(true);
    try {
      const saved = await adminUpdateSystemSettings(token, runtimeForm);
      setRuntimeForm(normalizeRuntimeSettings(saved.runtime));
      toast.success(t(lang, "admin.settingsSaved"));
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(lang, "common.error"));
    } finally {
      setRuntimeSaving(false);
    }
  };

  const toggleIP = (ip: string) => {
    setVisibleIPs((prev) => {
      const next = new Set(prev);
      if (next.has(ip)) {
        next.delete(ip);
      } else {
        next.add(ip);
      }
      return next;
    });
  };

  const displayIP = (ip: string, masked: string) => visibleIPs.has(ip) ? ip : masked || ip;

  const handleBanIP = async (item: AdminAbuseIPRankItem) => {
    if (!token) return;
    const durationInput = window.prompt(t(lang, "admin.ipBanDurationPrompt"), "24");
    if (durationInput === null) return;
    const durationHours = Number(durationInput.trim());
    if (!Number.isFinite(durationHours) || durationHours < 0) {
      toast.error(t(lang, "admin.ipBanInvalidDuration"));
      return;
    }
    if (!window.confirm(t(lang, "admin.ipBanConfirm", { ip: item.ip_address_masked || item.ip_address }))) return;
    setBusyKey(`ban:${item.ip_address}`);
    try {
      await adminCreateIPBan(token, {
        ip_address: item.ip_address,
        duration_hours: durationHours,
        reason: `Abusive upload from IP ${item.ip_address_masked || item.ip_address}`,
      });
      toast.success(t(lang, "admin.ipBanCreated"));
      loadData();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(lang, "admin.ipBanCreateError"));
    } finally {
      setBusyKey(null);
    }
  };

  const handleUnban = async (ban: AdminIPBan) => {
    if (!token) return;
    if (!window.confirm(t(lang, "admin.abuseUnbanConfirm", { ip: ban.ip_address_masked || ban.ip_address }))) return;
    setBusyKey(`unban:${ban.id}`);
    try {
      await adminDeleteIPBan(token, ban.id);
      toast.success(t(lang, "admin.abuseUnbanned"));
      loadData();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(lang, "admin.abuseUnbanError"));
    } finally {
      setBusyKey(null);
    }
  };

  const handleDeleteBanImages = async (ban: AdminIPBan) => {
    if (!token) return;
    if (!window.confirm(t(lang, "admin.abuseDeleteBanImagesConfirm", { ip: ban.ip_address_masked || ban.ip_address }))) return;
    setBusyKey(`delete:${ban.id}`);
    try {
      const result = await adminDeleteIPBanImages(token, ban.id);
      toast.success(t(lang, "admin.ipBanImagesDeleted", { count: result.deleted_count }));
      loadData();
    } catch (err) {
      toast.error(err instanceof Error ? err.message : t(lang, "admin.imagesDeleteError"));
    } finally {
      setBusyKey(null);
    }
  };

  const banForIP = (ip: string) => (Array.isArray(bans) ? bans : []).find((ban) => isActiveBan(ban) && ban.ip_address === ip);

  if (loading && !overview) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error && !overview) {
    return (
      <div className="flex items-center gap-2 text-destructive" role="alert">
        <AlertCircle className="h-5 w-5" />
        {error}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <AdminPageHeader
        eyebrow={t(lang, "admin.sidebarSecurity")}
        title={t(lang, "admin.abuseTitle")}
        description={t(lang, "admin.abuseDescription")}
        actions={(
          <>
            <label className="grid gap-1 text-xs text-muted-foreground">
              {t(lang, "admin.abuseFrom")}
              <Input type="datetime-local" value={fromInput} onChange={(event) => setFromInput(event.target.value)} />
            </label>
            <label className="grid gap-1 text-xs text-muted-foreground">
              {t(lang, "admin.abuseTo")}
              <Input type="datetime-local" value={toInput} onChange={(event) => setToInput(event.target.value)} />
            </label>
            <Button onClick={loadData} disabled={loading} className="cursor-pointer">
              {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCw className="h-4 w-4" />}
              {t(lang, "common.refresh")}
            </Button>
          </>
        )}
      />

      {error && (
        <Alert variant="destructive">
          <AlertCircle />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="grid gap-4 md:grid-cols-3">
        <StatCard icon={Upload} label={t(lang, "admin.abuseUploadCount")} value={(overview?.upload_count ?? 0).toLocaleString()} />
        <StatCard icon={HardDrive} label={t(lang, "admin.abuseUploadSize")} value={formatBytes(overview?.upload_size ?? 0)} />
        <StatCard icon={ShieldOff} label={t(lang, "admin.abuseActiveBans")} value={(overview?.active_ip_ban_count ?? 0).toLocaleString()} />
      </div>

      <div className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_minmax(420px,0.9fr)]">
        <Card>
          <CardHeader>
            <CardTitle>{t(lang, "admin.abuseRateLimitTitle")}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm text-muted-foreground">{t(lang, "admin.settingsRateLimitDescription")}</p>
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <Label htmlFor="security-rate-limit-window">{t(lang, "admin.settingsRateLimitWindow")}</Label>
                <Input
                  id="security-rate-limit-window"
                  type="number"
                  min={0}
                  value={runtimeForm.rate_limit_window_minutes}
                  onChange={(event) => updateRuntimeField("rate_limit_window_minutes", Number(event.target.value))}
                  className="h-8"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="security-rate-limit-requests">{t(lang, "admin.settingsRateLimitRequests")}</Label>
                <Input
                  id="security-rate-limit-requests"
                  type="number"
                  min={0}
                  value={runtimeForm.rate_limit_max_requests}
                  onChange={(event) => updateRuntimeField("rate_limit_max_requests", Number(event.target.value))}
                  className="h-8"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="security-upload-rate-limit-window">{t(lang, "admin.settingsUploadRateLimitWindow")}</Label>
                <Input
                  id="security-upload-rate-limit-window"
                  type="number"
                  min={0}
                  value={runtimeForm.upload_rate_limit_window_minutes}
                  onChange={(event) => updateRuntimeField("upload_rate_limit_window_minutes", Number(event.target.value))}
                  className="h-8"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="security-upload-rate-limit-requests">{t(lang, "admin.settingsUploadRateLimitRequests")}</Label>
                <Input
                  id="security-upload-rate-limit-requests"
                  type="number"
                  min={0}
                  value={runtimeForm.upload_rate_limit_max_requests}
                  onChange={(event) => updateRuntimeField("upload_rate_limit_max_requests", Number(event.target.value))}
                  className="h-8"
                />
              </div>
            </div>
            <p className="text-xs text-muted-foreground">{t(lang, "admin.settingsRateLimitHint")}</p>
            <Button size="sm" onClick={handleSaveRateLimits} disabled={runtimeSaving} className="cursor-pointer gap-1">
              {runtimeSaving ? <Loader2 /> : <Check />}
              {t(lang, "common.save")}
            </Button>
          </CardContent>
        </Card>

        <BansCard
          activeBanIds={activeBanIds}
          bans={bans}
          busyKey={busyKey}
          displayIP={displayIP}
          language={lang}
          onDeleteImages={handleDeleteBanImages}
          onToggleIP={toggleIP}
          onUnban={handleUnban}
        />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t(lang, "admin.abuseTopIPs")}</CardTitle>
        </CardHeader>
        <CardContent className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t(lang, "image.ip")}</TableHead>
                <TableHead>{t(lang, "admin.abuseUploadCount")}</TableHead>
                <TableHead>{t(lang, "admin.abuseUploadSize")}</TableHead>
                <TableHead>{t(lang, "admin.abuseLatestUpload")}</TableHead>
                <TableHead>{t(lang, "admin.imagesActions")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {(overview?.top_ips ?? []).map((item) => {
                const activeBan = banForIP(item.ip_address);
                return (
                  <TableRow key={item.ip_address}>
                    <TableCell className="font-mono text-xs">
                      <button type="button" onClick={() => toggleIP(item.ip_address)} className="rounded px-1 py-0.5 hover:bg-muted">
                        {displayIP(item.ip_address, item.ip_address_masked)}
                      </button>
                      {activeBan && <Badge variant="secondary" className="ml-2">{t(lang, "admin.abuseBanned")}</Badge>}
                    </TableCell>
                    <TableCell>{item.upload_count.toLocaleString()}</TableCell>
                    <TableCell>{formatBytes(item.total_size)}</TableCell>
                    <TableCell className="whitespace-nowrap text-xs">{formatDate(item.latest_upload_at)}</TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-2">
                        {activeBan ? (
                          <Button size="sm" variant="outline" onClick={() => handleDeleteBanImages(activeBan)} disabled={busyKey === `delete:${activeBan.id}`} className="h-7 cursor-pointer text-xs">
                            {busyKey === `delete:${activeBan.id}` ? <Loader2 className="h-3 w-3 animate-spin" /> : <Trash2 className="h-3 w-3" />}
                            {t(lang, "admin.abuseDeleteIPImages")}
                          </Button>
                        ) : (
                          <Button size="sm" variant="outline" onClick={() => handleBanIP(item)} disabled={busyKey === `ban:${item.ip_address}`} className="h-7 cursor-pointer text-xs">
                            {busyKey === `ban:${item.ip_address}` ? <Loader2 className="h-3 w-3 animate-spin" /> : <Ban className="h-3 w-3" />}
                            {t(lang, "admin.ipBanAction")}
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{t(lang, "admin.abuseTopTokens")}</CardTitle>
        </CardHeader>
        <CardContent className="overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t(lang, "image.token")}</TableHead>
                <TableHead>{t(lang, "admin.abuseUploadCount")}</TableHead>
                <TableHead>{t(lang, "admin.abuseUploadSize")}</TableHead>
                <TableHead>{t(lang, "admin.abuseLatestUpload")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {(overview?.top_tokens ?? []).map((item) => (
                <TableRow key={item.token}>
                  <TableCell className="font-mono text-xs">{item.token_preview}</TableCell>
                  <TableCell>{item.upload_count.toLocaleString()}</TableCell>
                  <TableCell>{formatBytes(item.total_size)}</TableCell>
                  <TableCell className="whitespace-nowrap text-xs">{formatDate(item.latest_upload_at)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

    </div>
  );
}

function BansCard({
  activeBanIds,
  bans,
  busyKey,
  displayIP,
  language,
  onDeleteImages,
  onToggleIP,
  onUnban,
}: {
  activeBanIds: Set<number>;
  bans: AdminIPBan[];
  busyKey: string | null;
  displayIP: (ip: string, masked: string) => string;
  language: "en" | "zh";
  onDeleteImages: (ban: AdminIPBan) => void;
  onToggleIP: (ip: string) => void;
  onUnban: (ban: AdminIPBan) => void;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{t(language, "admin.abuseBansTitle")}</CardTitle>
      </CardHeader>
      <CardContent className="overflow-x-auto">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t(language, "image.ip")}</TableHead>
              <TableHead>{t(language, "admin.abuseReason")}</TableHead>
              <TableHead>{t(language, "admin.abuseStatus")}</TableHead>
              <TableHead>{t(language, "admin.imagesActions")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {(Array.isArray(bans) ? bans : []).map((ban) => {
              const active = activeBanIds.has(ban.id);
              return (
                <TableRow key={ban.id}>
                  <TableCell className="font-mono text-xs">
                    <button type="button" onClick={() => onToggleIP(ban.ip_address)} className="rounded px-1 py-0.5 hover:bg-muted">
                      {displayIP(ban.ip_address, ban.ip_address_masked)}
                    </button>
                  </TableCell>
                  <TableCell className="max-w-40 truncate text-xs">{ban.reason}</TableCell>
                  <TableCell>
                    <Badge variant={active ? "destructive" : "secondary"}>{active ? t(language, "admin.abuseActive") : t(language, "admin.abuseExpired")}</Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex flex-wrap gap-2">
                      <Button size="sm" variant="outline" onClick={() => onUnban(ban)} disabled={busyKey === `unban:${ban.id}`} className="h-7 cursor-pointer text-xs">
                        {busyKey === `unban:${ban.id}` ? <Loader2 /> : <ShieldOff />}
                        {t(language, "admin.abuseUnban")}
                      </Button>
                      <Button size="sm" variant="outline" onClick={() => onDeleteImages(ban)} disabled={busyKey === `delete:${ban.id}`} className="h-7 cursor-pointer text-xs">
                        {busyKey === `delete:${ban.id}` ? <Loader2 /> : <Trash2 />}
                        {t(language, "admin.abuseDeleteIPImages")}
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}

function StatCard({ icon: Icon, label, value }: { icon: typeof Clock; label: string; value: string }) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
          <Icon className="h-4 w-4 text-violet-500" />
          {label}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-2xl font-bold">{value}</p>
      </CardContent>
    </Card>
  );
}

function normalizeOverview(value: AdminAbuseOverview): AdminAbuseOverview {
  return {
    ...value,
    top_ips: Array.isArray(value.top_ips) ? value.top_ips : [],
    top_tokens: Array.isArray(value.top_tokens) ? value.top_tokens : [],
  };
}

function isActiveBan(ban: AdminIPBan) {
  return !ban.expires_at || new Date(ban.expires_at).getTime() > Date.now();
}

function toDateTimeLocal(date: Date) {
  const offsetMs = date.getTimezoneOffset() * 60 * 1000;
  return new Date(date.getTime() - offsetMs).toISOString().slice(0, 16);
}

function formatDate(value: string) {
  return new Date(value).toLocaleString();
}
