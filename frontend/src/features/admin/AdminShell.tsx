"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { BarChart3, Images, LogOut, Settings } from "lucide-react";
import { useEffect, useState } from "react";

import { Button } from "@/components/ui/Button";
import { Card } from "@/components/ui/Card";
import { useUiTranslations } from "@/hooks/useUiPreferences";
import { adminStatus } from "@/lib/api";
import { cn } from "@/lib/utils";
import { useAdminSessionStore } from "@/stores/admin-session-store";
import type { AdminStatus } from "@/types/admin";

import { AdminStatusProvider } from "./admin-status-context";

export function AdminShell({ children }: { children: React.ReactNode }) {
  const token = useAdminSessionStore((state) => state.token);
  const hasHydrated = useAdminSessionStore((state) => state.hasHydrated);
  const clearToken = useAdminSessionStore((state) => state.clearToken);
  const pathname = usePathname();
  const router = useRouter();
  const [checking, setChecking] = useState(true);
  const [verifiedStatus, setVerifiedStatus] = useState<AdminStatus | null>(null);
  const t = useUiTranslations();

  const navItems = [
    { href: "/admin/dashboard", label: t.admin.nav.status, icon: <BarChart3 aria-hidden="true" className="h-4 w-4" /> },
    { href: "/admin/dashboard/images", label: t.admin.nav.images, icon: <Images aria-hidden="true" className="h-4 w-4" /> },
    { href: "/admin/dashboard/settings", label: t.admin.nav.settings, icon: <Settings aria-hidden="true" className="h-4 w-4" /> }
  ];

  useEffect(() => {
    if (!hasHydrated) {
      return;
    }

    let cancelled = false;

    async function verify() {
      if (!token) {
        setVerifiedStatus(null);
        setChecking(false);
        router.replace("/admin/login");
        return;
      }

      setChecking(true);
      try {
        const status = await adminStatus(token);
        if (!cancelled) {
          setVerifiedStatus(status);
          setChecking(false);
        }
      } catch {
        if (cancelled) {
          return;
        }
        clearToken();
        setVerifiedStatus(null);
        setChecking(false);
        router.replace("/admin/login");
      }
    }

    void verify();
    return () => {
      cancelled = true;
    };
  }, [clearToken, hasHydrated, router, token]);

  if (!hasHydrated || checking) {
    return (
      <Card className="flex items-center gap-4 p-6 text-sm text-muted-foreground" role="status" variant="strong">
        <span className="skeleton-glass h-10 w-10 rounded-md" />
        <span>{t.admin.checkingSession}</span>
      </Card>
    );
  }

  return (
    <AdminStatusProvider verifiedStatus={verifiedStatus}>
      <div className="grid gap-6 xl:grid-cols-[260px_1fr]">
        <aside className="xl:sticky xl:top-24 xl:h-fit">
          <Card className="overflow-hidden p-3" variant="strong">
            <div className="space-y-1 px-2 py-3">
              <p className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                {t.admin.shellEyebrow}
              </p>
              <h1 className="text-xl font-semibold text-foreground">{t.admin.shellTitle}</h1>
            </div>
            <nav className="mt-2 grid gap-1" aria-label={t.admin.shellTitle}>
              {navItems.map((item) => (
                <Link
                  className={cn(
                    "flex items-center gap-3 rounded-md px-3 py-2.5 text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
                    pathname === item.href
                      ? "bg-muted text-foreground"
                      : "text-muted-foreground hover:bg-muted/70 hover:text-foreground"
                  )}
                  href={item.href}
                  key={item.href}
                >
                  {item.icon}
                  {item.label}
                </Link>
              ))}
            </nav>
            <div className="mt-4 border-t border-border pt-3">
              <Button
                className="w-full justify-start"
                onClick={() => {
                  clearToken();
                  router.push("/admin/login");
                }}
                variant="ghost"
              >
                <LogOut aria-hidden="true" className="h-4 w-4" />
                {t.admin.signOut}
              </Button>
            </div>
          </Card>
        </aside>
        <section className="min-w-0">{children}</section>
      </div>
    </AdminStatusProvider>
  );
}
