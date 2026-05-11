import { afterEach, describe, expect, it, vi } from 'vitest';
import { getClientToken } from './client-token';

function createStorage(initial: Record<string, string> = {}) {
  const values = new Map(Object.entries(initial));
  return {
    getItem: vi.fn((key: string) => values.get(key) ?? null),
    setItem: vi.fn((key: string, value: string) => {
      values.set(key, value);
    }),
  };
}

describe('getClientToken', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('returns an empty token outside the browser', () => {
    expect(getClientToken()).toBe('');
  });

  it('reuses an existing browser token', () => {
    const storage = createStorage({ 'omepic-client-token': 'existing-token' });
    vi.stubGlobal('window', {});
    vi.stubGlobal('localStorage', storage);

    expect(getClientToken()).toBe('existing-token');
    expect(storage.setItem).not.toHaveBeenCalled();
  });

  it('generates and persists a browser token when missing', () => {
    const storage = createStorage();
    vi.stubGlobal('window', {});
    vi.stubGlobal('localStorage', storage);
    vi.stubGlobal('crypto', { randomUUID: () => 'generated-token' });

    expect(getClientToken()).toBe('generated-token');
    expect(storage.setItem).toHaveBeenCalledWith('omepic-client-token', 'generated-token');
  });
});
