"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";

type AdminSessionStore = {
  token: string;
  hasHydrated: boolean;
  setToken: (token: string) => void;
  clearToken: () => void;
  setHasHydrated: (value: boolean) => void;
};

export const useAdminSessionStore = create<AdminSessionStore>()(
  persist(
    (set) => ({
      token: "",
      hasHydrated: false,
      setToken: (token) => set({ token }),
      clearToken: () => set({ token: "" }),
      setHasHydrated: (value) => set({ hasHydrated: value })
    }),
    {
      name: "omepic-admin-session",
      onRehydrateStorage: () => (state) => {
        state?.setHasHydrated(true);
      }
    }
  )
);
