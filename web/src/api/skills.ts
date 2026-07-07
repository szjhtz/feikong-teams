import type { SkillCreateRequest, SkillFileEntry, SkillInfo } from "@/types/skills";
import { del, get, post, put } from "./client";

export function listSkills() {
  return get<{ skills: SkillInfo[]; total: number }>("/api/fkteams/skills");
}

export function searchSkills(q: string) {
  return get<{ skills: SkillInfo[]; total: number }>(`/api/fkteams/skills/search?q=${encodeURIComponent(q)}`);
}

export function installSkill(slug: string) {
  return post<{ slug: string }>("/api/fkteams/skills/install", { slug });
}

export function createSkill(body: SkillCreateRequest) {
  return post<{ skill: SkillInfo }>("/api/fkteams/skills", body);
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

export function saveSkillFile(slug: string, path: string, content: string) {
  return put<{ slug: string; path: string }>(`/api/fkteams/skills/${encodeURIComponent(slug)}/file`, { path, content });
}

export function createSkillFile(slug: string, path: string, content = "", isDir = false) {
  return post<{ slug: string; path: string }>(`/api/fkteams/skills/${encodeURIComponent(slug)}/files`, {
    path,
    content,
    is_dir: isDir,
  });
}

export function deleteSkillFile(slug: string, path: string) {
  return del<{ slug: string; path: string }>(
    `/api/fkteams/skills/${encodeURIComponent(slug)}/file?path=${encodeURIComponent(path)}`,
  );
}
