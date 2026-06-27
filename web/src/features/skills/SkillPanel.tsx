import {
  Box,
  Download,
  ExternalLink,
  FileText,
  Folder,
  PackageCheck,
  RefreshCcw,
  Search,
  Sparkles,
  Star,
  Trash2,
} from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import { installSkill, listSkillFiles, listSkills, readSkillFile, removeSkill, searchSkills } from "@/api/skills";
import { skillsActions, appActions } from "@/app/store";
import { useAppDispatch, useAppSelector } from "@/app/hooks";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Panel, PanelBody, PanelHeader } from "@/components/ui/panel";
import { cn } from "@/lib/cn";
import type { SkillFileEntry, SkillInfo } from "@/types/skills";

type SkillView = "installed" | "market";

export function SkillPanel() {
  const dispatch = useAppDispatch();
  const local = useAppSelector((state) => state.skills.local);
  const results = useAppSelector((state) => state.skills.results);
  const [keyword, setKeyword] = useState("");
  const [view, setView] = useState<SkillView>("installed");
  const [selectedSlug, setSelectedSlug] = useState("");
  const [filePath, setFilePath] = useState("");
  const [files, setFiles] = useState<SkillFileEntry[]>([]);
  const [content, setContent] = useState("");
  const [activeFile, setActiveFile] = useState("");
  const [loadingLocal, setLoadingLocal] = useState(false);
  const [searching, setSearching] = useState(false);
  const [busySlug, setBusySlug] = useState("");
  const installedSlugs = useMemo(() => new Set(local.map((skill) => skill.slug)), [local]);
  const selectedSkill =
    local.find((skill) => skill.slug === selectedSlug) || results.find((skill) => skill.slug === selectedSlug);

  async function loadLocal() {
    setLoadingLocal(true);
    try {
      const result = await listSkills();
      dispatch(skillsActions.setLocalSkills(result.skills || []));
    } catch (error) {
      dispatch(appActions.showToast(error instanceof Error ? error.message : String(error)));
    } finally {
      setLoadingLocal(false);
    }
  }

  async function search() {
    const query = keyword.trim();
    if (!query) return;
    setSearching(true);
    setView("market");
    try {
      const result = await searchSkills(query);
      dispatch(skillsActions.setSkillResults(result.skills || []));
    } catch (error) {
      dispatch(appActions.showToast(error instanceof Error ? error.message : String(error)));
    } finally {
      setSearching(false);
    }
  }

  async function select(skill: SkillInfo) {
    setSelectedSlug(skill.slug);
    setFilePath("");
    setActiveFile("");
    setContent("");
    if (!installedSlugs.has(skill.slug)) {
      setFiles([]);
      return;
    }
    await openDirectory(skill.slug, "");
  }

  async function openDirectory(slug = selectedSlug, path = "") {
    if (!slug) return;
    setFilePath(path);
    setActiveFile("");
    setContent("");
    try {
      const result = await listSkillFiles(slug, path);
      setFiles(result.files || []);
    } catch (error) {
      dispatch(appActions.showToast(error instanceof Error ? error.message : String(error)));
    }
  }

  async function openFile(path: string) {
    if (!selectedSlug) return;
    setActiveFile(path);
    try {
      const result = await readSkillFile(selectedSlug, path);
      setContent(result.content || "");
    } catch (error) {
      dispatch(appActions.showToast(error instanceof Error ? error.message : String(error)));
    }
  }

  async function install(slug: string) {
    setBusySlug(slug);
    try {
      await installSkill(slug);
      await loadLocal();
      dispatch(appActions.showToast("技能已安装"));
      setView("installed");
      const skill = results.find((item) => item.slug === slug) || { slug };
      await select(skill);
    } catch (error) {
      dispatch(appActions.showToast(error instanceof Error ? error.message : String(error)));
    } finally {
      setBusySlug("");
    }
  }

  async function remove(slug: string) {
    if (!window.confirm("确定删除这个技能吗？")) return;
    setBusySlug(slug);
    try {
      await removeSkill(slug);
      if (selectedSlug === slug) {
        setSelectedSlug("");
        setFiles([]);
        setContent("");
      }
      await loadLocal();
      dispatch(appActions.showToast("技能已删除"));
    } catch (error) {
      dispatch(appActions.showToast(error instanceof Error ? error.message : String(error)));
    } finally {
      setBusySlug("");
    }
  }

  useEffect(() => {
    void loadLocal();
  }, []);

  return (
    <div className="chat-scroll h-full overflow-auto p-6">
      <div className="mx-auto flex max-w-7xl flex-col gap-4">
        <Panel>
          <PanelHeader className="flex flex-col gap-4 xl:flex-row xl:items-center xl:justify-between">
            <div className="min-w-0">
              <div className="flex items-center gap-3">
                <Sparkles className="h-5 w-5 text-primary" />
                <h2 className="text-xl font-semibold">技能</h2>
              </div>
              <div className="mt-1 text-sm text-muted-foreground">管理本地技能，搜索市场技能，并直接查看技能文件内容。</div>
            </div>
            <div className="grid w-full min-w-0 grid-cols-[minmax(0,1fr)_auto_auto] gap-2 xl:w-[560px]">
              <Input
                className="min-w-0"
                value={keyword}
                onChange={(event) => setKeyword(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === "Enter") void search();
                }}
                placeholder="搜索技能市场"
              />
              <Button className="min-w-20 whitespace-nowrap" onClick={() => void search()} disabled={searching || !keyword.trim()}>
                <Search className="h-4 w-4" />
                搜索
              </Button>
              <Button className="min-w-20 whitespace-nowrap" variant="outline" onClick={() => void loadLocal()} disabled={loadingLocal}>
                <RefreshCcw className="h-4 w-4" />
                刷新
              </Button>
            </div>
          </PanelHeader>
          <PanelBody className="grid gap-3 border-t border-border/70 md:grid-cols-3">
            <MetricCard icon={PackageCheck} label="已安装" value={local.length} />
            <MetricCard icon={Search} label="搜索结果" value={results.length} />
            <MetricCard icon={Box} label="当前选择" value={selectedSkill ? 1 : 0} detail={selectedSkill?.name || selectedSkill?.slug || "未选择"} />
          </PanelBody>
        </Panel>

        <Panel>
          <PanelHeader className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
            <div className="flex gap-2">
              <ViewButton active={view === "installed"} onClick={() => setView("installed")} label="已安装" count={local.length} />
              <ViewButton active={view === "market"} onClick={() => setView("market")} label="技能市场" count={results.length} />
            </div>
            <div className="text-sm text-muted-foreground">
              {view === "installed" ? "点击技能卡片查看文件与说明" : "搜索后可直接安装，已安装技能会标记状态"}
            </div>
          </PanelHeader>
          <PanelBody>
            {view === "installed" ? (
              <SkillGrid
                emptyTitle={loadingLocal ? "正在加载本地技能" : "暂无本地技能"}
                emptyDescription="从技能市场搜索并安装后会显示在这里。"
                skills={local}
                selectedSlug={selectedSlug}
                installedSlugs={installedSlugs}
                busySlug={busySlug}
                onSelect={select}
                onInstall={install}
                onRemove={remove}
              />
            ) : (
              <SkillGrid
                emptyTitle={searching ? "正在搜索技能" : "暂无搜索结果"}
                emptyDescription="输入关键词后搜索技能市场。"
                skills={results}
                selectedSlug={selectedSlug}
                installedSlugs={installedSlugs}
                busySlug={busySlug}
                onSelect={select}
                onInstall={install}
                onRemove={remove}
              />
            )}
          </PanelBody>
        </Panel>

        <SkillDetail
          skill={selectedSkill}
          installed={Boolean(selectedSkill && installedSlugs.has(selectedSkill.slug))}
          files={files}
          filePath={filePath}
          activeFile={activeFile}
          content={content}
          busy={Boolean(selectedSkill && busySlug === selectedSkill.slug)}
          onInstall={(slug) => void install(slug)}
          onRemove={(slug) => void remove(slug)}
          onOpenDirectory={(path) => void openDirectory(selectedSkill?.slug, path)}
          onOpenFile={(path) => void openFile(path)}
        />
      </div>
    </div>
  );
}

