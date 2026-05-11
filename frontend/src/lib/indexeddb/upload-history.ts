import type { UploadHistoryRecord } from '@/types';

export type UploadHistoryPageOptions = {
  query?: string;
  page?: number;
  pageSize?: number;
};

export type UploadHistoryPage = {
  records: UploadHistoryRecord[];
  page: number;
  pageSize: number;
  total: number;
  filteredTotal: number;
  totalPages: number;
};

const DB_NAME = 'omepic';
const STORE_NAME = 'uploads';
const DB_VERSION = 1;
const CREATED_AT_INDEX = 'created_at';

function ensureUploadStore(db: IDBDatabase, tx: IDBTransaction | null): void {
  const store = db.objectStoreNames.contains(STORE_NAME)
    ? tx?.objectStore(STORE_NAME)
    : db.createObjectStore(STORE_NAME, { keyPath: 'uid' });

  if (store && !store.indexNames.contains(CREATED_AT_INDEX)) {
    store.createIndex(CREATED_AT_INDEX, 'created_at');
  }
}

function openDB(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, DB_VERSION);
    req.onupgradeneeded = () => {
      ensureUploadStore(req.result, req.transaction);
    };
    req.onsuccess = () => resolve(req.result);
    req.onerror = () => reject(req.error);
  });
}

function uploadTime(record: UploadHistoryRecord): number {
  const createdAt = Date.parse(record.created_at);
  if (Number.isFinite(createdAt)) return createdAt;

  const savedAt = Date.parse(record.saved_at);
  return Number.isFinite(savedAt) ? savedAt : 0;
}

export function sortUploadsByNewestUploadTime(records: readonly UploadHistoryRecord[]): UploadHistoryRecord[] {
  return records
    .map((record, index) => ({ record, index }))
    .sort((left, right) => {
      const timeDiff = uploadTime(right.record) - uploadTime(left.record);
      return timeDiff || left.index - right.index;
    })
    .map(({ record }) => record);
}

function uploadMatchesQuery(record: UploadHistoryRecord, normalizedQuery: string): boolean {
  if (!normalizedQuery) return true;
  return [record.original_filename, record.uid].some((value) => value.toLowerCase().includes(normalizedQuery));
}

export function buildUploadHistoryPage(records: readonly UploadHistoryRecord[], options: UploadHistoryPageOptions = {}): UploadHistoryPage {
  const pageSize = Math.max(1, Math.floor(options.pageSize ?? 20));
  const normalizedQuery = options.query?.trim().toLowerCase() ?? '';
  const filteredRecords = sortUploadsByNewestUploadTime(records).filter((record) => uploadMatchesQuery(record, normalizedQuery));
  const totalPages = Math.max(1, Math.ceil(filteredRecords.length / pageSize));
  const requestedPage = Math.max(1, Math.floor(options.page ?? 1));
  const page = Math.min(requestedPage, totalPages);
  const start = (page - 1) * pageSize;

  return {
    records: filteredRecords.slice(start, start + pageSize),
    page,
    pageSize,
    total: records.length,
    filteredTotal: filteredRecords.length,
    totalPages,
  };
}

function uidIsSelected(selectedUids: ReadonlySet<string> | readonly string[], uid: string): boolean {
  return 'has' in selectedUids ? selectedUids.has(uid) : selectedUids.includes(uid);
}

export function selectedUploadUidsOnPage(records: readonly UploadHistoryRecord[], selectedUids: ReadonlySet<string> | readonly string[]): string[] {
  return records.map((record) => record.uid).filter((uid) => uidIsSelected(selectedUids, uid));
}

export async function saveUploadToHistory(record: UploadHistoryRecord): Promise<void> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, 'readwrite');
    tx.objectStore(STORE_NAME).put(record);
    tx.oncomplete = () => resolve();
    tx.onerror = () => reject(tx.error);
  });
}

export async function getRecentUploads(limit = 10): Promise<UploadHistoryRecord[]> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, 'readonly');
    const records: UploadHistoryRecord[] = [];
    tx.objectStore(STORE_NAME).openCursor().onsuccess = (event) => {
      const cursor = (event.target as IDBRequest<IDBCursorWithValue>).result;
      if (cursor) {
        records.push(cursor.value);
        cursor.continue();
      } else {
        resolve(sortUploadsByNewestUploadTime(records).slice(0, limit));
      }
    };
    tx.onerror = () => reject(tx.error);
  });
}

export async function getAllUploads(): Promise<UploadHistoryRecord[]> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, 'readonly');
    const records: UploadHistoryRecord[] = [];
    tx.objectStore(STORE_NAME).openCursor().onsuccess = (event) => {
      const cursor = (event.target as IDBRequest<IDBCursorWithValue>).result;
      if (cursor) {
        records.push(cursor.value);
        cursor.continue();
      } else {
        resolve(sortUploadsByNewestUploadTime(records));
      }
    };
    tx.onerror = () => reject(tx.error);
  });
}

export async function getUploadHistoryPage(options: UploadHistoryPageOptions = {}): Promise<UploadHistoryPage> {
  const records = await getAllUploads();
  return buildUploadHistoryPage(records, options);
}

export async function deleteUploadFromHistory(uid: string): Promise<void> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, 'readwrite');
    tx.objectStore(STORE_NAME).delete(uid);
    tx.oncomplete = () => resolve();
    tx.onerror = () => reject(tx.error);
  });
}

export async function clearUploadHistory(): Promise<void> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, 'readwrite');
    tx.objectStore(STORE_NAME).clear();
    tx.oncomplete = () => resolve();
    tx.onerror = () => reject(tx.error);
  });
}

export async function getUploadCount(): Promise<number> {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, 'readonly');
    const req = tx.objectStore(STORE_NAME).count();
    req.onsuccess = () => resolve(req.result);
    tx.onerror = () => reject(tx.error);
  });
}
