import { CheckCircle2, Cpu, Loader2, UserRoundCheck } from "lucide-react";
import type { ToolCallDTO } from "@/types/events";
import { Badge } from "@/components/ui/badge";
import { Panel, PanelBody } from "@/components/ui/panel";

export function ToolCallCard({ tool }: { tool: ToolCallDTO }) {
  const isAgent = tool.kind === "agent";
  const isDone = tool.status === "completed";
  return (
    <Panel className={isAgent ? "border-emerald-200 bg-emerald-50/45" : "bg-muted/25"}>
      <PanelBody className="space-y-2 p-3">
        <div className="flex items-center gap-2 text-sm">
          {isAgent ? <UserRoundCheck className="h-4 w-4 text-emerald-600" /> : <Cpu className="h-4 w-4 text-muted-foreground" />}
          <span className="font-medium">{isAgent ? "子智能体" : "工具调用"}</span>
          <code className="rounded bg-background px-1.5 py-0.5 text-xs">{tool.display_name || tool.name}</code>
          <Badge>{tool.kind || "tool"}</Badge>
          {tool.member_name || tool.target ? <Badge>{tool.member_name || tool.target}</Badge> : null}
          <span className="ml-auto flex items-center gap-1 text-xs text-muted-foreground">
            {isDone ? <CheckCircle2 className="h-3.5 w-3.5" /> : <Loader2 className="h-3.5 w-3.5 animate-spin" />}
            {isDone ? "完成" : "运行中"}
          </span>
        </div>
        {tool.arguments ? (
          <pre className="sketch-inset max-h-40 overflow-auto rounded-md p-2 text-xs text-muted-foreground">
            {formatArgs(tool.arguments)}
          </pre>
        ) : null}
        {tool.result ? (
          <pre className="sketch-inset max-h-56 overflow-auto rounded-md p-2 text-xs">
            {formatResult(tool.result)}
          </pre>
        ) : null}
      </PanelBody>
    </Panel>
  );
}

function formatArgs(value: string) {
  try {
    return JSON.stringify(JSON.parse(value), null, 2);
  } catch {
    return value;
  }
}

function formatResult(value: string) {
  const trimmed = value.trim();
  if (!trimmed) return "";
  if (trimmed.length > 5000) return `${trimmed.slice(0, 5000)}\n...`;
  return trimmed;
}
