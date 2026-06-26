export interface SkillInfo {
  name?: string;
  slug: string;
  description?: string;
  description_zh?: string;
  owner?: string;
  homepage?: string;
  version?: string;
  downloads?: number;
  stars?: number;
}

export interface SkillFileEntry {
  name: string;
  path: string;
  is_dir?: boolean;
  size?: number;
}
