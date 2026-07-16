import type { MouseEvent } from "react";
import { copyText } from "@/lib/clipboard";
import { cn } from "@/lib/cn";
import { renderMarkdown } from "@/lib/markdown";

const copyResetTimers = new WeakMap<HTMLButtonElement, number>();

export function MarkdownContent({ content, className }: { content?: string; className?: string }) {
  return (
    <div
      className={cn("prose message-prose w-full max-w-none", className)}
      onClick={(event) => void handleMarkdownClick(event)}
      dangerouslySetInnerHTML={{ __html: renderMarkdown(content) }}
    />
  );
}

async function handleMarkdownClick(event: MouseEvent<HTMLDivElement>) {
  const target = event.target instanceof Element ? event.target : null;
  const button = target?.closest<HTMLButtonElement>("[data-markdown-copy]");
  if (!button) return;

  const block = button.closest(".markdown-code-block");
  const code = block?.querySelector("pre code")?.textContent || "";
  if (!code) return;

  try {
    await copyText(code);
  } catch {
    return;
  }
  const previous = button.textContent || "复制";
  button.textContent = "已复制";
  button.dataset.copied = "true";
  const previousTimer = copyResetTimers.get(button);
  if (previousTimer !== undefined) window.clearTimeout(previousTimer);
  const timer = window.setTimeout(() => {
    button.textContent = previous;
    delete button.dataset.copied;
    copyResetTimers.delete(button);
  }, 1200);
  copyResetTimers.set(button, timer);
}
