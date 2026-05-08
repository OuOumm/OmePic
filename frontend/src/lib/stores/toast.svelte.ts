export type ToastTone = 'success' | 'error' | 'info';

export type ToastMessage = {
  id: number;
  tone: ToastTone;
  message: string;
};

export const toasts = $state<{ items: ToastMessage[] }>({ items: [] });
let toastId = 0;

function push(tone: ToastTone, message: string) {
  const item = { id: ++toastId, tone, message };
  toasts.items = [...toasts.items, item];
  if (typeof window !== 'undefined') {
    window.setTimeout(() => {
      toasts.items = toasts.items.filter((toast) => toast.id !== item.id);
    }, 3200);
  }
}

export const toast = {
  success: (message: string) => push('success', message),
  error: (message: string) => push('error', message),
  info: (message: string) => push('info', message),
};
