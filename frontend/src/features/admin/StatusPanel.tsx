"use client";

import { useEffect, useRef, useState } from "react";
import { CalendarDays, Database, Images, KeyRound } from "lucide-react";

import { PageSectionHeader } from "@/components/shared/PageLayout";
import { Card } from "@/components/ui/Card";
import { useUiTranslations } from "@/hooks/useUiPreferences";
import { adminStatus } from "@/lib/api";
import { formatBytes } from "@/lib/format";
import { useAdminSessionStore } from "@/stores/admin-session-store";
import type { AdminStatus } from "@/types/admin";

import { useVerifiedAdminStatus } from "./admin-status-context";

export function StatusPanel() {
  const token = useAdminSessionStore((state) => state.token);
  const t = useUiTranslations();
  const verifiedStatus = useVerifiedAdminStatus();
  const [status, setStatus] = useState<AdminStatus | null>(null);
  const [error, setError] = useState("");
  const statusLoadFailedRef = useRef(t.admin.statusLoadFailed);

  useEffect(() => {
    statusLoadFailedRef.current = t.admin.statusLoadFailed;
  }, [t.admin.statusLoadFailed]);

  useEffect(() => {
    if (verifiedStatus) {
      return;
    }

    let cancelled = false;

    async function load() {
      try {
        const result = await adminStatus(token);
        if (!cancelled) {
          setStatus(result);
          setError("");
        }
      } catch (loadError) {
        if (!cancelled) {
          setError(loadError instanceof Error ? loadError.message : statusLoadFailedRef.current);
        }
      }
    }
    void load();

    return () => {
      cancelled = true;
    };
  }, [token, verifiedStatus]);

  const displayStatus = verifiedStatus ?? status;

  if (!verifiedStatus && error) {
    return (
      <Card className="border-rose-400/30 bg-rose-500/10 p-5 text-sm text-danger" role="alert" variant="subtle">
        {error}
      </Card>
    );
  }

  if (!displayStatus) {
    return (
      <Card className="space-y-5 p-5 sm:p-6" role="status" variant="strong">
        <PageSectionHeader title={t.admin.dashboardTitle} />
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {[0, 1, 2, 3].map((item) => (
            <Card className="p-5" key={item} variant="subtle">
              <div className="skeleton-glass h-4 w-24" />
              <div className="skeleton-glass mt-5 h-9 w-28" />
            </Card>
          ))}
        </div>
        <span className="sr-only">{t.admin.statusLoading}</span>
      </Card>
    );
  }

  return (
    <Card className="space-y-5 p-5 sm:p-6" variant="strong">
      <PageSectionHeader title={t.admin.dashboardTitle} />
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <StatCard icon={<Images aria-hidden="true" className="h-5 w-5" />} label={t.admin.stats.totalImages} value={String(displayStatus.total_images)} />
        <StatCard icon={<Database aria-hidden="true" className="h-5 w-5" />} label={t.admin.stats.storageSize} value={formatBytes(displayStatus.total_storage_size)} />
        <StatCard icon={<CalendarDays aria-hidden="true" className="h-5 w-5" />} label={t.admin.stats.todaysUploads} value={String(displayStatus.today_uploads)} />
        <StatCard icon={<KeyRound aria-hidden="true" className="h-5 w-5" />} label={t.admin.stats.uniqueTokens} value={String(displayStatus.unique_tokens)} />
      </div>
    </Card>
  );
}

function StatCard({ icon, label, value }: { icon: React.ReactNode; label: string; value: string }) {
  return (
    <Card className="p-5" variant="subtle">
      <div>
        <div className="flex h-10 w-10 items-center justify-center rounded-md border border-border bg-card text-muted-foreground">
          {icon}
        </div>
        <p className="mt-4 text-sm font-medium text-muted-foreground">{label}</p>
        <p className="mt-2 text-3xl font-semibold tabular-nums tracking-tight text-foreground">
          {value}
        </p>
      </div>
    </Card>
  );
}
