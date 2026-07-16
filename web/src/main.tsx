import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { App } from "./App";
import { clearLegacyAuthStorage } from "@/lib/storage";
import "katex/dist/katex.min.css";
import "./styles/globals.css";

clearLegacyAuthStorage();
syncAppViewportHeight();

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);

function syncAppViewportHeight() {
  const setHeight = () => {
    const viewport = window.visualViewport;
    const height = viewport?.height || window.innerHeight;
    const keyboardInset = viewport
      ? Math.max(0, window.innerHeight - viewport.height - viewport.offsetTop)
      : 0;

    document.documentElement.style.setProperty("--app-viewport-height", `${height}px`);
    document.documentElement.style.setProperty("--app-keyboard-inset-bottom", `${keyboardInset}px`);
  };
  const viewport = window.visualViewport;

  setHeight();
  window.addEventListener("resize", setHeight);
  window.addEventListener("orientationchange", setHeight);
  viewport?.addEventListener("resize", setHeight);
  viewport?.addEventListener("scroll", setHeight);
}
