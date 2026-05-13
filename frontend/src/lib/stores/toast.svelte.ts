export type ToastTone = 'success' | 'error' | 'info';

export type ToastMessage = {
  id: number;
  tone: ToastTone;
  message: string;
};

export const toasts = $state<{ items: ToastMessage[] }>({ items: [] });
let toastId = 0;

import { SvelteMap } from 'svelte/reactivity';

// Track pending auto-removal timers so they can be cleaned up if needed
const toastTimers = new SvelteMap<number, ReturnType<typeof setTimeout>>();

function push(tone: ToastTone, message: string) {
  const item = { id: ++toastId, tone, message };
  toasts.items = [...toasts.items, item];
  if (typeof window !== 'undefined') {
    const timer = setTimeout(() => {
      toastTimers.delete(item.id);
      toasts.items = toasts.items.filter((toast) => toast.id !== item.id);
    }, 3200);
    toastTimers.set(item.id, timer);
  }
}

/** Remove a toast immediately, cancelling its pending auto-removal timer. */
export function dismissToast(id: number) {
  const timer = toastTimers.get(id);
  if (timer !== undefined) {
    clearTimeout(timer);
    toastTimers.delete(id);
  }
  toasts.items = toasts.items.filter((toast) => toast.id !== id);
}

/** Remove all toasts immediately. */
export function clearToasts() {
  for (const timer of toastTimers.values()) clearTimeout(timer);
  toastTimers.clear();
  toasts.items = [];
}

export const toast = {
  success: (message: string) => push('success', message),
  error: (message: string) => push('error', message),
  info: (message: string) => push('info', message),
};
