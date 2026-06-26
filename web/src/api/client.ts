import { authToken, setAuthToken } from "@/lib/storage";
import type { APIResponse } from "@/types/api";

export class APIError extends Error {
  status: number;

  constructor(message: string, status: number) {
    super(message);
    this.name = "APIError";
    this.status = status;
  }
}

export async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  const token = authToken();
  if (token) headers.set("Authorization", `Bearer ${token}`);
  if (init.body && !headers.has("Content-Type") && !(init.body instanceof FormData)) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(path, { ...init, headers });
  if (response.status === 401) {
    setAuthToken("");
    window.dispatchEvent(new CustomEvent("fkteams:auth-expired"));
    throw new APIError("unauthorized", 401);
  }
  if (!response.ok) {
    throw new APIError(response.statusText || "request failed", response.status);
  }

  const payload = (await response.json()) as APIResponse<T>;
  if (payload.code !== 0) {
    throw new APIError(payload.message || "request failed", response.status);
  }
  return payload.data;
}

export function get<T>(path: string) {
  return request<T>(path);
}

export function post<T>(path: string, body?: unknown) {
  return request<T>(path, {
    method: "POST",
    body: body === undefined ? undefined : JSON.stringify(body),
  });
}

export function put<T>(path: string, body?: unknown) {
  return request<T>(path, {
    method: "PUT",
    body: body === undefined ? undefined : JSON.stringify(body),
  });
}

export function patch<T>(path: string, body?: unknown) {
  return request<T>(path, {
    method: "PATCH",
    body: body === undefined ? undefined : JSON.stringify(body),
  });
}

export function del<T>(path: string) {
  return request<T>(path, { method: "DELETE" });
}
