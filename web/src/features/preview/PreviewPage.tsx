import { useEffect, useState, type FormEvent } from "react";
import { authorizePreview, getPreviewInfo, type PreviewInfo } from "@/api/preview";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Panel, PanelBody, PanelHeader } from "@/components/ui/panel";

export function PreviewPage() {
  const linkID = previewLinkIDFromPath(location.pathname);
  const encodedLinkID = encodeURIComponent(linkID);
  const [info, setInfo] = useState<PreviewInfo>();
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(true);
  const [authorizing, setAuthorizing] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!linkID) {
      setError("分享链接无效");
      setLoading(false);
      return;
    }
    let active = true;
    setLoading(true);
    getPreviewInfo(linkID)
      .then((result) => {
        if (active) setInfo(result);
      })
      .catch((loadError) => {
        if (active) setError(loadError instanceof Error ? loadError.message : String(loadError));
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, [linkID]);

  const requiresPassword = Boolean(info?.require_password && !info.authorized);
  const previewable = Boolean(info?.previewable);

  async function submitPassword(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!password || authorizing) return;
    setAuthorizing(true);
    setError("");
    try {
      const result = await authorizePreview(linkID, password);
      if (!result.authenticated) throw new Error("访问凭证无效");
      setInfo((current) => ({ ...current, authorized: true }));
      setPassword("");
    } catch (authError) {
      setError(authError instanceof Error ? authError.message : String(authError));
    } finally {
      setAuthorizing(false);
    }
  }

  function download() {
    window.open(`/api/fkteams/preview/${encodedLinkID}?download=1`, "_blank", "noopener,noreferrer");
  }

  return (
    <div className="h-screen bg-muted/30 p-4">
      <Panel className="mx-auto flex h-full max-w-6xl flex-col">
        <PanelHeader className="flex items-center justify-between">
          <div>
            <div className="font-semibold">文件预览</div>
            <div className="text-sm text-muted-foreground">{info?.file_name || linkID}</div>
          </div>
          <Button onClick={download} disabled={loading || requiresPassword || !info}>下载</Button>
        </PanelHeader>
        <PanelBody className="flex min-h-0 flex-1 p-0">
          {loading ? <PreviewNotice title="正在读取分享信息" /> : null}
          {!loading && requiresPassword ? (
            <form className="m-auto w-full max-w-sm space-y-4 p-6" onSubmit={(event) => void submitPassword(event)}>
              <div>
                <div className="text-lg font-semibold">该分享需要访问密码</div>
                <div className="mt-1 text-sm text-muted-foreground">验证后将在一小时内保持访问状态。</div>
              </div>
              <Input
                type="password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                placeholder="访问密码"
                autoComplete="current-password"
                autoFocus
                disabled={authorizing}
              />
              {error ? <div className="text-sm text-destructive">{error}</div> : null}
              <Button className="w-full" type="submit" disabled={!password || authorizing}>
                {authorizing ? "验证中" : "验证并打开"}
              </Button>
            </form>
          ) : null}
          {!loading && !requiresPassword && error ? <PreviewNotice title="无法打开分享" description={error} /> : null}
          {!loading && !requiresPassword && !error && info && !previewable ? (
            <PreviewNotice title="该分享无法在线预览" description="分享包含多个文件或当前文件类型不支持在线预览，请使用下载按钮获取内容。" />
          ) : null}
          {!loading && !requiresPassword && !error && info && previewable ? (
            <iframe
              className="h-full w-full border-0"
              referrerPolicy="no-referrer"
              sandbox=""
              src={`/api/fkteams/preview/${encodedLinkID}/render/`}
              title="文件预览"
            />
          ) : null}
        </PanelBody>
      </Panel>
    </div>
  );
}

function PreviewNotice({ title, description }: { title: string; description?: string }) {
  return (
    <div className="m-auto max-w-lg p-8 text-center">
      <div className="text-lg font-semibold">{title}</div>
      {description ? <div className="mt-2 text-sm leading-6 text-muted-foreground">{description}</div> : null}
    </div>
  );
}

function previewLinkIDFromPath(pathname: string) {
  const encoded = pathname.split("/").filter(Boolean).pop() || "";
  try {
    return decodeURIComponent(encoded);
  } catch {
    return "";
  }
}
