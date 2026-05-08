import { create } from "zustand";

import type { PublicRuntimeSettings } from "@/types";

interface UploadState {
  selectedStorageKey: string;
  runtimeSettings: PublicRuntimeSettings | null;
  setSelectedStorageKey: (key: string) => void;
  setRuntimeSettings: (settings: PublicRuntimeSettings | null) => void;
}

export const useUploadStore = create<UploadState>((set) => ({
  selectedStorageKey: "",
  runtimeSettings: null,
  setSelectedStorageKey: (key) => set({ selectedStorageKey: key }),
  setRuntimeSettings: (settings) => set({ runtimeSettings: settings }),
}));