function SkillGrid({
  skills,
  selectedSlug,
  installedSlugs,
  busySlug,
  emptyTitle,
  emptyDescription,
  onSelect,
  onInstall,
  onRemove,
}: {
  skills: SkillInfo[];
  selectedSlug: string;
  installedSlugs: Set<string>;
  busySlug: string;
  emptyTitle: string;
  emptyDescription: string;
  onSelect: (skill: SkillInfo) => Promise<void>;
  onInstall: (slug: string) => Promise<void>;
  onRemove: (slug: string) => Promise<void>;
}) {
  if (!skills.length) {
    return <EmptyState title={emptyTitle} description={emptyDescription} />;
  }

  return (
    <div className="grid gap-3 md:grid-cols-2 2xl:grid-cols-3">
      {skills.map((skill) => {
        const installed = installedSlugs.has(skill.slug);
        const selected = selectedSlug === skill.slug;
        return (
          <button
            key={skill.slug}
            className={cn(
              "group flex min-h-44 flex-col rounded-xl border bg-card/65 p-4 text-left transition-[background,border-color,box-shadow]",
              selected ? "border-primary/60 bg-primary/5 shadow-[2px_3px_0_hsl(214_45%_30%/0.12)]" : "border-border/75 hover:bg-accent/45",
            )}
            onClick={() => void onSelect(skill)}
          >
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0">
                <div className="truncate text-base font-semibold">{skill.name || skill.slug}</div>
                <div className="mt-1 truncate text-xs text-muted-foreground">{skill.slug}</div>
              </div>
              <div className="flex shrink-0 flex-wrap justify-end gap-1">
                {installed ? <Badge>已安装</Badge> : null}
                {skill.version ? <Badge>{skill.version}</Badge> : null}
              </div>
            </div>
            <div className="mt-3 line-clamp-3 flex-1 text-sm leading-6 text-muted-foreground">
              {skill.description_zh || skill.description || "暂无描述"}
            </div>
            <div className="mt-4 flex items-center justify-between gap-3">
              <SkillMeta skill={skill} />
              <div className="flex shrink-0 gap-1 opacity-100 md:opacity-0 md:transition-opacity md:group-hover:opacity-100">
                {installed ? (
                  <Button
                    className="whitespace-nowrap"
                    size="sm"
                    variant="ghost"
                    disabled={busySlug === skill.slug}
                    onClick={(event) => {
                      event.stopPropagation();
                      void onRemove(skill.slug);
                    }}
                  >
                    <Trash2 className="h-4 w-4" />
                    删除
                  </Button>
                ) : (
                  <Button
                    className="whitespace-nowrap"
                    size="sm"
                    disabled={busySlug === skill.slug}
                    onClick={(event) => {
                      event.stopPropagation();
                      void onInstall(skill.slug);
                    }}
                  >
                    <Download className="h-4 w-4" />
                    安装
                  </Button>
                )}
              </div>
            </div>
          </button>
        );
      })}
    </div>
  );
}

