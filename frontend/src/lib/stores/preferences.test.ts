import { beforeEach, describe, expect, it, vi } from 'vitest';

const adminTokenKey = 'omepic-admin-token';
const uiPreferencesKey = 'omepic-ui-preferences';
const storage = new Map<string, string>();

Object.defineProperty(globalThis, 'localStorage', {
  configurable: true,
  value: {
    clear: () => storage.clear(),
    getItem: (key: string) => storage.get(key) ?? null,
    removeItem: (key: string) => storage.delete(key),
    setItem: (key: string, value: string) => storage.set(key, value),
  },
});

Object.defineProperty(globalThis, 'window', {
  configurable: true,
  value: {
    localStorage: globalThis.localStorage,
  },
});

Object.defineProperty(globalThis, 'document', {
  configurable: true,
  value: {
    documentElement: {
      lang: '',
    },
  },
});

async function loadPreferencesStore() {
  vi.resetModules();
  return import('./preferences.svelte');
}

describe('preferences store', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.lang = '';
  });

  it('defaults the theme preference to system', async () => {
    const { preferences } = await loadPreferencesStore();

    expect(preferences.theme).toBe('system');
  });

  it('normalizes invalid stored themes to system', async () => {
    localStorage.setItem(uiPreferencesKey, '{"theme":"unknown"}');

    const { preferences } = await loadPreferencesStore();

    expect(preferences.theme).toBe('system');
  });

  it('restores the admin token from localStorage when the store loads', async () => {
    localStorage.setItem(adminTokenKey, 'persisted-token');

    const { preferences } = await loadPreferencesStore();

    expect(preferences.adminToken).toBe('persisted-token');
  });

  it('persists and clears the admin token through the shared setters', async () => {
    const { clearAdminToken, preferences, setAdminToken } = await loadPreferencesStore();

    setAdminToken('next-token');

    expect(preferences.adminToken).toBe('next-token');
    expect(localStorage.getItem(adminTokenKey)).toBe('next-token');

    clearAdminToken();

    expect(preferences.adminToken).toBeNull();
    expect(localStorage.getItem(adminTokenKey)).toBeNull();
  });
});
