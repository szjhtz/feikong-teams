export const authExpiredEvent = "fkteams:auth-expired";
export const authRestoredEvent = "fkteams:auth-restored";

export function expireAuthentication() {
  window.dispatchEvent(new CustomEvent(authExpiredEvent));
}

export function restoreAuthentication() {
  window.dispatchEvent(new CustomEvent(authRestoredEvent));
}

export async function clearServerAuthentication() {
  await fetch("/api/fkteams/logout", {
    method: "POST",
    credentials: "same-origin",
  }).catch(() => undefined);
}
