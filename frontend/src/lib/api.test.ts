import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { adminCreateStorageInstance, adminDeleteImages, adminUpdateSystemSettings } from './api';
import type { RuntimeSettings, StorageInstance } from '@/types';

const jsonResponse = (status: number, payload: unknown) =>
  new Response(JSON.stringify(payload), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });

const storageInstance: Partial<StorageInstance> = {
  name: 'Archive',
  storage_backend: 'local',
  local_storage_path: 'data/archive',
};

const runtimeSettings: RuntimeSettings = {
  site_name: 'OmePic',
  site_tagline: 'Upload and share images',
  public_base_url: '',
  max_upload_size_mb: 20,
  allowed_mime_types: ['image/png'],
  allow_storage_selection: true,
  maintenance_mode: false,
  maintenance_message: '',
  rate_limit_window_minutes: 1,
  rate_limit_max_requests: 120,
  upload_rate_limit_window_minutes: 10,
  upload_rate_limit_max_requests: 20,
};

describe('admin API helpers', () => {
  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn());
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('throws ApiError with backend code and HTTP status when storage creation fails', async () => {
    vi.mocked(fetch).mockResolvedValueOnce(
      jsonResponse(400, {
        success: false,
        error: { code: 'invalid_input', message: 'storage instance name is required' },
      })
    );

    await expect(adminCreateStorageInstance('admin-token', storageInstance)).rejects.toMatchObject({
      name: 'ApiError',
      message: 'storage instance name is required',
      code: 'invalid_input',
      status: 400,
    });
  });

  it('uses the shared admin JSON request contract for image deletion', async () => {
    vi.mocked(fetch).mockResolvedValueOnce(jsonResponse(200, { success: true, data: {} }));

    await adminDeleteImages('admin-token', ['uid-1', 'uid-2']);

    expect(fetch).toHaveBeenCalledWith('http://localhost:8080/admin/images', {
      cache: 'no-store',
      method: 'DELETE',
      headers: {
        Authorization: 'Bearer admin-token',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ uids: ['uid-1', 'uid-2'] }),
    });
  });

  it('preserves typed response data when updating system settings', async () => {
    const system = {
      runtime: runtimeSettings,
      readonly: {
        environment: {
          http_addr: ':8080',
          database_path: 'data/app.db',
          redis_configured: false,
          public_base_url_source: 'request_host',
          env_public_base_url_set: false,
          runtime_public_base_url_set: false,
        },
        security: {
          jwt_secret: { configured: true, using_default: false },
          admin_password: { configured: true, using_default: false },
          uid_encryption_key: { configured: true, using_default: false },
        },
        storage: {
          default_storage_key: 'local',
          storage_config_count: 1,
          allow_storage_selection: true,
        },
        service: {
          health: 'ok',
          maintenance_mode: false,
        },
      },
    };
    vi.mocked(fetch).mockResolvedValueOnce(jsonResponse(200, { success: true, data: system }));

    await expect(adminUpdateSystemSettings('admin-token', runtimeSettings)).resolves.toEqual(system);
  });
});
