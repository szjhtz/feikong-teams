export interface APIResponse<T> {
  code: number;
  message?: string;
  data: T;
}

export interface VersionInfo {
  version: string;
  commit?: string;
  build_time?: string;
}

export interface AgentInfo {
  name: string;
  description?: string;
  role?: string;
}

export interface ProviderInfo {
  name: string;
  display_name?: string;
  base_url?: string;
  models?: string[];
}
