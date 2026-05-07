"use client";

import { useState, useCallback } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { Button } from "@/components/ui/Button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/DropdownMenu";
import { useUiPreferencesStore } from "@/stores/ui-preferences-store";
import { getClientToken } from "@/lib/preferences";
import { t } from "@/lib/i18n";
import {
  Upload,
  History,
  Code,
  Settings,
  Sun,
  Moon,
  Monitor,
  Languages,
  Copy,
  Check,
} from "lucide-react";
import type { Language, Theme } from "@/types";
import { cn } from "@/lib/utils";

export function AppHeader() {
  const pathname = usePathname();
  const router = useRouter();
  const language = useUiPreferencesStore((state) => state.language);
  const theme = useUiPreferencesStore((state) => state.theme);
  const setLanguage = useUiPreferencesStore((state) => state.setLanguage);
  const setTheme = useUiPreferencesStore((state) => state.setTheme);
  const [tokenCopied, setTokenCopied] = useState(false);

  const handleCopyToken = useCallback(async () => {
    const token = getClientToken();
    try {
      await navigator.clipboard.writeText(token);
    } catch {
      // fallback
    }
    setTokenCopied(true);
    setTimeout(() => setTokenCopied(false), 2000);
  }, []);

  const navItems = [
    { href: "/", label: t(language, "nav.upload"), icon: Upload },
    { href: "/history", label: t(language, "nav.history"), icon: History },
    { href: "/api", label: t(language, "nav.api"), icon: Code },
    { href: "/admin/dashboard", label: t(language, "nav.admin"), icon: Settings },
  ];

  const themeOptions: { key: Theme; icon: typeof Sun; label: string }[] = [
    { key: "light", icon: Sun, label: t(language, "common.themeLight") },
    { key: "dark", icon: Moon, label: t(language, "common.themeDark") },
    { key: "system", icon: Monitor, label: t(language, "common.themeSystem") },
  ];

  return (
    <header className="sticky top-0 z-40 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-14 items-center justify-between max-w-6xl mx-auto px-4">
        {/* Logo + Nav */}
        <div className="flex items-center gap-1">
          <Link href="/" className="font-bold text-lg mr-4 hover:text-primary transition-colors">
            OmePic
          </Link>
          <nav className="hidden sm:flex items-center gap-1" aria-label="Main navigation">
            {navItems.map((item) => {
              const isActive = pathname === item.href ||
                (item.href !== "/" && pathname.startsWith(item.href));
              return (
                <Button
                  key={item.href}
                  variant={isActive ? "secondary" : "ghost"}
                  size="sm"
                  onClick={() => router.push(item.href)}
                  className="cursor-pointer"
                >
                  <item.icon className="h-4 w-4" />
                  <span className="hidden md:inline">{item.label}</span>
                </Button>
              );
            })}
          </nav>
        </div>

        {/* Controls */}
        <div className="flex items-center gap-1">
          {/* Copy token */}
          <Button
            variant="ghost"
            size="sm"
            onClick={handleCopyToken}
            className="cursor-pointer"
            aria-label={t(language, "common.copyToken")}
          >
            {tokenCopied ? <Check className="h-3.5 w-3.5 text-green-500" /> : <Copy className="h-3.5 w-3.5" />}
            <span className="hidden md:inline text-xs">{t(language, "common.token")}</span>
          </Button>

          {/* Language */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="sm" className="cursor-pointer gap-1">
                <Languages className="h-4 w-4" />
                <span className="hidden md:inline text-xs">{language === "zh" ? "中文" : "EN"}</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => setLanguage("zh" as Language)}>中文</DropdownMenuItem>
              <DropdownMenuItem onClick={() => setLanguage("en" as Language)}>English</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>

          {/* Theme */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8 cursor-pointer">
                {theme === "dark" ? <Moon className="h-4 w-4" /> : theme === "light" ? <Sun className="h-4 w-4" /> : <Monitor className="h-4 w-4" />}
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {themeOptions.map((opt) => (
                <DropdownMenuItem
                  key={opt.key}
                  onClick={() => setTheme(opt.key)}
                >
                  <opt.icon className="h-4 w-4" />
                  {opt.label}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>

      {/* Mobile nav */}
      <nav className="sm:hidden flex items-center justify-around border-t px-2 py-1" aria-label="Mobile navigation">
        {navItems.map((item) => {
          const isActive = pathname === item.href ||
            (item.href !== "/" && pathname.startsWith(item.href));
          return (
            <Button
              key={item.href}
              variant="ghost"
              size="sm"
              onClick={() => router.push(item.href)}
              className={cn("flex-col h-auto py-1 gap-0.5 cursor-pointer", isActive && "text-primary")}
            >
              <item.icon className="h-4 w-4" />
              <span className="text-[10px]">{item.label}</span>
            </Button>
          );
        })}
      </nav>
    </header>
  );
}
