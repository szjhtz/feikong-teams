import { afterAll, describe, expect, test } from "bun:test";
import { JSDOM } from "jsdom";

const previousWindow = globalThis.window;
const previousDocument = globalThis.document;
const testDOM = new JSDOM("<!doctype html><html><body></body></html>");

Object.assign(globalThis, {
  document: testDOM.window.document,
  window: testDOM.window,
});

const { renderMarkdown } = await import("./markdown");

afterAll(() => {
  Object.assign(globalThis, {
    document: previousDocument,
    window: previousWindow,
  });
  testDOM.window.close();
});

describe("renderMarkdown", () => {
  test("removes executable HTML and event handlers", () => {
    const html = renderMarkdown('<script>globalThis.pwned = true</script><img src="x" onerror="globalThis.pwned = true"><iframe srcdoc="<script>alert(1)</script>"></iframe>');

    expect(html).not.toContain("<script");
    expect(html).not.toContain("onerror");
    expect(html).not.toContain("<iframe");
    expect(html).not.toContain("srcdoc");
  });

  test("removes unsafe link protocols", () => {
    const html = renderMarkdown('[unsafe](javascript:alert(1)) <a href="data:text/html,<script>alert(1)</script>">data</a>');

    expect(html).not.toContain("javascript:");
    expect(html).not.toContain("data:text/html");
  });

  test("sanitizes generated footnote content", () => {
    const html = renderMarkdown('正文[^x]\n\n[^x]: <img src="x" onerror="alert(1)">说明');

    expect(html).toContain("markdown-footnotes");
    expect(html).not.toContain("onerror");
  });

  test("preserves trusted markdown enhancements", () => {
    const html = renderMarkdown('[官网](https://example.com)\n\n```go\nfmt.Println("ok")\n```\n\n$x^2$');

    expect(html).toContain('target="_blank"');
    expect(html).toContain('rel="noreferrer noopener"');
    expect(html).toContain("data-markdown-copy");
    expect(html).toContain("katex");
  });
});
