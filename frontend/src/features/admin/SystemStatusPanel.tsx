import { Activity, Database, Globe2, HardDrive, KeyRound, Server, ShieldCheck, Wrench } from "lucide-react";

import { Badge } from "@/components/ui/Badge";
import { Card, CardContent } from "@/components/ui/Card";
import { cn } from "@/lib/utils";
import type { AdminSystemSettings, SecretStatus } from "@/types";

type StatusItem = {
  label: string;
  value: string;
  tone?: "good" | "warn" | "muted";
};

type StatusSection = {
  title: string;
  description: string;
  icon: typeof Server;
  items: StatusItem[];
};

type Props = {
  systemSettings: AdminSystemSettings | null;
};

export function SystemStatusPanel({ systemSettings }: Props) {
  const readonly = systemSettings?.readonly;
  const sections: StatusSection[] = [
    {
      title: "运行环境",
      description: "服务入口、数据路径与公开链接来源",
      icon: Server,
      items: [
        { label: "HTTP 地址", value: readonly?.environment?.http_addr || "-", tone: "muted" },
        { label: "数据库路径", value: readonly?.environment?.database_path || "-", tone: "muted" },
        { label: "Redis", value: readonly?.environment?.redis_configured ? "已配置" : "未配置", tone: readonly?.environment?.redis_configured ? "good" : "muted" },
        { label: "公开 URL 来源", value: readonly?.environment?.public_base_url_source || "-", tone: "muted" },
      ],
    },
    {
      title: "安全状态",
      description: "关键密钥只展示配置状态，不显示明文",
      icon: ShieldCheck,
      items: [
        secretItem("JWT Secret", readonly?.security?.jwt_secret),
        secretItem("Admin Password", readonly?.security?.admin_password),
        secretItem("UID Encryption Key", readonly?.security?.uid_encryption_key),
      ],
    },
    {
      title: "存储状态",
      description: "默认存储与上传侧选择策略",
      icon: HardDrive,
      items: [
        { label: "默认存储", value: readonly?.storage?.default_storage_key || "-", tone: "muted" },
        { label: "存储实例数量", value: String(readonly?.storage?.storage_config_count ?? "-"), tone: "muted" },
        { label: "多存储选择", value: readonly?.storage?.allow_storage_selection ? "开启" : "关闭", tone: readonly?.storage?.allow_storage_selection ? "good" : "muted" },
      ],
    },
    {
      title: "服务状态",
      description: "健康状态与维护模式开关",
      icon: Activity,
      items: [
        { label: "健康状态", value: readonly?.service?.health || "-", tone: readonly?.service?.health === "ok" ? "good" : "warn" },
        { label: "维护模式", value: readonly?.service?.maintenance_mode ? "开启" : "关闭", tone: readonly?.service?.maintenance_mode ? "warn" : "good" },
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
            <h2 className="text-lg font-semibold tracking-tight">系统状态</h2>
          </div>
          <p className="text-sm text-muted-foreground">集中查看运行环境、安全配置、存储和服务健康情况。</p>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant="outline" className={health ? toneClassName("good") : toneClassName("warn")}>
            {health ? "服务正常" : "服务异常"}
          </Badge>
          <Badge variant="outline" className={maintenance ? toneClassName("warn") : toneClassName("good")}>
            {maintenance ? "维护中" : "对外开放"}
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
                <ItemIcon label={item.label} />
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

function ItemIcon({ label }: { label: string }) {
  if (label.includes("Secret") || label.includes("Password") || label.includes("Key")) {
    return <KeyRound className="h-3.5 w-3.5" />;
  }
  if (label.includes("存储") || label.includes("数据库")) {
    return <Database className="h-3.5 w-3.5" />;
  }
  if (label.includes("URL") || label.includes("HTTP")) {
    return <Globe2 className="h-3.5 w-3.5" />;
  }
  return <Activity className="h-3.5 w-3.5" />;
}

function secretItem(label: string, status?: SecretStatus): StatusItem {
  if (!status) return { label, value: "-", tone: "muted" };
  if (!status.configured) return { label, value: "未配置", tone: "warn" };
  return { label, value: status.using_default ? "已配置，使用默认值" : "已配置", tone: status.using_default ? "warn" : "good" };
}

function toneClassName(tone: "good" | "warn" | "muted"): string {
  switch (tone) {
    case "good":
      return "border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300";
    case "warn":
      return "border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-300";
    default:
      return "border-border bg-muted text-muted-foreground";
  }
}

function toneTextClassName(tone: "good" | "warn" | "muted"): string {
  switch (tone) {
    case "good":
      return "text-emerald-700 dark:text-emerald-300";
    case "warn":
      return "text-amber-700 dark:text-amber-300";
    default:
      return "text-foreground";
  }
}
