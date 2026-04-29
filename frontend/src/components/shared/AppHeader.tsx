"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
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
      <header className="fixed left-0 right-0 top-0 z-50 h-16 border-b border-slate-200/35 bg-white/65 shadow-sm backdrop-blur-2xl transition-all duration-300 dark:border-slate-700/25 dark:bg-slate-950/65 dark:shadow-slate-950/20">
        <div className="mx-auto flex h-full max-w-[88rem] items-center justify-between gap-3 px-4 sm:px-6 lg:px-8">
          <Link
            className="flex min-w-0 items-center gap-2.5 rounded-xl focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-violet-400 focus-visible:ring-offset-2 focus-visible:ring-offset-surface"
            href="/"
            onClick={() => setMenuOpen(false)}
          >
            <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-violet-500 to-cyan-400 text-white shadow-lg shadow-violet-500/25">
              <LogoIcon />
            </span>
            <span className="hidden min-w-0 sm:block">
              <span className="block truncate text-xl font-bold tracking-tight gradient-text">OmePic</span>
              <span className="block -mt-1 text-[11px] font-medium text-muted">{t.header.tagline}</span>
            </span>
            <span className="text-lg font-bold tracking-tight gradient-text sm:hidden">OP</span>
          </Link>

          <nav aria-label={t.header.navLabel} className="hidden items-center gap-1 rounded-full border border-white/45 bg-white/50 p-1 shadow-sm backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/40 md:flex">
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

            <Link
              className="hidden min-h-10 items-center justify-center gap-2 rounded-2xl border border-violet-400/30 bg-gradient-to-r from-violet-600 to-cyan-600 px-4 py-2 text-sm font-semibold text-white shadow-lg shadow-violet-500/25 transition-all duration-300 hover:-translate-y-0.5 hover:from-violet-500 hover:to-cyan-500 hover:shadow-xl hover:shadow-violet-500/30 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-violet-400 focus-visible:ring-offset-2 focus-visible:ring-offset-surface sm:inline-flex"
              href="/"
              onClick={() => setMenuOpen(false)}
            >
              <PlusIcon />
              {t.common.upload}
            </Link>

            <Link
              aria-label={t.common.admin}
              className="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-br from-violet-400 via-fuchsia-400 to-cyan-400 p-[2px] shadow-md shadow-violet-400/25 transition-transform duration-200 hover:scale-105 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-violet-400 focus-visible:ring-offset-2 focus-visible:ring-offset-surface"
              href="/admin/login"
              onClick={() => setMenuOpen(false)}
            >
              <span className="flex h-full w-full items-center justify-center rounded-full bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-300">
                <UserIcon />
              </span>
            </Link>

            <Button
              aria-expanded={menuOpen}
              aria-label={menuOpen ? t.header.closeMenu : t.header.openMenu}
              className="md:hidden"
              onClick={() => setMenuOpen((value) => !value)}
              size="icon"
              variant="secondary"
            >
              {menuOpen ? <CloseIcon /> : <MenuIcon />}
            </Button>
          </div>
        </div>
      </header>

      <div
        className={cn(
          "fixed inset-0 z-40 bg-black/40 backdrop-blur-sm transition-opacity duration-300 md:hidden",
          menuOpen ? "opacity-100" : "pointer-events-none opacity-0"
        )}
        onClick={() => setMenuOpen(false)}
      />
      <div
        className={cn(
          "fixed left-0 right-0 top-16 z-50 border-b border-slate-200/40 bg-white/88 px-4 py-4 shadow-xl backdrop-blur-2xl transition-all duration-300 md:hidden dark:border-slate-700/30 dark:bg-slate-950/90",
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
        <div className="mt-4 grid gap-3 border-t border-slate-200/60 pt-4 dark:border-slate-700/40">
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
      <div className="flex items-center justify-between gap-2 text-xs text-muted" role="group" aria-label={label}>
      <span className="font-semibold">{label}</span>
      <div className="flex rounded-full border border-white/50 bg-white/60 p-1 shadow-sm backdrop-blur-xl dark:border-white/10 dark:bg-slate-950/50">
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
        "rounded-full px-3 py-1.5 text-xs font-semibold transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-violet-400 focus-visible:ring-offset-2 focus-visible:ring-offset-surface",
        active
          ? "bg-gradient-to-r from-violet-600 to-cyan-600 text-white shadow-md shadow-violet-500/20"
          : "text-muted hover:bg-white/70 hover:text-violet-700 dark:hover:bg-white/10 dark:hover:text-violet-200"
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
        "flex h-8 w-8 items-center justify-center rounded-full transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-violet-400 focus-visible:ring-offset-2 focus-visible:ring-offset-surface",
        active
          ? "bg-gradient-to-r from-violet-600 to-cyan-600 text-white shadow-md shadow-violet-500/20"
          : "text-muted hover:bg-white/70 hover:text-violet-700 dark:hover:bg-white/10 dark:hover:text-violet-200"
      )}
      onClick={onClick}
      title={label}
      type="button"
    >
      {mode === "light" ? <SunIcon /> : mode === "dark" ? <MoonIcon /> : <MonitorIcon />}
    </button>
  );
}

function navLinkClass(active: boolean) {
  return cn(
    "rounded-full px-4 py-2 text-sm font-semibold transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-violet-400 focus-visible:ring-offset-2 focus-visible:ring-offset-surface",
    active
      ? "bg-violet-500/10 text-violet-700 shadow-sm dark:bg-violet-500/15 dark:text-violet-200"
      : "text-muted hover:bg-white/60 hover:text-violet-700 dark:hover:bg-white/10 dark:hover:text-violet-200"
  );
}

function LogoIcon() {
  return (
    <svg aria-hidden="true" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.2}>
      <path d="m4 16 4.6-4.6a2 2 0 0 1 2.8 0L16 16m-2-2 1.6-1.6a2 2 0 0 1 2.8 0L20 14M14 8h.01M6 20h12a2 2 0 0 0 2-2V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2Z" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function PlusIcon() {
  return (
    <svg aria-hidden="true" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
      <path d="M12 4v16m8-8H4" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function UserIcon() {
  return (
    <svg aria-hidden="true" className="h-[18px] w-[18px]" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="M15.75 7.5a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.5 20.25a7.5 7.5 0 0 1 15 0" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function MenuIcon() {
  return (
    <svg aria-hidden="true" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.2}>
      <path d="M4 6h16M4 12h16M4 18h16" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function CloseIcon() {
  return (
    <svg aria-hidden="true" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.2}>
      <path d="M6 6l12 12M18 6 6 18" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function SunIcon() {
  return (
    <svg aria-hidden="true" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="M12 3v2m0 14v2m9-9h-2M5 12H3m15.4-6.4-1.4 1.4M7 17l-1.4 1.4m12.8 0L17 17M7 7 5.6 5.6M16 12a4 4 0 1 1-8 0 4 4 0 0 1 8 0Z" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function MoonIcon() {
  return (
    <svg aria-hidden="true" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="M21 12.8A8.5 8.5 0 1 1 11.2 3a6.5 6.5 0 0 0 9.8 9.8Z" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

function MonitorIcon() {
  return (
    <svg aria-hidden="true" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path d="M4 5.5h16v11H4zM9 21h6m-3-4.5V21" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}
