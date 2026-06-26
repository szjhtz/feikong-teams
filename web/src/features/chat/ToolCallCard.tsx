import { ChevronDown, Cpu, UserRoundCheck } from "lucide-react";
import type { ToolCallDTO } from "@/types/events";
import { Badge } from "@/components/ui/badge";
import { Panel, PanelBody } from "@/components/ui/panel";

export function ToolCallCard({ tool }: { tool: ToolCallDTO }) {
  const isAgent = tool.kind === "agent";
  return (
    <Panel className={isAgent ? "border-emerald-200 bg-emerald-50/50" : "bg-muted/30"}>
      <PanelBody className="space-y-2 p-3">
        <div className="flex items-center gap-2 text-sm">
          {isAgent ? <UserRoundCheck className="h-4 w-4 text-emerald-600" /> : <Cpu className="h-4 w-4 text-muted-foreground" />}
          <span className="font-medium">{isAgent ? "成员指派" : "工具调用"}</span>
          <code className="rounded bg-background px-1.5 py-0.5 text-xs">{tool.display_name || tool.name}</code>
          <Badge>{tool.kind || "tool"}</Badge>
          <ChevronDown className="ml-auto h-4 w-4 text-muted-foreground" />
        </div>
        {tool.arguments ? (
          <pre className="max-h-48 overflow-auto rounded-md bg-background p-2 text-xs text-muted-foreground">
            {formatArgs(tool.arguments)}
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
