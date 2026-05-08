"use client";

import { useEffect, useMemo, useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import { Button } from "@/components/ui/Button";
import { Separator } from "@/components/ui/Separator";
import { LoginForm } from "./LoginForm";
import { AdminStatusProvider } from "./admin-status-context";
import { useAdminSessionStore } from "@/stores/admin-session-store";
import { useUiPreferencesStore } from "@/stores/ui-preferences-store";
import { adminGetStatus } from "@/lib/api";
import { t } from "@/lib/i18n";
import { cn, formatBytes } from "@/lib/utils";
import {
  Activity,
  ArrowLeft,
  Image,
  LayoutDashboard,
  Loader2,
  LogOut,
  Menu,
  Settings,
  ShieldAlert,
} from "lucide-react";
import type { AdminStatus } from "@/types";

export function AdminShell({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();
  const token = useAdminSessionStore((state) => state.token);
  const clearToken = useAdminSessionStore((state) => state.clearToken);
  const hasHydrated = useAdminSessionStore((state) => state.hasHydrated);
  const language = useUiPreferencesStore((state) => state.language);
  const [validating, setValidating] = useState(!!token);
  const [verifiedStatus, setVerifiedStatus] = useState<AdminStatus | null>(null);
  const [mobileNavOpen, setMobileNavOpen] = useState(false);

  useEffect(() => {
    if (hasHydrated && token) {
      setValidating(true);
      adminGetStatus(token)
        .then((status) => {
          setVerifiedStatus(status);
          setValidating(false);
        })
        .catch(() => {
          clearToken();
          setValidating(false);
        });
    }
  }, [hasHydrated, token, clearToken]);

  const lang = language;
  const sidebarItems = useMemo(
    () => [
      {
        href: "/admin/dashboard",
        label: t(lang, "admin.sidebarStatus"),
        icon: LayoutDashboard,
      },
      {
        href: "/admin/dashboard/images",
        label: t(lang, "admin.sidebarImages"),
        icon: Image,
      },
      {
        href: "/admin/dashboard/security",
        label: t(lang, "admin.sidebarSecurity"),
        icon: ShieldAlert,
      },
      {
        href: "/admin/dashboard/settings",
        label: t(lang, "admin.sidebarSettings"),
        icon: Settings,
      },
    ],
    [lang]
  );

  const currentItem = sidebarItems.find((item) => pathname === item.href) ??
    sidebarItems.find((item) => item.href !== "/admin/dashboard" && pathname.startsWith(item.href)) ??
    sidebarItems[0];

  const handleLogout = () => {
    clearToken();
    router.push("/admin/dashboard");
  };

  const handleNavigate = (href: string) => {
    router.push(href);
    setMobileNavOpen(false);
  };

  if (!hasHydrated) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <Loader2 className="size-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (!token) {
    return (
      <div className="min-h-screen bg-background px-4 py-10">
        <LoginForm />
      </div>
    );
  }

  if (validating) {
    return (
      <div className="flex min-h-screen items-center justify-center gap-2 bg-background">
        <Loader2 className="size-5 animate-spin text-muted-foreground" />
        <span className="text-sm text-muted-foreground">{t(lang, "common.loading")}</span>
      </div>
    );
  }

  return (
    <AdminStatusProvider verifiedStatus={verifiedStatus}>
      <div className="min-h-screen bg-muted/20" id="main-content">
        <div className="flex min-h-screen">
          <aside className="hidden w-64 shrink-0 border-r bg-background/95 lg:block">
            <AdminSidebar
              currentPath={pathname}
              items={sidebarItems}
              language={lang}
              status={verifiedStatus}
              onLogout={handleLogout}
              onNavigate={handleNavigate}
            />
          </aside>

          {mobileNavOpen && (
            <div className="fixed inset-0 z-50 lg:hidden">
              <button
                type="button"
                aria-label={t(lang, "admin.closeNavigation")}
                className="absolute inset-0 bg-background/80 backdrop-blur-sm"
                onClick={() => setMobileNavOpen(false)}
              />
              <aside className="relative h-full w-72 border-r bg-background shadow-lg">
                <AdminSidebar
                  currentPath={pathname}
                  items={sidebarItems}
                  language={lang}
                  status={verifiedStatus}
                  onLogout={handleLogout}
                  onNavigate={handleNavigate}
                />
              </aside>
            </div>
          )}

          <div className="flex min-w-0 flex-1 flex-col">
            <header className="sticky top-0 z-30 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/80">
              <div className="flex h-14 items-center justify-between gap-3 px-4 lg:px-6">
                <div className="flex min-w-0 items-center gap-3">
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => setMobileNavOpen(true)}
                    className="cursor-pointer lg:hidden"
                    aria-label={t(lang, "admin.openNavigation")}
                  >
                    <Menu />
                  </Button>
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium">{currentItem.label}</p>
                    <p className="hidden text-xs text-muted-foreground sm:block">{t(lang, "admin.consoleSubtitle")}</p>
                  </div>
                </div>
                <Button variant="outline" size="sm" onClick={() => router.push("/")} className="cursor-pointer">
                  <ArrowLeft />
                  <span className="hidden sm:inline">{t(lang, "admin.backToSite")}</span>
                </Button>
              </div>
            </header>

            <main className="min-w-0 flex-1 px-4 py-5 lg:px-6 lg:py-6">
              <div className="mx-auto flex w-full max-w-7xl flex-col gap-6">{children}</div>
            </main>
          </div>
        </div>
      </div>
    </AdminStatusProvider>
  );
}

