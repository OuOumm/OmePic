"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";

import { UI_PREFERENCES_STORAGE_KEY } from "@/lib/preferences";
import type { Language, ThemeMode } from "@/types/preferences";

type UiPreferencesStore = {
  language: Language;
  theme: ThemeMode;
  hasHydrated: boolean;
  setLanguage: (language: Language) => void;
  setTheme: (theme: ThemeMode) => void;
  setHasHydrated: (value: boolean) => void;
};

export const useUiPreferencesStore = create<UiPreferencesStore>()(
  persist(
    (set) => ({
      language: "en",
      theme: "dark",
      hasHydrated: false,
      setLanguage: (language) => set({ language }),
      setTheme: (theme) => set({ theme }),
      setHasHydrated: (value) => set({ hasHydrated: value })
    }),
    {
      name: UI_PREFERENCES_STORAGE_KEY,
      partialize: (state) => ({
        language: state.language,
        theme: state.theme
      }),
      onRehydrateStorage: () => (state) => {
        state?.setHasHydrated(true);
      }
    }
  )
);
