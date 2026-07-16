export const storageKeys = {
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

export function clearLegacyAuthStorage() {
  const legacyToken = localStorage.getItem("fk_token");
  if (!legacyToken) return;
  localStorage.removeItem("fk_token");
  document.cookie = "fk_token=; Path=/; Max-Age=0; SameSite=Lax";
}
