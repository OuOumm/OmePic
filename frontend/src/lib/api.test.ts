import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { adminCreateStorageInstance, adminDeleteImages, adminGetImages, adminGetStatus, adminUpdateSystemSettings, deleteImageByUid } from './api';
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

  it('uses the shared public image path for user deletion', async () => {
    vi.mocked(fetch).mockResolvedValueOnce(jsonResponse(200, { success: true, data: {} }));

    await deleteImageByUid('uid-1', 'client-token');

    expect(fetch).toHaveBeenCalledWith('http://localhost:8080/i/uid-1.avif', {
      cache: 'no-store',
      method: 'DELETE',
      headers: { 'X-Token': 'client-token' },
    });
  });

  it('passes AbortSignal through admin image listing requests', async () => {
    const controller = new AbortController();
    vi.mocked(fetch).mockResolvedValueOnce(jsonResponse(200, { success: true, data: { items: [], total: 0 } }));

    await adminGetImages('admin-token', 2, 30, 'needle', controller.signal);

    const [url, options] = vi.mocked(fetch).mock.calls[0];
    expect(url).toBe('http://localhost:8080/admin/images?page=2&pageSize=30&search=needle');
    expect(options).toMatchObject({
      cache: 'no-store',
      headers: {
        Authorization: 'Bearer admin-token',
        'Content-Type': 'application/json',
      },
      signal: controller.signal,
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

  it('deduplicates concurrent GET requests within the same auth scope', async () => {
    vi.mocked(fetch).mockResolvedValueOnce(jsonResponse(200, {
      success: true,
      data: { total_images: 1, total_storage_size: 2, today_uploads: 3, unique_tokens: 4 },
    }));

    const [left, right] = await Promise.all([
      adminGetStatus('admin-token'),
      adminGetStatus('admin-token'),
    ]);

    expect(fetch).toHaveBeenCalledTimes(1);
    expect(left).toEqual(right);
  });

  it('keeps GET request deduplication scoped by auth headers', async () => {
    vi.mocked(fetch)
      .mockResolvedValueOnce(jsonResponse(200, {
        success: true,
        data: { total_images: 1, total_storage_size: 2, today_uploads: 3, unique_tokens: 4 },
      }))
      .mockResolvedValueOnce(jsonResponse(200, {
        success: true,
        data: { total_images: 5, total_storage_size: 6, today_uploads: 7, unique_tokens: 8 },
      }));

    await Promise.all([
      adminGetStatus('admin-token-a'),
      adminGetStatus('admin-token-b'),
    ]);

    expect(fetch).toHaveBeenCalledTimes(2);
  });
});
