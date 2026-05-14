import { ApiError } from '@/api';
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

const apiErrorTranslationKeys: Record<string, string> = {
  'forbidden:current password is incorrect': 'admin.passwordCurrentIncorrect',
  'invalid_input:new password must be at least 8 characters and include uppercase, lowercase, and symbol characters': 'admin.passwordStrengthError',
};

export function errorMessage(err: unknown, language: Language, fallbackKey = 'common.error'): string {
  if (err instanceof ApiError) {
    const translationKey = apiErrorTranslationKey(err);
    if (translationKey) return t(language, translationKey);
  }
  return err instanceof Error ? err.message : t(language, fallbackKey);
}

function apiErrorTranslationKey(err: ApiError): string | undefined {
  if (!err.code) return undefined;
  return apiErrorTranslationKeys[`${err.code}:${err.message}`];
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
