import { CornerDownRight, MoreHorizontal, Trash2 } from "lucide-react";
import { useState } from "react";
import { useAppDispatch, useAppSelector } from "@/app/hooks";
import { chatActions } from "@/app/store";
import { changeQueueKind, deleteQueueItem, moveQueueItem, updateQueueItem } from "@/api/stream";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/cn";

export function QueuePanel() {
  const dispatch = useAppDispatch();
  const sessionID = useAppSelector((state) => state.chat.activeSessionID);
  const queue = useAppSelector((state) => state.chat.queue);
  const [openMenuID, setOpenMenuID] = useState("");
  if (!queue.length) return null;

  async function refresh(action: Promise<{ queue: typeof queue }>) {
    const result = await action;
    setOpenMenuID("");
    dispatch(chatActions.setQueue(result.queue || []));
  }

  return (
    <div className="relative z-10 -mb-2 bg-transparent px-3 pb-0 pt-2 sm:px-6">
      <div className="mx-auto max-w-4xl overflow-hidden rounded-2xl border border-border/55 bg-card/85 px-3 py-3 shadow-[0_10px_30px_hsl(218_30%_25%/0.06)] backdrop-blur sm:rounded-[1.35rem] sm:px-5">
        {queue.map((item, index) => (
          <div
            key={item.queue_id}
            className={cn(
              "group relative flex min-h-10 items-center gap-2 sm:gap-3",
              index > 0 && "pt-1.5",
              index < queue.length - 1 && "pb-1.5",
            )}
          >
            <CornerDownRight className="h-4 w-4 shrink-0 text-muted-foreground/50" />
            <Input
              className="h-8 min-w-0 flex-1 border-0 bg-transparent px-0 text-base font-semibold text-muted-foreground shadow-none focus-visible:ring-0"
              defaultValue={queueItemText(item)}
              onBlur={(event) => refresh(updateQueueItem(sessionID, item.queue_id, event.target.value))}
            />
            <Button
              className={cn(
                "h-8 shrink-0 px-2 text-base font-semibold text-muted-foreground hover:text-foreground",
                item.kind === "steering" && "text-foreground",
              )}
              size="sm"
              variant="ghost"
              onClick={() => refresh(changeQueueKind(sessionID, item.queue_id, "steering"))}
            >
              <CornerDownRight className="h-4 w-4" />
              引导
            </Button>
            <Button className="h-8 w-8 text-muted-foreground hover:text-foreground" size="icon" variant="ghost" onClick={() => refresh(deleteQueueItem(sessionID, item.queue_id))} aria-label="删除">
              <Trash2 className="h-4 w-4" />
            </Button>
            <Button
              className="h-8 w-8 text-muted-foreground hover:text-foreground"
              size="icon"
              variant="ghost"
              onClick={() => setOpenMenuID(openMenuID === item.queue_id ? "" : item.queue_id)}
              aria-label="更多"
              aria-expanded={openMenuID === item.queue_id}
            >
              <MoreHorizontal className="h-4 w-4" />
            </Button>
            {openMenuID === item.queue_id ? (
              <div className="absolute right-0 top-9 z-20 w-28 overflow-hidden rounded-md border border-border bg-card py-1 text-sm shadow-[0_10px_24px_hsl(218_30%_25%/0.14)]">
                <button className="block w-full px-3 py-2 text-left hover:bg-muted" type="button" onClick={() => refresh(moveQueueItem(sessionID, item.queue_id, "up"))}>
                  上移
                </button>
                <button className="block w-full px-3 py-2 text-left hover:bg-muted" type="button" onClick={() => refresh(moveQueueItem(sessionID, item.queue_id, "down"))}>
                  下移
                </button>
                <button className="block w-full px-3 py-2 text-left hover:bg-muted" type="button" onClick={() => refresh(changeQueueKind(sessionID, item.queue_id, "follow_up"))}>
                  后续
                </button>
              </div>
            ) : null}
          </div>
        ))}
      </div>
    </div>
  );
}

function queueItemText(item: { display_text?: string; text?: string; content?: string; message?: string }) {
  return item.display_text || item.text || item.content || item.message || "";
}
