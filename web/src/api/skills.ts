import type { SkillFileEntry, SkillInfo } from "@/types/skills";
import { del, get, post } from "./client";

export function listSkills() {
  return get<{ skills: SkillInfo[]; total: number }>("/api/fkteams/skills");
}

export function searchSkills(q: string) {
  return get<{ skills: SkillInfo[]; total: number }>(`/api/fkteams/skills/search?q=${encodeURIComponent(q)}`);
}

export function installSkill(slug: string) {
  return post<{ slug: string }>("/api/fkteams/skills/install", { slug });
}

export function removeSkill(slug: string) {
  return del<{ slug: string }>(`/api/fkteams/skills/${encodeURIComponent(slug)}`);
}

export function listSkillFiles(slug: string, path = "") {
  const query = path ? `?path=${encodeURIComponent(path)}` : "";
  return get<{ slug: string; files: SkillFileEntry[] }>(
    `/api/fkteams/skills/${encodeURIComponent(slug)}/files${query}`,
  );
}

export function readSkillFile(slug: string, path: string) {
  return get<{ slug: string; path: string; content: string }>(
    `/api/fkteams/skills/${encodeURIComponent(slug)}/file?path=${encodeURIComponent(path)}`,
  );
}
