"use client";

import { create } from "zustand";

import type { UploadResponseData } from "@/types/upload";

type UploadPhase = "idle" | "uploading" | "success" | "error";

type UploadStore = {
  phase: UploadPhase;
  progress: number;
  result: UploadResponseData | null;
  error: string | null;
  selectedStorageKey: string;
  start: () => void;
  setProgress: (value: number) => void;
  setSelectedStorageKey: (value: string) => void;
  succeed: (result: UploadResponseData) => void;
  fail: (message: string) => void;
  reset: () => void;
};

export const useUploadStore = create<UploadStore>((set) => ({
  phase: "idle",
  progress: 0,
  result: null,
  error: null,
  selectedStorageKey: "",
  start: () => set({ phase: "uploading", progress: 0, result: null, error: null }),
  setProgress: (value) => set({ progress: value }),
  setSelectedStorageKey: (value) => set({ selectedStorageKey: value }),
  succeed: (result) => set({ phase: "success", progress: 100, result, error: null }),
  fail: (message) => set({ phase: "error", error: message }),
  reset: () => set({ phase: "idle", progress: 0, result: null, error: null })
}));
