import { ArrowDown, ArrowUp, Trash2 } from "lucide-react";
import { useAppDispatch, useAppSelector } from "@/app/hooks";
import { chatActions } from "@/app/store";
import { changeQueueKind, deleteQueueItem, moveQueueItem, updateQueueItem } from "@/api/stream";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";

export function QueuePanel() {
  const dispatch = useAppDispatch();
  const sessionID = useAppSelector((state) => state.chat.activeSessionID);
  const queue = useAppSelector((state) => state.chat.queue);
  if (!queue.length) return null;

  async function refresh(action: Promise<{ queue: typeof queue }>) {
    const result = await action;
    dispatch(chatActions.setQueue(result.queue || []));
  }

  return (
    <div className="border-t bg-muted/30 px-4 py-3">
      <div className="mx-auto max-w-5xl space-y-2">
        <div className="text-xs font-medium text-muted-foreground">运行中队列</div>
        {queue.map((item) => (
          <div key={item.queue_id} className="flex items-center gap-2 rounded-md border bg-background p-2">
            <Badge>{item.kind}</Badge>
            <Input
              defaultValue={item.content || item.message || ""}
              onBlur={(event) => refresh(updateQueueItem(sessionID, item.queue_id, event.target.value))}
            />
            <Button size="sm" variant="outline" onClick={() => refresh(changeQueueKind(sessionID, item.queue_id, item.kind === "steering" ? "follow_up" : "steering"))}>
              转换
            </Button>
            <Button size="icon" variant="ghost" onClick={() => refresh(moveQueueItem(sessionID, item.queue_id, "up"))} aria-label="上移">
              <ArrowUp className="h-4 w-4" />
            </Button>
            <Button size="icon" variant="ghost" onClick={() => refresh(moveQueueItem(sessionID, item.queue_id, "down"))} aria-label="下移">
              <ArrowDown className="h-4 w-4" />
            </Button>
            <Button size="icon" variant="ghost" onClick={() => refresh(deleteQueueItem(sessionID, item.queue_id))} aria-label="删除">
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>
        ))}
      </div>
    </div>
  );
}
