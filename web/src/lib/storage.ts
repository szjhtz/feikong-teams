export const storageKeys = {
  token: "fk_token",
  sessionID: "fk_session_id",
  sidebarCollapsed: "fk_sidebar_collapsed",
} as const;

export function readJSON<T>(key: string, fallback: T): T {
  try {
    const value = localStorage.getItem(key);
    return value ? (JSON.parse(value) as T) : fallback;
  } catch {
    return fallback;
  }
}

export function writeJSON(key: string, value: unknown) {
  localStorage.setItem(key, JSON.stringify(value));
}

export function authToken() {
  return localStorage.getItem(storageKeys.token) || "";
}

export function setAuthToken(token: string) {
  if (token) {
    localStorage.setItem(storageKeys.token, token);
    document.cookie = `fk_token=${encodeURIComponent(token)}; path=/; max-age=2592000; SameSite=Lax`;
    return;
  }
  localStorage.removeItem(storageKeys.token);
  document.cookie = "fk_token=; path=/; max-age=0";
}
