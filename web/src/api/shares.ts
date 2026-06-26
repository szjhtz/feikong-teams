import type { SessionDetail } from "@/types/chat";
import { del, get, post } from "./client";

export interface SessionShare {
  share_id: string;
  session_id: string;
  title?: string;
  created_at?: string;
  expires_at?: string;
}

export function createSessionShare(sessionID: string, password = "") {
  return post<SessionShare>("/api/fkteams/session-shares", { session_id: sessionID, password });
}

export function listSessionShares() {
  return get<{ shares?: SessionShare[] } | SessionShare[]>("/api/fkteams/session-shares");
}

export function deleteSessionShare(shareID: string) {
  return del<{ share_id: string }>(`/api/fkteams/session-shares/${encodeURIComponent(shareID)}`);
}

export function getPublicShareInfo(shareID: string) {
  return get<SessionShare>(`/api/fkteams/public/session-shares/${encodeURIComponent(shareID)}/info`);
}

export function accessPublicShare(shareID: string, password = "") {
  return post<SessionDetail>(`/api/fkteams/public/session-shares/${encodeURIComponent(shareID)}/access`, { password });
}
