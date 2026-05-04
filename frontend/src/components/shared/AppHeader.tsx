"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Database, Menu, Monitor, Moon, PanelTop, RefreshCw, Settings, Sun, Upload, UserRound, X } from "lucide-react";
import { useCallback, useEffect, useMemo, useState } from "react";

import { CopyButton } from "@/components/shared/CopyButton";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuTrigger
} from "@/components/ui/DropdownMenu";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/Select";
import { useClientToken } from "@/hooks/useClientToken";
import { useUiTranslations } from "@/hooks/useUiPreferences";
import { publicStorageOptions } from "@/lib/api";
import { cn } from "@/lib/utils";
import { useUploadStore } from "@/stores/upload-store";
import { useUiPreferencesStore } from "@/stores/ui-preferences-store";
import { languages, themeModes, type ThemeMode } from "@/types/preferences";
import type { PublicStorageOption } from "@/types/storage";

const defaultStorageSelectValue = "__default__";

export function AppHeader() {
  const pathname = usePathname();
  const [menuOpen, setMenuOpen] = useState(false);
  const { ready, token } = useClientToken();
  const phase = useUploadStore((state) => state.phase);
  const selectedStorageKey = useUploadStore((state) => state.selectedStorageKey);
  const setSelectedStorageKey = useUploadStore((state) => state.setSelectedStorageKey);
  const language = useUiPreferencesStore((state) => state.language);
  const theme = useUiPreferencesStore((state) => state.theme);
  const setLanguage = useUiPreferencesStore((state) => state.setLanguage);
  const setTheme = useUiPreferencesStore((state) => state.setTheme);
  const t = useUiTranslations();
  const [storageOptions, setStorageOptions] = useState<PublicStorageOption[]>([]);
  const [storageOptionsLoading, setStorageOptionsLoading] = useState(false);
  const [storageOptionsError, setStorageOptionsError] = useState<string | null>(null);
  const navItems = [
    { href: "/", label: t.common.upload, active: pathname === "/" },
    { href: "/history", label: t.common.history, active: pathname === "/history" },
    { href: "/api", label: t.common.api, active: pathname === "/api" },
    { href: "/admin/dashboard", label: t.common.admin, active: Boolean(pathname?.startsWith("/admin")) }
  ];
  const applyStorageOptions = useCallback((items: PublicStorageOption[]) => {
    setStorageOptions(items);
  }, []);

  const loadStorageOptions = useCallback(async (signal?: AbortSignal) => {
    setStorageOptionsLoading(true);
    setStorageOptionsError(null);
    try {
      const result = await publicStorageOptions(signal);
      if (signal?.aborted) {
        return;
      }
      applyStorageOptions(result.items);
    } catch (storageError) {
      if (signal?.aborted) {
        return;
      }
      setStorageOptionsError(storageError instanceof Error ? storageError.message : "");
    } finally {
      if (!signal?.aborted) {
        setStorageOptionsLoading(false);
      }
    }
  }, [applyStorageOptions]);

  useEffect(() => {
    if (pathname !== "/") {
      return;
    }
    const controller = new AbortController();
    void Promise.resolve().then(() => loadStorageOptions(controller.signal));
    return () => {
      controller.abort();
    };
  }, [loadStorageOptions, pathname]);

  useEffect(() => {
    if (
      !storageOptionsLoading &&
      storageOptions.length > 0 &&
      selectedStorageKey &&
      !storageOptions.some((item) => item.storage_key === selectedStorageKey)
    ) {
      setSelectedStorageKey("");
    }
  }, [selectedStorageKey, setSelectedStorageKey, storageOptions, storageOptionsLoading]);

  const defaultStorage = useMemo(
    () => storageOptions.find((item) => item.is_default) ?? null,
    [storageOptions]
  );

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
            {pathname === "/" ? (
              <HeaderStorageSelect
                defaultStorage={defaultStorage}
                disabled={phase === "uploading"}
                error={storageOptionsError}
                loading={storageOptionsLoading}
                onRefresh={() => void loadStorageOptions()}
                options={storageOptions}
                selectedStorageKey={selectedStorageKey}
                setSelectedStorageKey={setSelectedStorageKey}
              />
            ) : null}

            <SettingsMenu
              language={language}
              ready={ready}
              setLanguage={setLanguage}
              setTheme={setTheme}
              theme={theme}
              token={token}
            />

            <Button asChild className="hidden sm:inline-flex">
              <Link href="/" onClick={() => setMenuOpen(false)}>
                <Upload aria-hidden="true" className="h-4 w-4" />
                {t.common.upload}
              </Link>
            </Button>

            <Button asChild aria-label={t.common.admin} size="icon" variant="outline">
              <Link href="/admin/dashboard" onClick={() => setMenuOpen(false)}>
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
      </div>
    </>
  );
}

