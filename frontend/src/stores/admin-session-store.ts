import { create } from "zustand";
import { persist } from "zustand/middleware";

interface AdminSessionState {
  token: string | null;
  setToken: (token: string) => void;
  clearToken: () => void;
  hasHydrated: boolean;
  setHasHydrated: (v: boolean) => void;
}

export const useAdminSessionStore = create<AdminSessionState>()(
  persist(
    (set) => ({
      token: null,
      setToken: (token) => set({ token }),
      clearToken: () => set({ token: null }),
      hasHydrated: false,
      setHasHydrated: (v) => set({ hasHydrated: v }),
    }),
    {
      name: "omepic-admin-token",
      partialize: (state) => ({ token: state.token }),
      onRehydrateStorage: () => (state) => {
        state?.setHasHydrated(true);
      },
    }
  )
);
