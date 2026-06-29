import { Marked } from "marked";
import markedKatex from "marked-katex-extension";

const marked = new Marked({
  breaks: true,
  gfm: true,
});

marked.use(markedKatex({
  nonStandard: true,
  throwOnError: false,
}));

export function renderMarkdown(value?: string) {
  if (!value) return "";
  try {
    return marked.parse(normalizeMathDelimiters(value)) as string;
  } catch {
    return escapeHTML(value).replace(/\n/g, "<br />");
  }
}

function normalizeMathDelimiters(value: string) {
  return value
    .split(/(```[\s\S]*?```|~~~[\s\S]*?~~~|`[^`\n]*`)/g)
    .map((part) => {
      if (part.startsWith("```") || part.startsWith("~~~") || part.startsWith("`")) return part;
      return part
        .replace(/\\\[([\s\S]*?)\\\]/g, (_, formula: string) => `\n$$\n${formula.trim()}\n$$\n`)
        .replace(/\\\(([\s\S]*?)\\\)/g, (_, formula: string) => `$${formula.trim()}$`);
    })
    .join("");
}

export function escapeHTML(value: string) {
  return value
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}
