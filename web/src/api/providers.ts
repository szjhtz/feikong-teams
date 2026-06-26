import type { ProviderInfo } from "@/types/api";
import { get, post } from "./client";

export function listProviders() {
  return get<{ providers?: ProviderInfo[] } | ProviderInfo[]>("/api/fkteams/providers");
}

export function listProviderModels(provider: string, base_url?: string, api_key?: string) {
  return post<{ models?: string[] }>("/api/fkteams/providers/models", { provider, base_url, api_key });
}
