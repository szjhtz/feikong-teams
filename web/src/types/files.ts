export interface FileEntry {
  name: string;
  path: string;
  is_dir?: boolean;
  size?: number;
  mod_time?: string;
}

export interface PreviewLink {
  id?: string;
  link_id?: string;
  path?: string;
  url?: string;
  created_at?: string;
}
