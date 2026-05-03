"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Menu, Monitor, Moon, PanelTop, Sun, Upload, UserRound, X } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/Button";
import { useUiTranslations } from "@/hooks/useUiPreferences";
import { cn } from "@/lib/utils";
import { useUiPreferencesStore } from "@/stores/ui-preferences-store";
import { languages, themeModes, type ThemeMode } from "@/types/preferences";

export function AppHeader() {
  const pathname = usePathname();
  const [menuOpen, setMenuOpen] = useState(false);
  const language = useUiPreferencesStore((state) => state.language);
  const theme = useUiPreferencesStore((state) => state.theme);
  const setLanguage = useUiPreferencesStore((state) => state.setLanguage);
  const setTheme = useUiPreferencesStore((state) => state.setTheme);
  const t = useUiTranslations();
  const navItems = [
    { href: "/", label: t.common.upload, active: pathname === "/" },
    { href: "/history", label: t.common.history, active: pathname === "/history" },
    { href: "/api", label: t.common.api, active: pathname === "/api" },
    { href: "/admin/login", label: t.common.admin, active: Boolean(pathname?.startsWith("/admin")) }
  ];

  return (
    <>
      <header className="fixed left-0 right-0 top-0 z-50 h-16 border-b border-border bg-background/95 shadow-sm supports-[backdrop-filter]:bg-background/80 supports-[backdrop-filter]:backdrop-blur">
        <div className="mx-auto flex h-full max-w-[88rem] items-center justify-between gap-3 px-4 sm:px-6 lg:px-8">
          <Link
            className="flex min-w-0 items-center gap-2.5 rounded-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
            href="/"
            onClick={() => setMenuOpen(false)}
          >
            <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md border border-border bg-card text-foreground shadow-sm">
              <PanelTop aria-hidden="true" className="h-5 w-5" />
            </span>
            <span className="hidden min-w-0 sm:block">
              <span className="block truncate text-lg font-semibold tracking-tight text-foreground">OmePic</span>
              <span className="block -mt-0.5 text-[11px] font-medium text-muted-foreground">{t.header.tagline}</span>
            </span>
            <span className="text-sm font-semibold tracking-tight text-foreground sm:hidden">OP</span>
          </Link>

          <nav aria-label={t.header.navLabel} className="hidden items-center gap-1 rounded-md border border-border bg-muted p-1 md:flex">
            {navItems.map((item) => (
              <Link
                aria-current={item.active ? "page" : undefined}
                className={navLinkClass(item.active)}
                href={item.href}
                key={item.href}
                onClick={() => setMenuOpen(false)}
              >
                {item.label}
              </Link>
            ))}
          </nav>

          <div className="flex shrink-0 items-center gap-2">
            <div className="hidden items-center gap-2 xl:flex">
              <PreferenceGroup label={t.header.languageLabel}>
                {languages.map((value) => (
                  <PreferenceToggle
                    active={language === value}
                    key={value}
                    onClick={() => setLanguage(value)}
                  >
                    {t.header.languages[value]}
                  </PreferenceToggle>
                ))}
              </PreferenceGroup>
              <PreferenceGroup label={t.header.themeLabel}>
                {themeModes.map((value) => (
                  <ThemeToggle
                    active={theme === value}
                    key={value}
                    label={t.header.themes[value]}
                    mode={value}
                    onClick={() => setTheme(value)}
                  />
                ))}
              </PreferenceGroup>
            </div>

            <Button asChild className="hidden sm:inline-flex">
              <Link href="/" onClick={() => setMenuOpen(false)}>
                <Upload aria-hidden="true" className="h-4 w-4" />
                {t.common.upload}
              </Link>
            </Button>

            <Button asChild aria-label={t.common.admin} size="icon" variant="outline">
              <Link href="/admin/login" onClick={() => setMenuOpen(false)}>
                <UserRound aria-hidden="true" className="h-4 w-4" />
              </Link>
            </Button>

            <Button
              aria-expanded={menuOpen}
              aria-label={menuOpen ? t.header.closeMenu : t.header.openMenu}
              className="md:hidden"
              onClick={() => setMenuOpen((value) => !value)}
              size="icon"
              variant="outline"
            >
              {menuOpen ? <X aria-hidden="true" className="h-4 w-4" /> : <Menu aria-hidden="true" className="h-4 w-4" />}
            </Button>
          </div>
        </div>
      </header>

      <div
        className={cn(
          "fixed inset-0 z-40 bg-background/80 transition-opacity md:hidden",
          menuOpen ? "opacity-100" : "pointer-events-none opacity-0"
        )}
        onClick={() => setMenuOpen(false)}
      />
      <div
        className={cn(
          "fixed left-0 right-0 top-16 z-50 border-b border-border bg-background px-4 py-4 shadow-md transition-all md:hidden",
          menuOpen ? "translate-y-0 opacity-100" : "pointer-events-none -translate-y-4 opacity-0"
        )}
      >
        <nav aria-label={t.header.navLabel} className="grid gap-1">
          {navItems.map((item) => (
            <Link
              aria-current={item.active ? "page" : undefined}
              className={navLinkClass(item.active)}
              href={item.href}
              key={item.href}
              onClick={() => setMenuOpen(false)}
            >
              {item.label}
            </Link>
          ))}
        </nav>
        <div className="mt-4 grid gap-3 border-t border-border pt-4">
          <PreferenceGroup label={t.header.languageLabel}>
            {languages.map((value) => (
              <PreferenceToggle
                active={language === value}
                key={value}
                onClick={() => setLanguage(value)}
              >
                {t.header.languages[value]}
              </PreferenceToggle>
            ))}
          </PreferenceGroup>
          <PreferenceGroup label={t.header.themeLabel}>
            {themeModes.map((value) => (
              <ThemeToggle
                active={theme === value}
                key={value}
                label={t.header.themes[value]}
                mode={value}
                onClick={() => setTheme(value)}
              />
            ))}
          </PreferenceGroup>
        </div>
      </div>
    </>
  );
}

