import { post } from "./client";

export interface StartStreamRequest {
  session_id?: string;
  message?: string;
  mode?: string;
  agent_name?: string;
  contents?: unknown[];
}

export interface StartStreamResponse {
  session_id: string;
  run_id?: string;
  status?: "processing" | "queued" | string;
  queue_kind?: string;
  queued_count?: number;
}

export function startStream(req: StartStreamRequest) {
  return post<StartStreamResponse>("/api/fkteams/stream/start", req);
}

export function stopStream(sessionID: string) {
  return post<{ session_id: string }>(`/api/fkteams/stream/stop/${encodeURIComponent(sessionID)}`);
}

export function sendSteering(sessionID: string, message: string) {
  return post<{ session_id: string }>("/api/fkteams/stream/steer", { session_id: sessionID, message });
}
