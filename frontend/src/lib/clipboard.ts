import { t } from '@/i18n';
import { toast } from '@/stores/toast.svelte';
import type { Language } from '@/types';

export async function copyToClipboard(value: string, language: Language): Promise<boolean> {
  if (typeof navigator === 'undefined' || !navigator.clipboard?.writeText) {
    toast.error(t(language, 'common.copyFailed'));
    return false;
  }

  try {
    await navigator.clipboard.writeText(value);
    toast.success(t(language, 'common.copied'));
    return true;
  } catch {
    toast.error(t(language, 'common.copyFailed'));
    return false;
  }
}
