"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { usePathname, useRouter } from "next/navigation";

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
    { href: "/admin/dashboard", label: t.admin.nav.status, icon: <PulseIcon /> },
    { href: "/admin/dashboard/images", label: t.admin.nav.images, icon: <GalleryIcon /> },
    { href: "/admin/dashboard/settings", label: t.admin.nav.settings, icon: <SettingsIcon /> }
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
      <Card className="flex items-center gap-4 p-6 text-sm text-muted" role="status" variant="strong">
        <span className="skeleton-glass h-11 w-11 rounded-xl" />
        <span>{t.admin.checkingSession}</span>
      </Card>
    );
  }

  return (
    <AdminStatusProvider verifiedStatus={verifiedStatus}>
      <div className="grid gap-5 xl:grid-cols-[280px_1fr]">
        <aside className="lg:sticky lg:top-24 lg:h-fit">
          <Card className="overflow-hidden p-4" variant="strong">
            <div className="rounded-[24px] border border-white/45 bg-white/45 p-4 backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/35">
              <div className="space-y-1">
                <p className="text-xs font-bold uppercase tracking-[0.22em] text-violet-600 dark:text-violet-300">
              {t.admin.shellEyebrow}
                </p>
                <h1 className="text-xl font-bold text-slate-900 dark:text-white">{t.admin.shellTitle}</h1>
              </div>
              <nav className="mt-5 grid gap-1.5" aria-label={t.admin.shellTitle}>
                {navItems.map((item) => (
                  <Link
                    className={cn(
                      "flex items-center gap-3 rounded-2xl px-3.5 py-3 text-sm font-semibold transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-violet-400 focus-visible:ring-offset-2 focus-visible:ring-offset-surface",
                      pathname === item.href
                        ? "bg-gradient-to-r from-violet-600 to-cyan-600 text-white shadow-lg shadow-violet-500/20"
                        : "text-muted hover:bg-white/70 hover:text-violet-700 dark:hover:bg-white/10 dark:hover:text-violet-200"
                    )}
                    href={item.href}
                    key={item.href}
                  >
                    {item.icon}
                    {item.label}
                  </Link>
                ))}
              </nav>
            </div>
            <Button
              className="mt-4 w-full"
              onClick={() => {
                clearToken();
                router.push("/admin/login");
              }}
              variant="secondary"
            >
              <SignOutIcon />
              {t.admin.signOut}
            </Button>
          </Card>
        </aside>
        <section className="min-w-0">{children}</section>
      </div>
    </AdminStatusProvider>
  );
}

function PulseIcon() {
  return (
    <svg aria-hidden="true" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="M4 12h4l2-5 4 10 2-5h4" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function GalleryIcon() {
  return (
    <svg aria-hidden="true" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="M4 5a2 2 0 0 1 2-2h12a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V5Zm3 12 3.5-4 2.5 3 2-2.4 2 3.4M15 8h.01" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function SettingsIcon() {
  return (
    <svg aria-hidden="true" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="M12 8a4 4 0 1 0 0 8 4 4 0 0 0 0-8Zm8.5 4a8.5 8.5 0 0 0-.2-1.8l2-1.5-2-3.4-2.4 1a8.5 8.5 0 0 0-3-1.7L14.5 2h-5l-.4 2.6a8.5 8.5 0 0 0-3 1.7l-2.4-1-2 3.4 2 1.5a8.5 8.5 0 0 0 0 3.6l-2 1.5 2 3.4 2.4-1a8.5 8.5 0 0 0 3 1.7l.4 2.6h5l.4-2.6a8.5 8.5 0 0 0 3-1.7l2.4 1 2-3.4-2-1.5c.1-.6.2-1.2.2-1.8Z" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function SignOutIcon() {
  return (
    <svg aria-hidden="true" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="M15 8V5a2 2 0 0 0-2-2H5v18h8a2 2 0 0 0 2-2v-3m-4-4h10m0 0-3-3m3 3-3 3" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}
