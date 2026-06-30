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
    const footnoteInput = normalizeFootnotes(value);
    const html = withCodeCopyButtons(marked.parse(normalizeMathDelimiters(footnoteInput.markdown)) as string) + footnoteInput.html;
    return withExternalLinkTargets(html);
  } catch {
    return escapeHTML(value).replace(/\n/g, "<br />");
  }
}

function withExternalLinkTargets(html: string) {
  return html.replace(/<a\s+([^>]*href="(?!#)[^"]+"[^>]*)>/g, (match, attrs: string) => {
    const nextAttrs = attrs
      .replace(/\s+target="[^"]*"/g, "")
      .replace(/\s+rel="[^"]*"/g, "");
    return `<a ${nextAttrs} target="_blank" rel="noreferrer noopener">`;
  });
}

function normalizeFootnotes(value: string) {
  const footnotes: Array<{ id: string; content: string }> = [];
  const lines = value.split(/\r?\n/);
  const bodyLines: string[] = [];
  for (const line of lines) {
    const match = line.match(/^\[\^([^\]]+)\]:\s*(.*)$/);
    if (!match) {
      bodyLines.push(line);
      continue;
    }
    footnotes.push({ id: match[1], content: match[2] });
  }

  if (!footnotes.length) return { markdown: value, html: "" };

  return {
    markdown: bodyLines.join("\n").replace(/\[\^([^\]]+)\]/g, (_match, id: string) => (
      `<sup class="markdown-footnote-ref"><a href="#fn-${slugFootnoteID(id)}" id="fnref-${slugFootnoteID(id)}">${escapeHTML(id)}</a></sup>`
    )),
    html: renderFootnotes(footnotes),
  };
}

function renderFootnotes(footnotes: Array<{ id: string; content: string }>) {
  const items = footnotes.map(({ id, content }) => {
    const safeID = slugFootnoteID(id);
    const rendered = marked.parseInline(normalizeMathDelimiters(content)) as string;
    return `<li id="fn-${safeID}"><span class="markdown-footnote-label">${escapeHTML(id)}</span><span class="markdown-footnote-content">${rendered}</span><a class="markdown-footnote-backref" href="#fnref-${safeID}" aria-label="返回正文">↩</a></li>`;
  }).join("");
  return `<section class="markdown-footnotes"><ol>${items}</ol></section>`;
}

function slugFootnoteID(value: string) {
  return encodeURIComponent(value.trim()).replace(/%/g, "");
}

function withCodeCopyButtons(html: string) {
  return html.replace(/<pre><code([^>]*)>([\s\S]*?)<\/code><\/pre>/g, (_match, attrs: string, code: string) => (
    `<div class="markdown-code-block"><button class="markdown-code-copy" type="button" data-markdown-copy title="复制代码" aria-label="复制代码">复制</button><pre><code${attrs}>${code}</code></pre></div>`
  ));
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