function SkillDetail({
  skill,
  installed,
  files,
  filePath,
  activeFile,
  content,
  busy,
  onInstall,
  onRemove,
  onOpenDirectory,
  onOpenFile,
}: {
  skill?: SkillInfo;
  installed: boolean;
  files: SkillFileEntry[];
  filePath: string;
  activeFile: string;
  content: string;
  busy: boolean;
  onInstall: (slug: string) => void;
  onRemove: (slug: string) => void;
  onOpenDirectory: (path: string) => void;
  onOpenFile: (path: string) => void;
}) {
  if (!skill) {
    return (
      <Panel>
        <PanelBody>
          <EmptyState title="选择一个技能" description="技能详情、文件列表和内容预览会在这里展示。" />
        </PanelBody>
      </Panel>
    );
  }

  return (
    <Panel>
      <PanelHeader className="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <h3 className="text-lg font-semibold">{skill.name || skill.slug}</h3>
            {installed ? <Badge>已安装</Badge> : <Badge>未安装</Badge>}
            {skill.version ? <Badge>{skill.version}</Badge> : null}
          </div>
          <div className="mt-1 text-sm text-muted-foreground">{skill.description_zh || skill.description || "暂无描述"}</div>
          <div className="mt-3 flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
            <span>{skill.slug}</span>
            {skill.owner ? <span>{skill.owner}</span> : null}
            {skill.homepage ? (
              <a className="inline-flex items-center gap-1 hover:text-foreground" href={skill.homepage} target="_blank" rel="noreferrer">
                主页
                <ExternalLink className="h-3.5 w-3.5" />
              </a>
            ) : null}
          </div>
        </div>
        <div className="flex shrink-0 gap-2">
          {installed ? (
            <Button className="min-w-20 whitespace-nowrap" variant="outline" disabled={busy} onClick={() => onRemove(skill.slug)}>
              <Trash2 className="h-4 w-4" />
              删除
            </Button>
          ) : (
            <Button className="min-w-20 whitespace-nowrap" disabled={busy} onClick={() => onInstall(skill.slug)}>
              <Download className="h-4 w-4" />
              安装
            </Button>
          )}
        </div>
      </PanelHeader>
      <PanelBody className="space-y-4 border-t border-border/70">
        {installed ? (
          <>
            <div className="space-y-2">
              <div className="flex flex-wrap items-center gap-2">
                <button
                  className={cn(
                    "rounded-full border px-3 py-1 text-sm",
                    filePath ? "border-border bg-card/70 hover:bg-accent/60" : "border-primary/50 bg-primary/10 text-primary",
                  )}
                  onClick={() => onOpenDirectory("")}
                >
                  根目录
                </button>
                {filePath ? <span className="text-sm text-muted-foreground">/ {filePath}</span> : null}
              </div>
              <div className="flex flex-wrap gap-2">
                {files.map((file) => {
                  const Icon = file.is_dir ? Folder : FileText;
                  return (
                    <button
                      key={file.path}
                      className={cn(
                        "inline-flex h-9 items-center gap-2 rounded-lg border px-3 text-sm transition-colors hover:bg-accent/60",
                        activeFile === file.path ? "border-primary/50 bg-primary/10 text-primary" : "border-border/75 bg-card/70",
                      )}
                      onClick={() => (file.is_dir ? onOpenDirectory(file.path) : onOpenFile(file.path))}
                    >
                      <Icon className="h-4 w-4" />
                      <span>{file.name}</span>
                      {file.size !== undefined && !file.is_dir ? <span className="text-xs text-muted-foreground">{formatSize(file.size)}</span> : null}
                    </button>
                  );
                })}
                {!files.length ? <div className="text-sm text-muted-foreground">暂无文件</div> : null}
              </div>
            </div>
            <div className="rounded-xl border border-border/75 bg-card/65">
              <div className="flex h-11 items-center justify-between border-b border-border/70 px-4">
                <div className="truncate text-sm font-medium">{activeFile || "文件预览"}</div>
                {content ? <Badge>{content.length} 字符</Badge> : null}
              </div>
              <pre className="chat-scroll max-h-[46vh] min-h-64 overflow-auto whitespace-pre-wrap p-5 text-sm leading-7">
                {content || "选择文件查看内容"}
              </pre>
            </div>
          </>
        ) : (
          <EmptyState title="技能尚未安装" description="安装后可查看本地文件和技能说明。" />
        )}
      </PanelBody>
    </Panel>
  );
}

