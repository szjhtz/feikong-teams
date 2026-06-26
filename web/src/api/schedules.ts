import type { ScheduleHistoryEntry, ScheduleTask } from "@/types/schedules";
import { get, post } from "./client";

export function listSchedules(status = "") {
  const query = status ? `?status=${encodeURIComponent(status)}` : "";
  return get<{ tasks: ScheduleTask[] }>(`/api/fkteams/schedules${query}`);
}

export function cancelSchedule(id: string) {
  return post<{ id: string }>(`/api/fkteams/schedules/${encodeURIComponent(id)}/cancel`);
}

export function getScheduleResult(id: string) {
  return get<{ result?: string; content?: string }>(`/api/fkteams/schedules/${encodeURIComponent(id)}/result`);
}

export function getScheduleHistory(id: string) {
  return get<{ entries?: ScheduleHistoryEntry[]; count?: number }>(
    `/api/fkteams/schedules/${encodeURIComponent(id)}/history`,
  );
}
