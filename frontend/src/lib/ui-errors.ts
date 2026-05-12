import { t } from '@/i18n';
import { toast } from '@/stores/toast.svelte';
import type { Language } from '@/types';

type SuccessMessage<T> = string | ((result: T) => string);

type RunAsyncActionOptions<T> = {
  action: () => Promise<T>;
  language: Language;
  setBusy?: (busy: boolean) => void;
  successMessage?: SuccessMessage<T>;
  fallbackErrorKey?: string;
  onSuccess?: (result: T) => void | Promise<void>;
  onError?: (err: unknown) => void | Promise<void>;
};

export function errorMessage(err: unknown, language: Language, fallbackKey = 'common.error'): string {
  return err instanceof Error ? err.message : t(language, fallbackKey);
}

export function toastApiError(err: unknown, language: Language, fallbackKey = 'common.error') {
  toast.error(errorMessage(err, language, fallbackKey));
}

export async function runAsyncAction<T>({
  action,
  language,
  setBusy,
  successMessage,
  fallbackErrorKey = 'common.error',
  onSuccess,
  onError,
}: RunAsyncActionOptions<T>): Promise<T | undefined> {
  setBusy?.(true);
  try {
    const result = await action();
    const message = typeof successMessage === 'function' ? successMessage(result) : successMessage;
    if (message) toast.success(message);
    await onSuccess?.(result);
    return result;
  } catch (err) {
    if (onError) {
      await onError(err);
    } else {
      toastApiError(err, language, fallbackErrorKey);
    }
    return undefined;
  } finally {
    setBusy?.(false);
  }
}