type AdminSidebarProps = {
  currentPath: string;
  items: Array<{ href: string; label: string; icon: typeof LayoutDashboard }>;
  language: "en" | "zh";
  status: AdminStatus | null;
  onLogout: () => void;
  onNavigate: (href: string) => void;
};

function AdminSidebar({ currentPath, items, language, status, onLogout, onNavigate }: AdminSidebarProps) {
  return (
    <div className="flex h-full min-h-screen flex-col gap-4 p-4">
      <div className="rounded-xl border bg-card p-4">
        <div className="flex items-center gap-3">
          <div className="flex size-10 items-center justify-center rounded-lg bg-primary text-primary-foreground">
            <ShieldAlert className="size-5" />
          </div>
          <div className="min-w-0">
            <p className="truncate font-semibold">OmePic Admin</p>
            <p className="truncate text-xs text-muted-foreground">{t(language, "admin.consoleSubtitle")}</p>
          </div>
        </div>
      </div>

      <nav className="flex flex-col gap-1" aria-label={t(language, "admin.navigation")}> 
        {items.map((item) => {
          const isActive = currentPath === item.href || (item.href !== "/admin/dashboard" && currentPath.startsWith(item.href));
          return (
            <Button
              key={item.href}
              variant={isActive ? "secondary" : "ghost"}
              size="sm"
              onClick={() => onNavigate(item.href)}
              className={cn("h-9 justify-start cursor-pointer", isActive && "font-semibold")}
            >
              <item.icon />
              {item.label}
            </Button>
          );
        })}
      </nav>

      <div className="mt-auto flex flex-col gap-4">
        <div className="rounded-xl border bg-card p-3">
          <div className="mb-3 flex items-center gap-2 text-xs font-medium text-muted-foreground">
            <Activity className="size-3.5" />
            {t(language, "admin.systemSnapshot")}
          </div>
          <div className="grid gap-2 text-xs">
            <SnapshotRow label={t(language, "admin.totalImages")} value={(status?.total_images ?? 0).toLocaleString()} />
            <SnapshotRow label={t(language, "admin.totalSize")} value={formatBytes(status?.total_storage_size ?? 0)} />
            <SnapshotRow label={t(language, "admin.todayUploads")} value={(status?.today_uploads ?? 0).toLocaleString()} />
          </div>
        </div>
        <Separator />
        <Button
          variant="ghost"
          size="sm"
          onClick={onLogout}
          className="w-full justify-start cursor-pointer text-muted-foreground hover:text-destructive"
        >
          <LogOut />
          {t(language, "admin.logout")}
        </Button>
      </div>
    </div>
  );
}

function SnapshotRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-3">
      <span className="truncate text-muted-foreground">{label}</span>
      <span className="shrink-0 font-mono font-medium">{value}</span>
    </div>
  );
}
