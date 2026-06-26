import { useEffect, useState } from "react";
import { accessPublicShare, getPublicShareInfo } from "@/api/shares";
import type { SessionDetail } from "@/types/chat";
import { renderMarkdown } from "@/lib/markdown";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Panel, PanelBody, PanelHeader } from "@/components/ui/panel";

export function SharePage() {
  const shareID = decodeURIComponent(location.pathname.split("/").filter(Boolean).pop() || "");
  const [title, setTitle] = useState("");
  const [password, setPassword] = useState("");
  const [detail, setDetail] = useState<SessionDetail | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    if (shareID) getPublicShareInfo(shareID).then((info) => setTitle(info.title || info.share_id));
  }, [shareID]);

  async function load() {
    setError("");
    try {
      setDetail(await accessPublicShare(shareID, password));
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  return (
    <div className="min-h-screen bg-muted/30 p-4">
      <Panel className="mx-auto max-w-4xl">
        <PanelHeader>
          <div className="font-semibold">{title || "共享会话"}</div>
          <div className="text-sm text-muted-foreground">公开只读视图</div>
        </PanelHeader>
        <PanelBody className="space-y-4">
          {!detail ? (
            <div className="flex max-w-sm gap-2">
              <Input value={password} onChange={(event) => setPassword(event.target.value)} placeholder="访问密码（如需要）" />
              <Button onClick={load}>查看</Button>
            </div>
          ) : null}
          {error ? <div className="text-sm text-destructive">{error}</div> : null}
          {detail?.messages?.map((message, index) => (
            <div key={index} className="rounded-md border p-3">
              <div className="mb-2 text-xs text-muted-foreground">{message.agent_name || message.role}</div>
              <div
                className="prose prose-sm max-w-none"
                dangerouslySetInnerHTML={{
                  __html: renderMarkdown(
                    (message.events || []).map((event) => event.content || event.tool_call?.result || "").join("\n") ||
                      message.content ||
                      "",
                  ),
                }}
              />
            </div>
          ))}
        </PanelBody>
      </Panel>
    </div>
  );
}
