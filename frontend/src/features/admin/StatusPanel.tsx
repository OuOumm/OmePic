"use client";

import { useEffect, useRef, useState } from "react";

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
        <StatCard icon={<GalleryIcon />} label={t.admin.stats.totalImages} value={String(displayStatus.total_images)} />
        <StatCard icon={<StorageIcon />} label={t.admin.stats.storageSize} value={formatBytes(displayStatus.total_storage_size)} />
        <StatCard icon={<TodayIcon />} label={t.admin.stats.todaysUploads} value={String(displayStatus.today_uploads)} />
        <StatCard icon={<TokenIcon />} label={t.admin.stats.uniqueTokens} value={String(displayStatus.unique_tokens)} />
      </div>
    </Card>
  );
}

function StatCard({ icon, label, value }: { icon: React.ReactNode; label: string; value: string }) {
  return (
    <Card className="group relative overflow-hidden p-5 transition-all duration-300 hover:-translate-y-1 hover:border-violet-300/50 hover:shadow-glow dark:hover:border-violet-400/30" variant="subtle">
      <div className="absolute inset-0 bg-gradient-to-br from-violet-500/10 via-transparent to-cyan-500/10 opacity-0 transition-opacity duration-300 group-hover:opacity-100" />
      <div className="relative">
        <div className="flex h-11 w-11 items-center justify-center rounded-2xl bg-gradient-to-br from-violet-500/20 to-cyan-500/20 text-violet-600 dark:text-violet-200">
          {icon}
        </div>
        <p className="mt-4 text-sm font-semibold text-muted">{label}</p>
        <p className="mt-2 text-3xl font-bold tabular-nums tracking-tight text-slate-900 dark:text-white">
          {value}
        </p>
      </div>
    </Card>
  );
}

function GalleryIcon() {
  return (
    <svg aria-hidden="true" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="M4 5a2 2 0 0 1 2-2h12a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V5Zm3 12 3.5-4 2.5 3 2-2.4 2 3.4M15 8h.01" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function StorageIcon() {
  return (
    <svg aria-hidden="true" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="M4 7c0 1.7 3.6 3 8 3s8-1.3 8-3-3.6-3-8-3-8 1.3-8 3Zm0 0v10c0 1.7 3.6 3 8 3s8-1.3 8-3V7M4 12c0 1.7 3.6 3 8 3s8-1.3 8-3" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function TodayIcon() {
  return (
    <svg aria-hidden="true" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="M7 3v3m10-3v3M4 9h16M6 5h12a2 2 0 0 1 2 2v11a3 3 0 0 1-3 3H7a3 3 0 0 1-3-3V7a2 2 0 0 1 2-2Z" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function TokenIcon() {
  return (
    <svg aria-hidden="true" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="M15.75 7.5a5.25 5.25 0 1 0-4 5.1L5 19.35V22h2.65l1.2-1.2V19h1.8l1.4-1.4v-1.8l2.85-2.85a5.22 5.22 0 0 0 .85-5.45ZM16.5 6.75h.01" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}
