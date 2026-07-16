export function clearLegacyAuthStorage() {
  const legacyToken = localStorage.getItem("fk_token");
  if (!legacyToken) return;
  localStorage.removeItem("fk_token");
  document.cookie = "fk_token=; Path=/; Max-Age=0; SameSite=Lax";
}
