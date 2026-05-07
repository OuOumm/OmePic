import type { Metadata } from "next";
import "./globals.css";
import { SkipLink } from "@/components/shared/SkipLink";
import { UiPreferenceSync } from "@/components/shared/UiPreferenceSync";
import { Toaster } from "react-hot-toast";

export const metadata: Metadata = {
  title: "OmePic - Image Hosting",
  description: "Upload and share images easily",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" data-theme="dark" data-theme-mode="dark" className="dark" suppressHydrationWarning>
      <head>
        {/* Prevent FOUC: apply dark mode immediately */}
        <script
          dangerouslySetInnerHTML={{
            __html: `
              (function() {
                try {
                  var prefs = JSON.parse(localStorage.getItem('omepic-ui-preferences'));
                  var theme = prefs && prefs.theme;
                  var resolved = theme === 'light' ? 'light' : (theme === 'system' ? (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light') : 'dark');
                  document.documentElement.dataset.theme = resolved;
                  document.documentElement.dataset.themeMode = theme || 'dark';
                  document.documentElement.lang = (prefs && prefs.language) || navigator.language.startsWith('zh') ? 'zh' : 'en';
                  if (resolved === 'dark') document.documentElement.classList.add('dark');
                  else document.documentElement.classList.remove('dark');
                } catch(e) {}
              })();
            `,
          }}
        />
      </head>
      <body className="min-h-screen bg-background text-foreground antialiased">
        <SkipLink />
        <UiPreferenceSync />
        {children}
        <Toaster
          position="top-center"
          toastOptions={{
            duration: 3000,
            style: {
              fontSize: "14px",
              borderRadius: "8px",
            },
          }}
        />
      </body>
    </html>
  );
}
