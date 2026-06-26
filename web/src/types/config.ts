export interface ModelConfig {
  name: string;
  provider: string;
  model: string;
  api_key?: string;
  base_url?: string;
}

export interface ToolInfo {
  name: string;
  display_name?: string;
  description?: string;
  category?: string;
  builtin?: boolean;
  read_only?: boolean;
  destructive?: boolean;
  included_tools?: string[];
}

export interface AppConfig {
  models?: ModelConfig[];
  server?: Record<string, unknown>;
  agents?: Record<string, unknown>;
  custom?: Record<string, unknown>;
  channels?: Record<string, unknown>;
  memory?: Record<string, unknown>;
  [key: string]: unknown;
}
