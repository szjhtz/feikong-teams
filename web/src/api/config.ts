import type { AppConfig, ToolInfo } from "@/types/config";
import { get, put } from "./client";

export function getConfig() {
  return get<AppConfig>("/api/fkteams/config");
}

export function saveConfig(config: AppConfig) {
  return put<AppConfig>("/api/fkteams/config", config);
}

export function getToolCatalog() {
  return get<ToolInfo[]>("/api/fkteams/config/tool-catalog");
}

export function getTemplateVars() {
  return get<Record<string, string>>("/api/fkteams/config/template-vars");
}