function HeaderStorageSelect({
  defaultStorage,
  disabled,
  error,
  loading,
  onRefresh,
  options,
  selectedStorageKey,
  setSelectedStorageKey
}: {
  defaultStorage: PublicStorageOption | null;
  disabled: boolean;
  error: string | null;
  loading: boolean;
  onRefresh: () => void;
  options: PublicStorageOption[];
  selectedStorageKey: string;
  setSelectedStorageKey: (value: string) => void;
}) {
  const t = useUiTranslations();
  const selectedStorage = selectedStorageKey
    ? options.find((item) => item.storage_key === selectedStorageKey) ?? null
    : defaultStorage;
  const hintId = "header-storage-hint";
  const errorId = "header-storage-error";
  const statusId = "header-storage-status";
  const descriptionIds = [
    hintId,
    error ? errorId : "",
    loading ? statusId : ""
  ]
    .filter(Boolean)
    .join(" ");

  return (
    <div className="flex items-center gap-2">
      <Select
        disabled={disabled || loading}
        onValueChange={(value) =>
          setSelectedStorageKey(value === defaultStorageSelectValue ? "" : value)
        }
        value={selectedStorageKey || defaultStorageSelectValue}
      >
        <SelectTrigger
          aria-busy={loading}
          aria-describedby={descriptionIds}
          aria-invalid={error ? true : undefined}
          aria-label={t.upload.storageTarget}
          className="h-10 w-10 justify-center bg-background px-0 sm:w-48 sm:justify-between sm:px-3 lg:w-52"
          id="header-storage"
        >
          <Database aria-hidden="true" className="h-4 w-4 text-muted-foreground" />
          <span className="hidden min-w-0 sm:inline">
            <SelectValue />
          </span>
        </SelectTrigger>
        <SelectContent align="end" className="min-w-64">
          <SelectItem value={defaultStorageSelectValue}>
            {defaultStorage
              ? t.upload.defaultStorageOption(defaultStorage.name, defaultStorage.storage_backend)
              : t.upload.backendDefaultStorage}
          </SelectItem>
          {options.map((option) => (
            <SelectItem key={option.storage_key} value={option.storage_key}>
              {option.is_default
                ? t.upload.storageOptionDefault(option.name, option.storage_backend)
                : t.upload.storageOption(option.name, option.storage_backend)}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {selectedStorage ? (
        <Badge className="hidden max-w-28 truncate xl:inline-flex">{selectedStorage.storage_backend}</Badge>
      ) : null}
      <Button
        aria-controls="header-storage"
        aria-label={loading ? t.common.loading : t.common.refresh}
        disabled={disabled || loading}
        onClick={onRefresh}
        size="icon"
        variant="outline"
      >
        <RefreshCw aria-hidden="true" className={loading ? "h-4 w-4 animate-spin" : "h-4 w-4"} />
      </Button>
      <span className="sr-only" id={hintId}>
        {selectedStorage
          ? t.upload.storageSelectionHint(selectedStorage.name, selectedStorage.storage_backend)
          : t.upload.backendDefaultStorage}
      </span>
      {loading ? (
        <span className="sr-only" id={statusId} role="status">
          {t.common.loading}
        </span>
      ) : null}
      {error ? (
        <span className="sr-only" id={errorId} role="alert">
          {error || t.upload.storageOptionsFailed}
        </span>
      ) : null}
    </div>
  );
}

function SettingsMenu({
  language,
  ready,
  setLanguage,
  setTheme,
  theme,
  token
}: {
  language: (typeof languages)[number];
  ready: boolean;
  setLanguage: (language: (typeof languages)[number]) => void;
  setTheme: (theme: ThemeMode) => void;
  theme: ThemeMode;
  token: string | null;
}) {
  const t = useUiTranslations();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button aria-label={t.header.settings} size="icon" variant="outline">
          <Settings aria-hidden="true" className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-[min(calc(100vw-2rem),24rem)] p-4">
        <div className="space-y-4">
          <div className="space-y-2">
            <div className="flex items-center justify-between gap-3">
              <p className="text-sm font-semibold text-foreground">{t.common.clientToken}</p>
              {ready && token ? (
                <CopyButton label={t.common.clientToken} value={token} />
              ) : null}
            </div>
            <p className="break-all rounded-md border border-border bg-muted px-3 py-2 font-mono text-xs text-muted-foreground">
              {ready ? token : t.common.preparingToken}
            </p>
          </div>

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
      </DropdownMenuContent>
    </DropdownMenu>
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
