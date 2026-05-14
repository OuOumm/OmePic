import { afterEach, describe, expect, it, vi } from 'vitest';
import { ApiError } from './api';
import { errorMessage, runAsyncAction, toastApiError } from './ui-errors';
import { toast } from './stores/toast.svelte';

vi.mock('./stores/toast.svelte', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
  },
}));

describe('errorMessage', () => {
  it('uses Error messages before translated fallbacks', () => {
    expect(errorMessage(new Error('from api'), 'en')).toBe('from api');
    expect(errorMessage('unknown', 'zh')).toBe('操作失败');
  });

  it('localizes known admin password API errors', () => {
    expect(errorMessage(new ApiError('current password is incorrect', { code: 'forbidden', status: 403 }), 'zh')).toBe('当前密码错误');
    expect(errorMessage(new ApiError('new password must be at least 8 characters and include uppercase, lowercase, and symbol characters', { code: 'invalid_input', status: 400 }), 'zh')).toBe('新密码至少 8 位，并包含大写字母、小写字母和符号');
  });
});

describe('toastApiError', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it('reports a translated fallback for unknown errors', () => {
    toastApiError(null, 'en', 'admin.imagesDeleteError');

    expect(toast.error).toHaveBeenCalledWith('Failed to delete images');
  });
});

describe('runAsyncAction', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it('sets busy state, reports success, and runs onSuccess', async () => {
    const busyStates: boolean[] = [];
    const onSuccess = vi.fn();

    const result = await runAsyncAction({
      language: 'en',
      setBusy: (busy) => busyStates.push(busy),
      successMessage: (value: number) => `Saved ${value}`,
      action: () => Promise.resolve(3),
      onSuccess,
    });

    expect(result).toBe(3);
    expect(busyStates).toEqual([true, false]);
    expect(toast.success).toHaveBeenCalledWith('Saved 3');
    expect(onSuccess).toHaveBeenCalledWith(3);
  });

  it('reports errors and still clears busy state', async () => {
    const busyStates: boolean[] = [];

    const result = await runAsyncAction({
      language: 'zh',
      setBusy: (busy) => busyStates.push(busy),
      fallbackErrorKey: 'admin.imagesDeleteError',
      action: () => Promise.reject(new Error('denied')),
    });

    expect(result).toBeUndefined();
    expect(busyStates).toEqual([true, false]);
    expect(toast.error).toHaveBeenCalledWith('denied');
  });

  it('lets callers override error handling', async () => {
    const onError = vi.fn();

    await runAsyncAction({
      language: 'en',
      action: () => Promise.reject(new Error('boom')),
      onError,
    });

    expect(onError).toHaveBeenCalledWith(expect.any(Error));
    expect(toast.error).not.toHaveBeenCalled();
  });
});
