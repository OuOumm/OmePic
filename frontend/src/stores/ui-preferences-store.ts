import { create } from "zustand";
import { persist } from "zustand/middleware";

import type { Language, Theme } from "@/types";
import { detectLanguage } from "@/lib/i18n";

interface UiPreferencesState {
  language: Language;
  theme: Theme;
  setLanguage: (lang: Language) => void;
  setTheme: (theme: Theme) => void;
  hasHydrated: boolean;
  setHasHydrated: (v: boolean) => void;
}

export const useUiPreferencesStore = create<UiPreferencesState>()(
  persist(
    (set) => ({
      language: "zh" as Language,
      theme: "dark" as Theme,
      setLanguage: (language) => set({ language }),
      setTheme: (theme) => set({ theme }),
      hasHydrated: false,
      setHasHydrated: (v) => set({ hasHydrated: v }),
    }),
    {
      name: "omepic-ui-preferences",
      partialize: (state) => ({
        language: state.language,
        theme: state.theme,
      }),
      onRehydrateStorage: () => (state) => {
        if (state) {
          // Validate stored values
          if (!["en", "zh"].includes(state.language)) {
            state.setLanguage(detectLanguage());
          }
          if (!["light", "dark", "system"].includes(state.theme)) {
            state.setTheme("dark");
          }
          state.setHasHydrated(true);
        }
      },
    }
  )
);
