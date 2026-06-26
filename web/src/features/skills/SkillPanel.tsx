import { Download, FileText, RefreshCcw, Search, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";
import { installSkill, listSkillFiles, listSkills, readSkillFile, removeSkill, searchSkills } from "@/api/skills";
import { skillsActions, appActions } from "@/app/store";
import { useAppDispatch, useAppSelector } from "@/app/hooks";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Panel, PanelBody, PanelHeader } from "@/components/ui/panel";
import type { SkillFileEntry } from "@/types/skills";

export function SkillPanel() {
  const dispatch = useAppDispatch();
  const local = useAppSelector((state) => state.skills.local);
  const results = useAppSelector((state) => state.skills.results);
  const [keyword, setKeyword] = useState("");
  const [selected, setSelected] = useState("");
  const [files, setFiles] = useState<SkillFileEntry[]>([]);
  const [content, setContent] = useState("");

  async function loadLocal() {
    const result = await listSkills();
    dispatch(skillsActions.setLocalSkills(result.skills || []));
  }

  async function search() {
    if (!keyword.trim()) return;
    const result = await searchSkills(keyword.trim());
    dispatch(skillsActions.setSkillResults(result.skills || []));
  }

  async function select(slug: string) {
    setSelected(slug);
    setContent("");
    const result = await listSkillFiles(slug);
    setFiles(result.files || []);
  }

  async function openFile(path: string) {
    if (!selected) return;
    const result = await readSkillFile(selected, path);
    setContent(result.content || "");
  }

  useEffect(() => {
    void loadLocal();
  }, []);

  return (
    <div className="grid h-full grid-cols-[380px_1fr] overflow-hidden">
      <div className="space-y-4 overflow-auto border-r p-4">
        <Panel>
          <PanelHeader className="flex items-center justify-between">
            <div className="font-semibold">本地技能</div>
            <Button size="icon" variant="ghost" onClick={loadLocal} aria-label="刷新">
              <RefreshCcw className="h-4 w-4" />
            </Button>
          </PanelHeader>
          <PanelBody className="space-y-2">
            {local.map((skill) => (
              <button key={skill.slug} className="w-full rounded-md border p-3 text-left hover:bg-accent" onClick={() => select(skill.slug)}>
                <div className="font-medium">{skill.name || skill.slug}</div>
                <div className="mt-1 line-clamp-2 text-xs text-muted-foreground">{skill.description || skill.description_zh}</div>
              </button>
            ))}
          </PanelBody>
        </Panel>
        <Panel>
          <PanelHeader>
            <div className="font-semibold">技能市场</div>
          </PanelHeader>
          <PanelBody className="space-y-3">
            <div className="flex gap-2">
              <Input value={keyword} onChange={(event) => setKeyword(event.target.value)} placeholder="搜索技能" />
              <Button onClick={search}>
                <Search className="h-4 w-4" />
              </Button>
            </div>
            {results.map((skill) => (
              <div key={skill.slug} className="rounded-md border p-3">
                <div className="flex items-center justify-between gap-2">
                  <div className="font-medium">{skill.name || skill.slug}</div>
                  <Badge>{skill.version || "latest"}</Badge>
                </div>
                <div className="mt-1 text-xs text-muted-foreground">{skill.description_zh || skill.description}</div>
                <Button
                  className="mt-2"
                  size="sm"
                  onClick={() => installSkill(skill.slug).then(loadLocal).then(() => dispatch(appActions.showToast("技能已安装")))}
                >
                  <Download className="h-4 w-4" />
                  安装
                </Button>
              </div>
            ))}
          </PanelBody>
        </Panel>
      </div>
      <div className="grid min-w-0 grid-cols-[320px_1fr] overflow-hidden">
        <div className="overflow-auto border-r p-4">
          <div className="mb-3 flex items-center justify-between">
            <div className="font-semibold">{selected || "选择技能"}</div>
            {selected ? (
              <Button variant="ghost" size="icon" onClick={() => removeSkill(selected).then(loadLocal)} aria-label="删除技能">
                <Trash2 className="h-4 w-4" />
              </Button>
            ) : null}
          </div>
          {files.map((file) => (
            <button key={file.path} className="mb-1 flex w-full items-center gap-2 rounded-md px-2 py-2 text-left text-sm hover:bg-accent" onClick={() => !file.is_dir && openFile(file.path)}>
              <FileText className="h-4 w-4" />
              {file.name}
            </button>
          ))}
        </div>
        <pre className="overflow-auto p-5 text-sm">{content || "选择文件查看内容"}</pre>
      </div>
    </div>
  );
}