function ViewButton({ active, label, count, onClick }: { active: boolean; label: string; count: number; onClick: () => void }) {
  return (
    <button
      className={cn(
        "inline-flex h-10 items-center gap-2 rounded-lg border px-3 text-sm transition-colors",
        active ? "border-primary/50 bg-primary/10 text-primary" : "border-transparent text-muted-foreground hover:border-border hover:bg-card",
      )}
      onClick={onClick}
    >
      {label}
      <span className="rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground">{count}</span>
    </button>
  );
}

function MetricCard({ icon: Icon, label, value, detail }: { icon: typeof PackageCheck; label: string; value: number; detail?: string }) {
  return (
    <div className="rounded-xl border border-border/75 bg-card/65 p-4">
      <div className="flex items-center justify-between gap-3">
        <div className="text-sm text-muted-foreground">{label}</div>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </div>
      <div className="mt-2 text-3xl font-semibold">{value}</div>
      {detail ? <div className="mt-1 truncate text-xs text-muted-foreground">{detail}</div> : null}
    </div>
  );
}

function SkillMeta({ skill }: { skill: SkillInfo }) {
  return (
    <div className="flex min-w-0 flex-wrap items-center gap-2 text-xs text-muted-foreground">
      {skill.stars !== undefined ? (
        <span className="inline-flex items-center gap-1">
          <Star className="h-3.5 w-3.5" />
          {skill.stars}
        </span>
      ) : null}
      {skill.downloads !== undefined ? (
        <span className="inline-flex items-center gap-1">
          <Download className="h-3.5 w-3.5" />
          {skill.downloads}
        </span>
      ) : null}
      {skill.owner ? <span className="truncate">{skill.owner}</span> : null}
    </div>
  );
}

function EmptyState({ title, description }: { title: string; description: string }) {
  return (
    <div className="rounded-xl border border-dashed border-border p-10 text-center">
      <div className="font-medium">{title}</div>
      <div className="mt-1 text-sm text-muted-foreground">{description}</div>
    </div>
  );
}

function formatSize(value: number) {
  if (value < 1024) return `${value} B`;
  if (value < 1024 * 1024) return `${(value / 1024).toFixed(1)} KB`;
  return `${(value / 1024 / 1024).toFixed(1)} MB`;
}
