const TOKEN_STORAGE_KEY = 'omepic-client-token';

function generateToken(): string {
  const random = globalThis.crypto;
  if (random?.randomUUID) return random.randomUUID();

  const bytes = new Uint8Array(32);
  if (random?.getRandomValues) {
    random.getRandomValues(bytes);
    return Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('');
  }

  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  let token = '';
  for (let i = 0; i < 32; i += 1) {
    token += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return token;
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
