import { afterEach, describe, expect, it, vi } from 'vitest';
import { copyToClipboard } from './clipboard';
import { toast, toasts } from './stores/toast.svelte';

vi.mock('./stores/toast.svelte', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
  },
  toasts: { items: [] },
}));

describe('copyToClipboard', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
    toasts.items = [];
  });

  it('writes to the clipboard and reports success', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    vi.stubGlobal('navigator', { clipboard: { writeText } });

    await expect(copyToClipboard('hello', 'en')).resolves.toBe(true);

    expect(writeText).toHaveBeenCalledWith('hello');
    expect(toast.success).toHaveBeenCalledWith('Copied!');
  });

  it('reports failure when clipboard access is unavailable', async () => {
    vi.stubGlobal('navigator', {});

    await expect(copyToClipboard('hello', 'zh')).resolves.toBe(false);

    expect(toast.error).toHaveBeenCalledWith('复制失败');
  });

  it('reports failure when clipboard write rejects', async () => {
    const writeText = vi.fn().mockRejectedValue(new Error('denied'));
    vi.stubGlobal('navigator', { clipboard: { writeText } });

    await expect(copyToClipboard('hello', 'en')).resolves.toBe(false);

    expect(toast.error).toHaveBeenCalledWith('Copy failed');
  });
});
