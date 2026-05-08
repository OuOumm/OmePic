import { Activity, Database, Globe2, HardDrive, KeyRound, Server, ShieldCheck, Wrench } from "lucide-react";

import { Badge } from "@/components/ui/Badge";
import { Card, CardContent } from "@/components/ui/Card";
import { t } from "@/lib/i18n";
import { cn } from "@/lib/utils";
import type { AdminSystemSettings, Language, SecretStatus } from "@/types";

type StatusTone = "good" | "warn" | "muted";

type StatusItem = {
  label: string;
  value: string;
  tone?: StatusTone;
  iconKind?: "secret" | "storage" | "database" | "network" | "activity";
};

type StatusSection = {
  title: string;
  description: string;
  icon: typeof Server;
  items: StatusItem[];
};

type Props = {
  language: Language;
  systemSettings: AdminSystemSettings | null;
};

export function SystemStatusPanel({ language, systemSettings }: Props) {
  const readonly = systemSettings?.readonly;
  const booleanLabel = (value: boolean | undefined) => value ? t(language, "common.enabled") : t(language, "common.disabled");
  const configuredLabel = (value: boolean | undefined) => value ? t(language, "admin.systemConfigured") : t(language, "admin.systemNotConfigured");
  const sections: StatusSection[] = [
    {
      title: t(language, "admin.systemRuntimeTitle"),
      description: t(language, "admin.systemRuntimeDescription"),
      icon: Server,
      items: [
        { label: t(language, "admin.systemHttpAddress"), value: readonly?.environment?.http_addr || "-", tone: "muted", iconKind: "network" },
        { label: t(language, "admin.systemDatabasePath"), value: readonly?.environment?.database_path || "-", tone: "muted", iconKind: "database" },
        { label: "Redis", value: configuredLabel(readonly?.environment?.redis_configured), tone: readonly?.environment?.redis_configured ? "good" : "muted", iconKind: "activity" },
        { label: t(language, "admin.systemPublicUrlSource"), value: readonly?.environment?.public_base_url_source || "-", tone: "muted", iconKind: "network" },
      ],
    },
    {
      title: t(language, "admin.systemSecurityTitle"),
      description: t(language, "admin.systemSecurityDescription"),
      icon: ShieldCheck,
      items: [
        secretItem(language, "JWT Secret", readonly?.security?.jwt_secret),
        secretItem(language, "Admin Password", readonly?.security?.admin_password),
        secretItem(language, "UID Encryption Key", readonly?.security?.uid_encryption_key),
      ],
    },
    {
      title: t(language, "admin.systemStorageTitle"),
      description: t(language, "admin.systemStorageDescription"),
      icon: HardDrive,
      items: [
        { label: t(language, "admin.systemDefaultStorage"), value: readonly?.storage?.default_storage_key || "-", tone: "muted", iconKind: "storage" },
        { label: t(language, "admin.systemStorageInstanceCount"), value: String(readonly?.storage?.storage_config_count ?? "-"), tone: "muted", iconKind: "storage" },
        { label: t(language, "admin.systemStorageSelection"), value: booleanLabel(readonly?.storage?.allow_storage_selection), tone: readonly?.storage?.allow_storage_selection ? "good" : "muted", iconKind: "storage" },
      ],
    },
    {
      title: t(language, "admin.systemServiceTitle"),
      description: t(language, "admin.systemServiceDescription"),
      icon: Activity,
      items: [
        { label: t(language, "admin.systemHealthStatus"), value: readonly?.service?.health || "-", tone: readonly?.service?.health === "ok" ? "good" : "warn", iconKind: "activity" },
        { label: t(language, "admin.systemMaintenanceMode"), value: booleanLabel(readonly?.service?.maintenance_mode), tone: readonly?.service?.maintenance_mode ? "warn" : "good", iconKind: "activity" },
      ],
    },
  ];

  const health = readonly?.service?.health === "ok";
  const maintenance = readonly?.service?.maintenance_mode;

  return (
    <section className="space-y-4">
      <div className="flex flex-col gap-3 rounded-2xl border bg-gradient-to-br from-background via-muted/20 to-muted/50 p-5 shadow-sm md:flex-row md:items-center md:justify-between">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <div className="rounded-xl border bg-background p-2 shadow-sm">
              <Wrench className="h-4 w-4 text-primary" />
            </div>
            <h2 className="text-lg font-semibold tracking-tight">{t(language, "admin.statusTitle")}</h2>
          </div>
          <p className="text-sm text-muted-foreground">{t(language, "admin.systemPanelDescription")}</p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant="outline" className={health ? toneClassName("good") : toneClassName("warn")}>
            {health ? t(language, "admin.systemServiceNormal") : t(language, "admin.systemServiceAbnormal")}
          </Badge>
          <Badge variant="outline" className={maintenance ? toneClassName("warn") : toneClassName("good")}>
            {maintenance ? t(language, "admin.systemServiceMaintenance") : t(language, "admin.systemServicePublic")}
          </Badge>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-2">
        {sections.map((section) => (
          <SystemStatusCard key={section.title} section={section} />
        ))}
      </div>
    </section>
  );
}

function SystemStatusCard({ section }: { section: StatusSection }) {
  return (
    <Card className="overflow-hidden">
      <CardContent className="p-0">
        <div className="flex items-start gap-3 border-b bg-muted/30 px-5 py-4">
          <div className="rounded-xl border bg-background p-2 shadow-sm">
            <section.icon className="h-4 w-4 text-primary" />
          </div>
          <div>
            <h3 className="font-semibold">{section.title}</h3>
            <p className="mt-0.5 text-xs text-muted-foreground">{section.description}</p>
          </div>
        </div>
        <div className="divide-y">
          {section.items.map((item) => (
            <div key={item.label} className="flex items-center justify-between gap-4 px-5 py-3 text-sm">
              <span className="flex items-center gap-2 text-muted-foreground">
                <ItemIcon iconKind={item.iconKind} />
                {item.label}
              </span>
              <span className={cn("max-w-[55%] break-all text-right font-medium", item.tone && toneTextClassName(item.tone))}>
                {item.value}
              </span>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

function ItemIcon({ iconKind }: { iconKind?: StatusItem["iconKind"] }) {
  switch (iconKind) {
    case "secret":
      return <KeyRound className="h-3.5 w-3.5" />;
    case "storage":
    case "database":
      return <Database className="h-3.5 w-3.5" />;
    case "network":
      return <Globe2 className="h-3.5 w-3.5" />;
    default:
      return <Activity className="h-3.5 w-3.5" />;
  }
}

function secretItem(language: Language, label: string, status?: SecretStatus): StatusItem {
  if (!status) return { label, value: "-", tone: "muted", iconKind: "secret" };
  if (!status.configured) return { label, value: t(language, "admin.systemNotConfigured"), tone: "warn", iconKind: "secret" };
  return {
    label,
    value: status.using_default ? t(language, "admin.systemConfiguredDefault") : t(language, "admin.systemConfigured"),
    tone: status.using_default ? "warn" : "good",
    iconKind: "secret",
  };
}

function toneClassName(tone: StatusTone): string {
  switch (tone) {
    case "good":
      return "border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300";
    case "warn":
      return "border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-300";
    default:
      return "border-border bg-muted text-muted-foreground";
  }
}

function toneTextClassName(tone: StatusTone): string {
  switch (tone) {
    case "good":
      return "text-emerald-700 dark:text-emerald-300";
    case "warn":
      return "text-amber-700 dark:text-amber-300";
    default:
      return "text-foreground";
  }
}
