"use client";

import { useSyncExternalStore } from "react";

import { ensureClientToken } from "@/lib/tokens";

function subscribe() {
  return () => {};
}

export function useClientToken() {
  const token = useSyncExternalStore(
    subscribe,
    () => ensureClientToken(),
    () => ""
  );

  return { token, ready: token !== "" };
}
