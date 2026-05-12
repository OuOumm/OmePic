import { describe, expect, it } from 'vitest';
import { getApiExampleBaseUrl, getImagePath, getImageUrl, isAllowedImageMimeType, normalizeDownloadFilename, safeImageUrl } from './utils';

describe('image URL helpers', () => {
  it('builds public image paths from canonical UIDs', () => {
    expect(getImagePath('uid-1')).toBe('/i/uid-1.avif');
    expect(getImageUrl('uid-1')).toBe('http://localhost:8080/i/uid-1.avif');
  });

  it('uses custom runtime base URL for API examples before falling back to the current origin', () => {
    expect(getApiExampleBaseUrl('https://cdn.example.com/')).toBe('https://cdn.example.com');
    expect(getApiExampleBaseUrl('')).toBe('$ORIGIN');
  });
});

describe('safeImageUrl', () => {
  it('allows relative and same-origin image URLs', () => {
    expect(safeImageUrl('/i/demo.avif', 'https://example.test')).toBe('/i/demo.avif');
    expect(safeImageUrl('https://example.test/i/demo.avif', 'https://example.test')).toBe('https://example.test/i/demo.avif');
  });

  it('rejects javascript, data, and cross-origin URLs', () => {
    expect(safeImageUrl('javascript:alert(1)', 'https://example.test')).toBeNull();
    expect(safeImageUrl('data:image/svg+xml,<svg></svg>', 'https://example.test')).toBeNull();
    expect(safeImageUrl('https://evil.test/i/demo.avif', 'https://example.test')).toBeNull();
  });
});

describe('isAllowedImageMimeType', () => {
  it('uses the configured allow-list and rejects svg', () => {
    const allowed = ['image/png', 'image/jpeg', 'image/svg+xml'];

    expect(isAllowedImageMimeType('image/png; charset=binary', allowed)).toBe(true);
    expect(isAllowedImageMimeType('image/webp', allowed)).toBe(false);
    expect(isAllowedImageMimeType('image/svg+xml', allowed)).toBe(false);
  });
});

describe('normalizeDownloadFilename', () => {
  it('removes unsafe filename characters and preserves a fallback', () => {
    expect(normalizeDownloadFilename('..\\bad/name?.png', 'image')).toBe('..badname.png');
    expect(normalizeDownloadFilename('', 'image')).toBe('image');
  });
});
