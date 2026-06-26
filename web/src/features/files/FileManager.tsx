import { Download, Eye, Folder, RefreshCcw, Trash2, Upload } from "lucide-react";
import { useEffect, useState } from "react";
import { createPreviewLink, deleteFile, listFiles, uploadFile } from "@/api/files";
import { filesActions, appActions } from "@/app/store";
import { useAppDispatch, useAppSelector } from "@/app/hooks";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Panel, PanelBody, PanelHeader } from "@/components/ui/panel";
import { formatBytes, formatTime } from "@/lib/format";

export function FileManager() {
  const dispatch = useAppDispatch();
  const path = useAppSelector((state) => state.files.path);
  const entries = useAppSelector((state) => state.files.entries);
  const [uploading, setUploading] = useState(false);

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

  async function preview(filePath: string) {
    const link = await createPreviewLink(filePath);
    const id = link.link_id || link.id;
    if (id) window.open(`/p/${encodeURIComponent(id)}`, "_blank");
  }

  useEffect(() => {
    void load("");
  }, []);

  return (
    <div className="h-full overflow-auto p-6">
      <Panel className="mx-auto max-w-6xl">
        <PanelHeader className="flex items-center justify-between">
          <div>
            <div className="font-semibold">文件管理</div>
            <div className="text-sm text-muted-foreground">当前路径：{path || "."}</div>
          </div>
          <div className="flex items-center gap-2">
            <Input value={path} onChange={(event) => dispatch(filesActions.setPath(event.target.value))} placeholder="路径" />
            <Button variant="outline" onClick={() => load()}>
              <RefreshCcw className="h-4 w-4" />
              刷新
            </Button>
            <label>
              <input className="hidden" type="file" onChange={(event) => void handleUpload(event.target.files?.[0])} />
              <span className="inline-flex h-9 cursor-pointer items-center gap-2 rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground">
                <Upload className="h-4 w-4" />
                {uploading ? "上传中" : "上传"}
              </span>
            </label>
          </div>
        </PanelHeader>
        <PanelBody>
          <div className="overflow-hidden rounded-md border">
            <table className="w-full text-sm">
              <thead className="bg-muted text-muted-foreground">
                <tr>
                  <th className="px-3 py-2 text-left">名称</th>
                  <th className="px-3 py-2 text-left">大小</th>
                  <th className="px-3 py-2 text-left">修改时间</th>
                  <th className="px-3 py-2 text-right">操作</th>
                </tr>
              </thead>
              <tbody>
                {path ? (
                  <tr className="border-t">
                    <td className="px-3 py-2" colSpan={4}>
                      <button className="text-primary" onClick={() => load(path.split("/").slice(0, -1).join("/"))}>
                        返回上级
                      </button>
                    </td>
                  </tr>
                ) : null}
                {entries.map((entry) => (
                  <tr key={entry.path} className="border-t">
                    <td className="px-3 py-2">
                      <button className="flex items-center gap-2" onClick={() => entry.is_dir && load(entry.path)}>
                        {entry.is_dir ? <Folder className="h-4 w-4" /> : <span className="h-4 w-4" />}
                        {entry.name}
                      </button>
                    </td>
                    <td className="px-3 py-2 text-muted-foreground">{entry.is_dir ? "-" : formatBytes(entry.size)}</td>
                    <td className="px-3 py-2 text-muted-foreground">{formatTime(entry.mod_time)}</td>
                    <td className="space-x-1 px-3 py-2 text-right">
                      {!entry.is_dir ? (
                        <>
                          <Button size="icon" variant="ghost" onClick={() => preview(entry.path)} aria-label="预览">
                            <Eye className="h-4 w-4" />
                          </Button>
                          <Button size="icon" variant="ghost" onClick={() => window.open(`/api/fkteams/files/download?path=${encodeURIComponent(entry.path)}`)} aria-label="下载">
                            <Download className="h-4 w-4" />
                          </Button>
                        </>
                      ) : null}
                      <Button
                        size="icon"
                        variant="ghost"
                        onClick={() => deleteFile(entry.path).then(() => load()).then(() => dispatch(appActions.showToast("已删除")))}
                        aria-label="删除"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </PanelBody>
      </Panel>
    </div>
  );
}
