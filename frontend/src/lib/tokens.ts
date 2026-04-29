const USER_TOKEN_KEY = "omepic:user-token";

export function ensureClientToken() {
  const existing = window.localStorage.getItem(USER_TOKEN_KEY);
  if (existing) {
    return existing;
  }
  const token = crypto.randomUUID();
  window.localStorage.setItem(USER_TOKEN_KEY, token);
  return token;
}

export function getClientToken() {
  return window.localStorage.getItem(USER_TOKEN_KEY);
}