function PreferenceGroup({
  children,
  label
}: {
  children: React.ReactNode;
  label: string;
}) {
  return (
    <div className="flex items-center justify-between gap-2 text-xs text-muted-foreground" role="group" aria-label={label}>
      <span className="font-medium">{label}</span>
      <div className="flex rounded-md border border-border bg-muted p-1">
        {children}
      </div>
    </div>
  );
}

function PreferenceToggle({
  active,
  children,
  onClick
}: {
  active: boolean;
  children: React.ReactNode;
  onClick: () => void;
}) {
  return (
    <button
      aria-pressed={active}
      className={cn(
        "rounded-sm px-2.5 py-1.5 text-xs font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
        active ? "bg-background text-foreground shadow-sm" : "text-muted-foreground hover:bg-background/70 hover:text-foreground"
      )}
      onClick={onClick}
      type="button"
    >
      {children}
    </button>
  );
}

function ThemeToggle({
  active,
  label,
  mode,
  onClick
}: {
  active: boolean;
  label: string;
  mode: ThemeMode;
  onClick: () => void;
}) {
  return (
    <button
      aria-label={label}
      aria-pressed={active}
      className={cn(
        "flex h-8 w-8 items-center justify-center rounded-sm transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
        active ? "bg-background text-foreground shadow-sm" : "text-muted-foreground hover:bg-background/70 hover:text-foreground"
      )}
      onClick={onClick}
      title={label}
      type="button"
    >
      {mode === "light" ? <Sun aria-hidden="true" className="h-4 w-4" /> : mode === "dark" ? <Moon aria-hidden="true" className="h-4 w-4" /> : <Monitor aria-hidden="true" className="h-4 w-4" />}
    </button>
  );
}

function navLinkClass(active: boolean) {
  return cn(
    "rounded-sm px-3 py-1.5 text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
    active ? "bg-background text-foreground shadow-sm" : "text-muted-foreground hover:bg-background/70 hover:text-foreground"
  );
}
