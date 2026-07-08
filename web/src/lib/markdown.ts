import { Marked } from "marked";
import hljs from "highlight.js/lib/core";
import bash from "highlight.js/lib/languages/bash";
import css from "highlight.js/lib/languages/css";
import diff from "highlight.js/lib/languages/diff";
import go from "highlight.js/lib/languages/go";
import javascript from "highlight.js/lib/languages/javascript";
import json from "highlight.js/lib/languages/json";
import markdown from "highlight.js/lib/languages/markdown";
import python from "highlight.js/lib/languages/python";
import sql from "highlight.js/lib/languages/sql";
import typescript from "highlight.js/lib/languages/typescript";
import xml from "highlight.js/lib/languages/xml";
import yaml from "highlight.js/lib/languages/yaml";
import markedKatex from "marked-katex-extension";

const marked = new Marked({
  breaks: true,
  gfm: true,
});

marked.use(markedKatex({
  nonStandard: true,
  throwOnError: false,
  strict: "ignore",
}));

hljs.registerLanguage("bash", bash);
hljs.registerLanguage("sh", bash);
hljs.registerLanguage("shell", bash);
hljs.registerLanguage("zsh", bash);
hljs.registerLanguage("css", css);
hljs.registerLanguage("diff", diff);
hljs.registerLanguage("go", go);
hljs.registerLanguage("golang", go);
hljs.registerLanguage("javascript", javascript);
hljs.registerLanguage("js", javascript);
hljs.registerLanguage("jsx", javascript);
hljs.registerLanguage("json", json);
hljs.registerLanguage("markdown", markdown);
hljs.registerLanguage("md", markdown);
hljs.registerLanguage("python", python);
hljs.registerLanguage("py", python);
hljs.registerLanguage("sql", sql);
hljs.registerLanguage("typescript", typescript);
hljs.registerLanguage("ts", typescript);
hljs.registerLanguage("tsx", typescript);
hljs.registerLanguage("html", xml);
hljs.registerLanguage("xml", xml);
hljs.registerLanguage("yaml", yaml);
hljs.registerLanguage("yml", yaml);

export function renderMarkdown(value?: string) {
  if (!value) return "";
  try {
    const footnoteInput = normalizeFootnotes(value);
    const html = withTableWrappers(withCodeCopyButtons(marked.parse(normalizeMathDelimiters(footnoteInput.markdown)) as string)) + footnoteInput.html;
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
  return html.replace(/<pre><code([^>]*)>([\s\S]*?)<\/code><\/pre>/g, (_match, attrs: string, code: string) => {
    const language = codeLanguageInfo(attrs);
    const highlighted = highlightCodeBlock(code, language.key);
    return `<div class="markdown-code-block"><div class="markdown-code-header"><span class="markdown-code-language">${language.label}</span><button class="markdown-code-copy" type="button" data-markdown-copy title="复制代码" aria-label="复制代码">复制</button></div><pre><code${attrs}>${highlighted}</code></pre></div>`;
  });
}

function codeLanguageInfo(attrs: string) {
  const match = attrs.match(/class="[^"]*\blanguage-([^"\s]+)[^"]*"/);
  if (!match) return { key: "text", label: "text" };
  return { key: match[1].toLowerCase(), label: escapeHTML(match[1].replace(/[-_]/g, " ")) };
}

function highlightCodeBlock(code: string, language: string) {
  const raw = decodeHTML(code);
  return highlightCode(raw, language);
}

export function highlightCode(value: string, language = "text") {
  const raw = value || "";
  if (language !== "text" && hljs.getLanguage(language)) {
    return hljs.highlight(raw, { language, ignoreIllegals: true }).value;
  }
  return hljs.highlightAuto(raw).value;
}

function decodeHTML(value: string) {
  return value
    .replace(/&quot;/g, "\"")
    .replace(/&#039;/g, "'")
    .replace(/&apos;/g, "'")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&amp;/g, "&");
}

function withTableWrappers(html: string) {
  return html.replace(/<table([^>]*)>([\s\S]*?)<\/table>/g, (_match, attrs: string, content: string) => (
    `<div class="markdown-table-wrap"><table${attrs}>${content}</table></div>`
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
