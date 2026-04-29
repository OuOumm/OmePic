import type { UploadHistoryRecord } from "@/types/upload";

const DATABASE_NAME = "omepic";
const DATABASE_VERSION = 1;
const STORE_NAME = "uploads";

function openDatabase() {
  return new Promise<IDBDatabase>((resolve, reject) => {
    const request = window.indexedDB.open(DATABASE_NAME, DATABASE_VERSION);
    request.onerror = () => reject(request.error);
    request.onupgradeneeded = () => {
      const database = request.result;
      if (!database.objectStoreNames.contains(STORE_NAME)) {
        database.createObjectStore(STORE_NAME, { keyPath: "uid" });
      }
    };
    request.onsuccess = () => resolve(request.result);
  });
}

async function withStore<T>(mode: IDBTransactionMode, handler: (store: IDBObjectStore) => void | Promise<T>) {
  const database = await openDatabase();
  return new Promise<T>((resolve, reject) => {
    const transaction = database.transaction(STORE_NAME, mode);
    const store = transaction.objectStore(STORE_NAME);
    Promise.resolve(handler(store))
      .then((value) => {
        transaction.oncomplete = () => resolve(value as T);
      })
      .catch(reject);
    transaction.onerror = () => reject(transaction.error);
  });
}

export async function saveUploadRecord(record: UploadHistoryRecord) {
  await withStore<void>("readwrite", (store) => {
    store.put(record);
  });
}

export async function listUploadRecords() {
  return withStore<UploadHistoryRecord[]>("readonly", (store) => {
    return new Promise<UploadHistoryRecord[]>((resolve, reject) => {
      const request = store.getAll();
      request.onerror = () => reject(request.error);
      request.onsuccess = () => {
        const items = (request.result as UploadHistoryRecord[]).sort((a, b) =>
          b.created_at.localeCompare(a.created_at)
        );
        resolve(items);
      };
    });
  });
}

export async function listRecentUploadRecords(limit: number) {
  const items = await listUploadRecords();
  return items.slice(0, limit);
}

export async function removeUploadRecord(uid: string) {
  await withStore<void>("readwrite", (store) => {
    store.delete(uid);
  });
}

export async function clearUploadRecords() {
  await withStore<void>("readwrite", (store) => {
    store.clear();
  });
}
