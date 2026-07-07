import { post } from "@/api/client";
import type { AgentConfig } from "@/types/config";
import type { SkillDraft } from "@/types/skills";

export interface AgentDraftRequest {
  instruction: string;
  existing_agents?: string[];
  available_tools?: string[];
  available_models?: string[];
  default_model_id?: string;
}

export interface AgentDraftResponse {
  agents: AgentConfig[];
}

export interface RewriteTextRequest {
  scenario: string;
  instruction: string;
  text: string;
  context?: Record<string, unknown>;
}

export interface RewriteTextResponse {
  text: string;
}

export interface SkillDraftRequest {
  instruction: string;
  existing_skills?: string[];
}

export interface SkillDraftResponse {
  skill: SkillDraft;
}

export function generateAgentDrafts(body: AgentDraftRequest) {
  return post<AgentDraftResponse>("/api/fkteams/ai/agents/draft", body);
}

export function rewriteText(body: RewriteTextRequest) {
  return post<RewriteTextResponse>("/api/fkteams/ai/text/rewrite", body);
}

export function generateSkillDraft(body: SkillDraftRequest) {
  return post<SkillDraftResponse>("/api/fkteams/ai/skills/draft", body);
}
