import {
  Archive,
  ArrowLeft,
  Code2,
  Download,
  File,
  FilePenLine,
  FileText,
  Folder,
  Image,
  MoreVertical,
  RefreshCcw,
  Save,
  Share2,
  Trash2,
  Upload,
} from "lucide-react";
import { useEffect, useState } from "react";
import type { LucideIcon } from "lucide-react";
import { deleteFile, listFiles, readFileContent, saveFileContent, uploadFile } from "@/api/files";
import { filesActions, appActions } from "@/app/store";
import { useAppDispatch, useAppSelector } from "@/app/hooks";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Panel, PanelBody, PanelHeader } from "@/components/ui/panel";
import { formatBytes, formatTime } from "@/lib/format";
import { cn } from "@/lib/cn";
import type { FileContent, FileEntry } from "@/types/files";
import { FileShareDialog } from "./FileShareDialog";

export function FileManager() {
  const dispatch = useAppDispatch();
  const path = useAppSelector((state) => state.files.path);
  const entries = useAppSelector((state) => state.files.entries);
  const [uploading, setUploading] = useState(false);
  const [editor, setEditor] = useState<FileContent | null>(null);
  const [draft, setDraft] = useState("");
  const [editorError, setEditorError] = useState("");
  const [saving, setSaving] = useState(false);
  const [shareTarget, setShareTarget] = useState<FileEntry | null>(null);
  const [openActionPath, setOpenActionPath] = useState("");
  const actionEntry = entries.find((entry) => entry.path === openActionPath);

  async function load(nextPath = path) {
    const files = await listFiles(nextPath);
    dispatch(filesActions.setPath(nextPath));
    dispatch(filesActions.setFiles(files || []));
  }

  async function handleUpload(file?: File) {
    if (!file) return;
    setUploading(true);
    try {
      await uploadFile(file, path);
      await load();
    } finally {
      setUploading(false);
    }
  }

  async function editFile(filePath: string) {
    setEditorError("");
    try {
      const file = await readFileContent(filePath);
      setEditor(file);
      setDraft(file.content || "");
    } catch (error) {
      setEditorError(error instanceof Error ? error.message : String(error));
    }
  }

  async function saveEditor() {
    if (!editor) return;
    setSaving(true);
    setEditorError("");
    try {
      const saved = await saveFileContent(editor.path, draft);
      setEditor({ ...editor, ...saved, content: draft });
      dispatch(appActions.showToast("已保存"));
      await load();
    } catch (error) {
      setEditorError(error instanceof Error ? error.message : String(error));
    } finally {
      setSaving(false);
    }
  }

  function openEntry(entry: FileEntry) {
    setOpenActionPath("");
    if (entry.is_dir) {
      setEditor(null);
      void load(entry.path);
      return;
    }
    void editFile(entry.path);
  }

  function downloadEntry(entry: FileEntry) {
    setOpenActionPath("");
    window.open(`/api/fkteams/files/download?path=${encodeURIComponent(entry.path)}`);
  }

  async function removeEntry(entry: FileEntry) {
    setOpenActionPath("");
    await deleteFile(entry.path);
    await load();
    dispatch(appActions.showToast("已删除"));
  }

  function shareEntry(entry: FileEntry) {
    setOpenActionPath("");
    setShareTarget(entry);
  }

  useEffect(() => {
    void load("");
  }, []);

  return (
    <div className={cn("h-full p-3 sm:p-6", editor ? "overflow-hidden" : "overflow-auto")}>
      <Panel className={cn("flex min-h-0 flex-col", editor ? "h-full w-full" : "mx-auto max-w-6xl")}>
        <PanelHeader className="flex flex-wrap items-center justify-between gap-4">
          {editor ? (
            <>
              <div className="min-w-0">
                <div className="flex min-w-0 items-center gap-2 font-semibold">
                  <FileIcon entry={{ name: editor.name || editor.path, path: editor.path }} />
                  <span className="truncate">{editor.name || editor.path}</span>
                </div>
                <div className="mt-0.5 truncate text-sm text-muted-foreground">{editor.path}</div>
              </div>
              <div className="flex items-center gap-2">
                <Button className="whitespace-nowrap" variant="outline" onClick={() => setEditor(null)}>
                  <ArrowLeft className="h-4 w-4" />
                  返回
                </Button>
                <Button className="min-w-20 whitespace-nowrap" onClick={() => void saveEditor()} disabled={saving}>
                  <Save className="h-4 w-4" />
                  {saving ? "保存中" : "保存"}
                </Button>
              </div>
            </>
          ) : (
            <>
              <div>
                <div className="font-semibold">文件管理</div>
                <div className="text-sm text-muted-foreground">当前路径：{path || "."}</div>
              </div>
              <div className="grid w-full min-w-0 grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-2 sm:w-auto sm:grid-cols-none sm:flex">
                <Input
                  className="min-w-0"
                  value={path}
                  onChange={(event) => dispatch(filesActions.setPath(event.target.value))}
                  onKeyDown={(event) => {
                    if (event.key === "Enter") void load();
                  }}
                  placeholder="路径"
                />
                <Button className="min-w-20 justify-center whitespace-nowrap px-3 sm:px-4" variant="outline" onClick={() => load()}>
                  <RefreshCcw className="h-4 w-4" />
                  刷新
                </Button>
                <label>
                  <input className="hidden" type="file" onChange={(event) => void handleUpload(event.target.files?.[0])} />
                  <span className="inline-flex h-9 min-w-20 cursor-pointer items-center justify-center gap-2 whitespace-nowrap rounded-md border border-primary/70 bg-primary px-3 text-sm font-semibold text-primary-foreground shadow-[2px_3px_0_hsl(214_45%_30%/0.16)] transition-colors hover:bg-primary/90 sm:px-4">
                    <Upload className="h-4 w-4" />
                    {uploading ? "上传中" : "上传"}
                  </span>
                </label>
              </div>
            </>
          )}
        </PanelHeader>
        <PanelBody className={cn(editor && "flex min-h-0 flex-1 flex-col")}>
          {editorError && !editor ? <div className="mb-3 rounded-lg border border-destructive/40 bg-destructive/5 px-3 py-2 text-sm text-destructive">{editorError}</div> : null}
          {editor ? (
            <div className="flex min-h-0 flex-1 flex-col gap-3">
              {editorError ? <div className="rounded-lg border border-destructive/40 bg-destructive/5 px-3 py-2 text-sm text-destructive">{editorError}</div> : null}
              <textarea
                className="min-h-0 flex-1 w-full resize-none rounded-lg border border-input bg-card/80 p-4 font-mono text-sm leading-6 text-foreground outline-none transition-colors placeholder:text-muted-foreground focus:border-ring focus:ring-2 focus:ring-ring/30"
                value={draft}
                onChange={(event) => setDraft(event.target.value)}
                spellCheck={false}
              />
            </div>
          ) : (
            <div className="overflow-hidden rounded-md border">
              <table className="w-full text-sm">
                <thead className="bg-muted text-muted-foreground">
                  <tr>
                    <th className="px-3 py-2 text-left">名称</th>
                    <th className="hidden px-3 py-2 text-left sm:table-cell">大小</th>
                    <th className="px-3 py-2 text-left">修改时间</th>
                    <th className="px-3 py-2 text-right">操作</th>
                  </tr>
                </thead>
                <tbody>
                  {path ? (
                    <tr className="border-t">
                      <td className="px-3 py-2" colSpan={4}>
                        <button
                          className="text-primary"
                          onClick={() => {
                            setEditor(null);
                            void load(path.split("/").slice(0, -1).join("/"));
                          }}
                        >
                          返回上级
                        </button>
                      </td>
                    </tr>
                  ) : null}
                  {entries.map((entry) => (
                    <tr key={entry.path} className="group border-t transition-colors hover:bg-card/70">
                      <td className="min-w-0 px-3 py-2">
                        <button className="flex max-w-full min-w-0 items-center gap-2 text-left" onClick={() => openEntry(entry)}>
                          <FileIcon entry={entry} />
                          <span className="min-w-0 truncate">{entry.name}</span>
                        </button>
                      </td>
                      <td className="hidden px-3 py-2 text-muted-foreground sm:table-cell">{entry.is_dir ? "-" : formatBytes(entry.size)}</td>
                      <td className="px-3 py-2 text-muted-foreground">{formatTime(entry.mod_time)}</td>
                      <td className="px-2 py-2 text-right sm:px-3">
                        <div className="hidden justify-end gap-1 sm:flex">
                          <FileActionButtons
                            entry={entry}
                            onEdit={() => void editFile(entry.path)}
                            onShare={() => shareEntry(entry)}
                            onDownload={() => downloadEntry(entry)}
                            onDelete={() => void removeEntry(entry)}
                          />
                        </div>
                        <div className="relative flex justify-end sm:hidden">
                          <Button
                            size="icon"
                            variant="ghost"
                            onClick={() => setOpenActionPath(openActionPath === entry.path ? "" : entry.path)}
                            aria-label="更多操作"
                            aria-expanded={openActionPath === entry.path}
                          >
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </PanelBody>
      </Panel>
      {actionEntry ? (
        <FileActionSheet
          entry={actionEntry}
          onOpen={() => openEntry(actionEntry)}
          onEdit={() => void editFile(actionEntry.path)}
          onShare={() => shareEntry(actionEntry)}
          onDownload={() => downloadEntry(actionEntry)}
          onDelete={() => void removeEntry(actionEntry)}
          onClose={() => setOpenActionPath("")}
        />
      ) : null}
      <FileShareDialog file={shareTarget} onClose={() => setShareTarget(null)} />
    </div>
  );
}

function FileIcon({ entry }: { entry: FileEntry }) {
  const Icon = fileIcon(entry);
  return <Icon className="h-4 w-4 shrink-0 text-muted-foreground group-hover:text-foreground" />;
}

function FileActionButtons({
  entry,
  onEdit,
  onShare,
  onDownload,
  onDelete,
}: {
  entry: FileEntry;
  onEdit: () => void;
  onShare: () => void;
  onDownload: () => void;
  onDelete: () => void;
}) {
  return (
    <>
      {!entry.is_dir ? (
        <>
          <Button size="icon" variant="ghost" onClick={onEdit} aria-label="编辑">
            <FilePenLine className="h-4 w-4" />
          </Button>
          <Button size="icon" variant="ghost" onClick={onShare} aria-label="分享文件">
            <Share2 className="h-4 w-4" />
          </Button>
          <Button size="icon" variant="ghost" onClick={onDownload} aria-label="下载">
            <Download className="h-4 w-4" />
          </Button>
        </>
      ) : null}
      <Button size="icon" variant="ghost" onClick={onDelete} aria-label="删除">
        <Trash2 className="h-4 w-4" />
      </Button>
    </>
  );
}

function FileActionSheet({
  entry,
  onOpen,
  onEdit,
  onShare,
  onDownload,
  onDelete,
  onClose,
}: {
  entry: FileEntry;
  onOpen: () => void;
  onEdit: () => void;
  onShare: () => void;
  onDownload: () => void;
  onDelete: () => void;
  onClose: () => void;
}) {
  function run(action: () => void) {
    onClose();
    action();
  }

  return (
    <div className="fixed inset-0 z-50 sm:hidden" role="dialog" aria-modal="true">
      <button className="absolute inset-0 bg-foreground/15 backdrop-blur-[1px]" type="button" aria-label="关闭文件操作菜单" onClick={onClose} />
      <div className="sketch-surface absolute inset-x-3 bottom-3 rounded-2xl bg-card p-2 text-sm shadow-[0_18px_48px_hsl(218_30%_20%/0.2)]">
        <div className="px-3 pb-2 pt-1">
          <div className="truncate text-base font-semibold text-foreground">{entry.name}</div>
          <div className="mt-0.5 truncate text-xs text-muted-foreground">{entry.path}</div>
        </div>
        {entry.is_dir ? (
          <button className="flex h-11 w-full items-center gap-3 rounded-xl px-3 text-left hover:bg-accent/65" type="button" onClick={() => run(onOpen)}>
            <Folder className="h-4 w-4" />
            打开
          </button>
        ) : (
          <>
            <button className="flex h-11 w-full items-center gap-3 rounded-xl px-3 text-left hover:bg-accent/65" type="button" onClick={() => run(onEdit)}>
              <FilePenLine className="h-4 w-4" />
              编辑
            </button>
            <button className="flex h-11 w-full items-center gap-3 rounded-xl px-3 text-left hover:bg-accent/65" type="button" onClick={() => run(onShare)}>
              <Share2 className="h-4 w-4" />
              分享
            </button>
            <button className="flex h-11 w-full items-center gap-3 rounded-xl px-3 text-left hover:bg-accent/65" type="button" onClick={() => run(onDownload)}>
              <Download className="h-4 w-4" />
              下载
            </button>
          </>
        )}
        <div className="my-1 border-t border-border/70" />
        <button className="flex h-11 w-full items-center gap-3 rounded-xl px-3 text-left text-destructive hover:bg-destructive/10" type="button" onClick={() => run(onDelete)}>
          <Trash2 className="h-4 w-4" />
          删除
        </button>
      </div>
    </div>
  );
}

function fileIcon(entry: FileEntry): LucideIcon {
  if (entry.is_dir) return Folder;
  const ext = entry.name.split(".").pop()?.toLowerCase() || "";
  if (["png", "jpg", "jpeg", "gif", "webp", "svg", "bmp"].includes(ext)) return Image;
  if (["md", "txt", "log", "json", "yaml", "yml", "toml", "csv"].includes(ext)) return FileText;
  if (["go", "ts", "tsx", "js", "jsx", "css", "html", "sh", "py", "rs", "java", "c", "cpp", "h"].includes(ext)) return Code2;
  if (["zip", "tar", "gz", "tgz", "rar", "7z"].includes(ext)) return Archive;
  return File;
}
