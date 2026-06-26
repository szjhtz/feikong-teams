import { useEffect, useRef } from "react";
import anime from "animejs";
import { Bot, User } from "lucide-react";
import { useAppSelector } from "@/app/hooks";
import { renderMarkdown } from "@/lib/markdown";
import { cn } from "@/lib/cn";
import { ActivityCanvas } from "@/components/layout/ActivityCanvas";
import { ToolCallCard } from "./ToolCallCard";

export function MessageList() {
  const messages = useAppSelector((state) => state.chat.messages);
  const events = useAppSelector((state) => state.chat.events);
  const bottomRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ block: "end" });
    anime({
      targets: ".message-row:last-of-type",
      opacity: [0, 1],
      translateY: [8, 0],
      duration: 180,
      easing: "easeOutQuad",
    });
  }, [messages.length, events.length]);

  if (messages.length === 0 && events.length === 0) {
    return (
      <div className="flex h-full items-center justify-center p-10 text-center">
        <div className="max-w-md">
          <ActivityCanvas />
          <div className="mb-3 text-lg font-semibold">准备接收任务</div>
          <p className="text-sm text-muted-foreground">当前会话尚无消息。</p>
        </div>
      </div>
    );
  }

  const toolEvents = events.flatMap((event) => event.tool_calls || (event.tool_call ? [event.tool_call] : []));

  return (
    <div className="h-full overflow-auto px-6 py-5">
      <div className="mx-auto max-w-5xl space-y-4">
        {messages.map((message) => (
          <article key={message.id} className={cn("message-row flex gap-3", message.role === "user" && "justify-end")}>
            {message.role !== "user" ? (
              <div className="mt-1 flex h-8 w-8 items-center justify-center rounded-md bg-secondary">
                <Bot className="h-4 w-4" />
              </div>
            ) : null}
            <div
              className={cn(
                "max-w-[78%] rounded-md border px-4 py-3 text-sm shadow-sm",
                message.role === "user" ? "bg-primary text-primary-foreground" : "bg-card",
              )}
            >
              {message.agent ? <div className="mb-2 text-xs text-muted-foreground">{message.agent}</div> : null}
              <div
                className="prose prose-sm max-w-none dark:prose-invert"
                dangerouslySetInnerHTML={{ __html: renderMarkdown(message.content) }}
              />
            </div>
            {message.role === "user" ? (
              <div className="mt-1 flex h-8 w-8 items-center justify-center rounded-md bg-primary text-primary-foreground">
                <User className="h-4 w-4" />
              </div>
            ) : null}
          </article>
        ))}
        {toolEvents.slice(-8).map((tool, index) => (
          <ToolCallCard key={`${tool.ref || tool.id || tool.name}-${index}`} tool={tool} />
        ))}
        <div ref={bottomRef} />
      </div>
    </div>
  );
}
