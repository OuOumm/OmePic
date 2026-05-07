"use client";

import { useState, useEffect } from "react";
import { useRouter, usePathname } from "next/navigation";
import { Button } from "@/components/ui/Button";
import { Separator } from "@/components/ui/Separator";
import { LoginForm } from "./LoginForm";
import { AdminStatusProvider } from "./admin-status-context";
import { useAdminSessionStore } from "@/stores/admin-session-store";
import { useUiPreferencesStore } from "@/stores/ui-preferences-store";
import { adminGetStatus } from "@/lib/api";
import { t } from "@/lib/i18n";
import {
  LayoutDashboard,
  Image,
  Settings,
  LogOut,
  Loader2,
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

  const handleLogout = () => {
    clearToken();
    router.push("/admin/dashboard");
  };

  const lang = language;

  // Not hydrated yet
  if (!hasHydrated) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  // No token, show login
  if (!token) {
    return <LoginForm />;
  }

  // Validating token
  if (validating) {
    return (
      <div className="flex items-center justify-center py-20 gap-2">
        <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
        <span className="text-sm text-muted-foreground">{t(lang, "common.loading")}</span>
      </div>
    );
  }

  const sidebarItems = [
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
      href: "/admin/dashboard/settings",
      label: t(lang, "admin.sidebarSettings"),
      icon: Settings,
    },
  ];

  return (
    <AdminStatusProvider verifiedStatus={verifiedStatus}>
      <div className="flex gap-6" id="main-content">
        {/* Sidebar */}
        <aside className="w-48 shrink-0 hidden md:block">
          <nav className="sticky top-20 space-y-1">
            {sidebarItems.map((item) => {
              const isActive = pathname === item.href;
              return (
                <Button
                  key={item.href}
                  variant={isActive ? "secondary" : "ghost"}
                  size="sm"
                  onClick={() => router.push(item.href)}
                  className="w-full justify-start cursor-pointer"
                >
                  <item.icon className="h-4 w-4" />
                  {item.label}
                </Button>
              );
            })}
            <Separator className="my-2" />
            <Button
              variant="ghost"
              size="sm"
              onClick={handleLogout}
              className="w-full justify-start cursor-pointer text-muted-foreground hover:text-destructive"
            >
              <LogOut className="h-4 w-4" />
              {t(lang, "admin.logout")}
            </Button>
          </nav>
        </aside>

        {/* Content */}
        <div className="flex-1 min-w-0">{children}</div>
      </div>
    </AdminStatusProvider>
  );
}
