"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/Card";
import { useAdminSessionStore } from "@/stores/admin-session-store";
import { useUiPreferencesStore } from "@/stores/ui-preferences-store";
import { useAdminStatus } from "./admin-status-context";
import { adminGetStatus } from "@/lib/api";
import { t } from "@/lib/i18n";
import { formatBytes } from "@/lib/utils";
import { Image, HardDrive, TrendingUp, Users, Loader2, AlertCircle } from "lucide-react";
import type { AdminStatus } from "@/types";

export function DashboardOverview() {
  const token = useAdminSessionStore((state) => state.token);
  const language = useUiPreferencesStore((state) => state.language);
  const verifiedStatus = useAdminStatus();

  const [status, setStatus] = useState<AdminStatus | null>(verifiedStatus);
  const [loading, setLoading] = useState(!verifiedStatus);
  const [error, setError] = useState("");

  useEffect(() => {
    if (verifiedStatus) {
      setStatus(verifiedStatus);
      setLoading(false);
      return;
    }
    if (!token) return;
    setLoading(true);
    setError("");
    adminGetStatus(token)
      .then((s) => { setStatus(s); setLoading(false); })
      .catch((err) => { setError(err instanceof Error ? err.message : "Failed"); setLoading(false); });
  }, [token, verifiedStatus]);

  const lang = language;

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center gap-2 text-destructive" role="alert">
        <AlertCircle className="h-5 w-5" />
        {error}
      </div>
    );
  }

  const stats = status
    ? [
        { label: t(lang, "admin.totalImages"), value: status.total_images.toLocaleString(), icon: Image, color: "text-violet-500" },
        { label: t(lang, "admin.totalSize"), value: formatBytes(status.total_storage_size), icon: HardDrive, color: "text-cyan-500" },
        { label: t(lang, "admin.todayUploads"), value: status.today_uploads.toLocaleString(), icon: TrendingUp, color: "text-green-500" },
        { label: t(lang, "admin.uniqueTokens"), value: status.unique_tokens.toLocaleString(), icon: Users, color: "text-amber-500" },
      ]
    : [];

  return (
    <div className="space-y-6">
      <h1 className="text-xl font-bold">{t(lang, "admin.statusTitle")}</h1>
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((s) => (
          <Card key={s.label}>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground flex items-center gap-2">
                <s.icon className={`h-4 w-4 ${s.color}`} />
                {s.label}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-2xl font-bold">{s.value}</p>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
