import { describe, expect, it } from 'vitest';
import type { UploadHistoryRecord } from '@/types';
import { buildUploadHistoryPage, selectedUploadUidsOnPage, sortUploadsByNewestUploadTime } from './upload-history';

function record(overrides: Partial<UploadHistoryRecord>): UploadHistoryRecord {
  return {
    uid: 'uid',
    url: '/i/uid.avif',
    mime_type: 'image/avif',
    size: 1024,
    created_at: '2026-05-10T00:00:00.000Z',
    is_duplicate: false,
    storage_key: 'local',
    storage_backend: 'local',
    markdown: '![uid](/i/uid.avif)',
    bbcode: '[img]/i/uid.avif[/img]',
    client_token: 'token',
    original_filename: 'uid.png',
    saved_at: '2026-05-10T00:00:01.000Z',
    ...overrides,
  };
}

describe('sortUploadsByNewestUploadTime', () => {
  it('sorts uploads by created_at descending instead of uid order', () => {
    const records = [
      record({ uid: 'z-oldest', created_at: '2026-05-09T10:00:00.000Z' }),
      record({ uid: 'a-newest', created_at: '2026-05-11T10:00:00.000Z' }),
      record({ uid: 'm-middle', created_at: '2026-05-10T10:00:00.000Z' }),
    ];

    expect(sortUploadsByNewestUploadTime(records).map((item) => item.uid)).toEqual(['a-newest', 'm-middle', 'z-oldest']);
  });

  it('falls back to saved_at for missing or invalid upload timestamps and keeps ties stable', () => {
    const records = [
      record({ uid: 'first-tie', created_at: '2026-05-11T10:00:00.000Z' }),
      record({ uid: 'fallback-newest', created_at: '', saved_at: '2026-05-12T10:00:00.000Z' }),
      record({ uid: 'invalid-fallback', created_at: 'not-a-date', saved_at: '2026-05-10T10:00:00.000Z' }),
      record({ uid: 'second-tie', created_at: '2026-05-11T10:00:00.000Z' }),
    ];

    expect(sortUploadsByNewestUploadTime(records).map((item) => item.uid)).toEqual([
      'fallback-newest',
      'first-tie',
      'second-tie',
      'invalid-fallback',
    ]);
  });
});

describe('buildUploadHistoryPage', () => {
  const records = [
    record({ uid: 'newest-cat', original_filename: 'Cat Sketch.png', created_at: '2026-05-12T10:00:00.000Z' }),
    record({ uid: 'middle-dog', original_filename: 'Dog Poster.jpg', created_at: '2026-05-11T10:00:00.000Z' }),
    record({ uid: 'oldest-bird', original_filename: 'Bird Notes.webp', created_at: '2026-05-10T10:00:00.000Z' }),
  ];

  it('filters by filename or uid with trimmed case-insensitive query', () => {
    expect(buildUploadHistoryPage(records, { query: '  cat  ', page: 1, pageSize: 20 }).records.map((item) => item.uid)).toEqual(['newest-cat']);
    expect(buildUploadHistoryPage(records, { query: 'DOG', page: 1, pageSize: 20 }).records.map((item) => item.uid)).toEqual(['middle-dog']);
  });

  it('returns paginated records with totals and corrected page bounds', () => {
    const secondPage = buildUploadHistoryPage(records, { query: '', page: 2, pageSize: 1 });
    expect(secondPage.records.map((item) => item.uid)).toEqual(['middle-dog']);
    expect(secondPage).toMatchObject({ page: 2, pageSize: 1, total: 3, filteredTotal: 3, totalPages: 3 });

    const overflowPage = buildUploadHistoryPage(records, { query: '', page: 99, pageSize: 2 });
    expect(overflowPage.records.map((item) => item.uid)).toEqual(['oldest-bird']);
    expect(overflowPage).toMatchObject({ page: 2, pageSize: 2, total: 3, filteredTotal: 3, totalPages: 2 });
  });

  it('handles empty search results without producing invalid pages', () => {
    const page = buildUploadHistoryPage(records, { query: 'missing', page: 3, pageSize: 20 });

    expect(page.records).toEqual([]);
    expect(page).toMatchObject({ page: 1, pageSize: 20, total: 3, filteredTotal: 0, totalPages: 1 });
  });
});

describe('selectedUploadUidsOnPage', () => {
  it('keeps only selected uids that are visible on the current page', () => {
    const records = [
      record({ uid: 'visible-a' }),
      record({ uid: 'visible-b' }),
    ];
    const selected = new Set(['visible-b', 'hidden-c']);

    expect(selectedUploadUidsOnPage(records, selected)).toEqual(['visible-b']);
  });

  it('supports array selection state for selecting every visible upload', () => {
    const records = [
      record({ uid: 'visible-a', client_token: 'old-token' }),
      record({ uid: 'visible-b', client_token: 'current-token' }),
    ];

    expect(selectedUploadUidsOnPage(records, records.map((item) => item.uid))).toEqual(['visible-a', 'visible-b']);
  });
});
