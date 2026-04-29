"use client";

import { PageIntro } from "@/components/shared/PageLayout";
import { useUiTranslations } from "@/hooks/useUiPreferences";

import { StatusPanel } from "./StatusPanel";

export function DashboardOverview() {
  const t = useUiTranslations();

  return (
    <div className="space-y-5 animate-fade-in">
      <PageIntro eyebrow={t.admin.dashboardEyebrow} title={t.admin.dashboardTitle} />
      <StatusPanel />
    </div>
  );
}
