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
  display_name?: string;
  description?: string;
  role?: string;
  aliases?: string[];
  builtin?: boolean;
  team_member?: boolean;
  prompt?: string;
  model_id?: string;
  tools?: string[];
}

export interface ProviderInfo {
  name: string;
  display_name?: string;
  base_url?: string;
  models?: string[];
}

export interface ModelInfo {
  id: string;
}
