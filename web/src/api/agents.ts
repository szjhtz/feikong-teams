import type { AgentInfo } from "@/types/api";
import { get } from "./client";

export function listAgents() {
  return get<{ agents?: AgentInfo[] } | AgentInfo[]>("/api/fkteams/agents");
}
