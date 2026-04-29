import type { Metadata } from "next";
import Script from "next/script";
import { Toaster } from "react-hot-toast";

import { AppHeader } from "@/components/shared/AppHeader";
import { SkipLink } from "@/components/shared/SkipLink";
import { UiPreferenceSync } from "@/components/shared/UiPreferenceSync";
import { createPreferenceInitScript } from "@/lib/preferences";

import "./globals.css";

export const metadata: Metadata = {
  title: "OmePic",
  description: "A full-stack image hosting service"
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html className="dark" data-theme="dark" data-theme-mode="dark" lang="en" suppressHydrationWarning>
      <body className="selection:bg-violet-500/25">
        <Script id="omepic-preferences-init" strategy="beforeInteractive">
          {createPreferenceInitScript()}
        </Script>
        <UiPreferenceSync />
        <SkipLink />
        <AppHeader />
        <main
          className="relative z-10 mx-auto min-h-[calc(100vh-64px)] w-full max-w-[88rem] px-4 pb-14 pt-24 sm:px-6 lg:px-8"
          id="main-content"
          tabIndex={-1}
        >
          {children}
        </main>
        <Toaster
          position="top-right"
          toastOptions={{
            className:
              "glass-panel-strong rounded-2xl border-white/50 text-sm text-ink shadow-glow dark:border-white/10"
          }}
        />
      </body>
    </html>
  );
}
