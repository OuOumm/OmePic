const TOKEN_STORAGE_KEY = 'omepic-client-token';

function generateToken(): string {
  const crypto = globalThis.crypto;
  if (!crypto) {
    throw new Error('Web Crypto API is required. Please use a modern browser.');
  }
  if (crypto.randomUUID) return crypto.randomUUID();
  const bytes = new Uint8Array(32);
  crypto.getRandomValues(bytes);
  return Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('');
}

export function getClientToken(): string {
  if (typeof window === 'undefined') return '';
  let token = localStorage.getItem(TOKEN_STORAGE_KEY);
  if (!token) {
    token = generateToken();
    localStorage.setItem(TOKEN_STORAGE_KEY, token);
  }
  return token;
}
