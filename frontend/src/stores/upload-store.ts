import { create } from "zustand";

interface UploadState {
  selectedStorageKey: string;
  setSelectedStorageKey: (key: string) => void;
}

export const useUploadStore = create<UploadState>((set) => ({
  selectedStorageKey: "",
  setSelectedStorageKey: (key) => set({ selectedStorageKey: key }),
}));
