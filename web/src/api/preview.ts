import { get, post } from "./client";

export interface PreviewInfo {
  file_name?: string;
  file_size?: number;
  file_count?: number;
  files?: Array<{ path: string; name: string; is_dir: boolean; size?: number }>;
  content_type?: string;
  require_password?: boolean;
  previewable?: boolean;
  authorized?: boolean;
  expires_at?: number;
}

export function getPreviewInfo(linkID: string) {
  return get<PreviewInfo>(`/api/fkteams/preview/${encodeURIComponent(linkID)}/info`, {
    authFailure: "ignore",
  });
}

export function authorizePreview(linkID: string, password: string) {
  return post<{ authenticated: boolean }>(
    `/api/fkteams/preview/${encodeURIComponent(linkID)}/auth`,
    { password },
    { authFailure: "ignore" },
  );
}
