import type { ModelInfo, ProviderInfo } from "@/types/api";
import { get, post } from "./client";

export function listProviders() {
  return get<{ providers?: ProviderInfo[] } | ProviderInfo[]>("/api/fkteams/providers");
}

export function listProviderModels(request: {
  provider: string;
  base_url?: string;
  api_key?: string;
  model_id?: string;
  original_id?: string;
  extra_headers?: string;
}) {
  return post<ModelInfo[]>("/api/fkteams/providers/models", request);
}
