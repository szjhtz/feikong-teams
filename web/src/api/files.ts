import type { FileEntry, PreviewLink } from "@/types/files";
import { del, get, post, request } from "./client";

export function listFiles(path = "") {
  const query = path ? `?path=${encodeURIComponent(path)}` : "";
  return get<FileEntry[]>(`/api/fkteams/files${query}`);
}

export function searchFiles(q: string) {
  return get<FileEntry[]>(`/api/fkteams/files/search?q=${encodeURIComponent(q)}`);
}

export function deleteFile(path: string) {
  return del<{ path: string }>(`/api/fkteams/files?path=${encodeURIComponent(path)}`);
}

export function uploadFile(file: File, path = "") {
  const form = new FormData();
  form.append("file", file);
  if (path) form.append("path", path);
  return request<{ path: string }>("/api/fkteams/files/upload", { method: "POST", body: form });
}

export function createPreviewLink(path: string) {
  return post<PreviewLink>("/api/fkteams/preview", { path });
}

export function listPreviewLinks() {
  return get<{ links?: PreviewLink[] } | PreviewLink[]>("/api/fkteams/preview");
}

export function deletePreviewLink(linkID: string) {
  return del<{ link_id: string }>(`/api/fkteams/preview/${encodeURIComponent(linkID)}`);
}
