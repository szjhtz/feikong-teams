import { Marked } from "marked";

const marked = new Marked({
  breaks: true,
  gfm: true,
});

export function renderMarkdown(value?: string) {
  if (!value) return "";
  try {
    return marked.parse(value) as string;
  } catch {
    return escapeHTML(value).replace(/\n/g, "<br />");
  }
}

export function escapeHTML(value: string) {
  return value
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}
