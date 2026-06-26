import { useEffect, useRef } from "react";
import type { Application } from "pixi.js";

export function ActivityCanvas() {
  const hostRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    let disposed = false;
    let app: Application | undefined;

    async function mount() {
      const host = hostRef.current;
      if (!host) return;
      const pixi = await import("pixi.js");
      app = new pixi.Application();
      await app.init({
        width: 360,
        height: 120,
        backgroundAlpha: 0,
        antialias: true,
        autoDensity: true,
        resolution: window.devicePixelRatio || 1,
      });
      if (disposed || !hostRef.current) {
        app.destroy(true);
        return;
      }

      app.canvas.className = "h-full w-full";
      host.appendChild(app.canvas);

      const graphics = new pixi.Graphics();
      app.stage.addChild(graphics);

      app.ticker.add((ticker) => {
        const t = ticker.lastTime / 1000;
        graphics.clear();
        for (let x = 20; x <= 340; x += 40) {
          graphics.moveTo(x, 18).lineTo(x, 102).stroke({ color: 0xd7dde8, width: 1, alpha: 0.55 });
        }
        for (let y = 24; y <= 96; y += 24) {
          graphics.moveTo(16, y).lineTo(344, y).stroke({ color: 0xd7dde8, width: 1, alpha: 0.55 });
        }
        for (let i = 0; i < 6; i += 1) {
          const x = 32 + ((t * 38 + i * 54) % 296);
          const y = 24 + ((i % 4) * 24);
          graphics.roundRect(x, y - 5, 18, 10, 3).fill({ color: 0x2563eb, alpha: 0.72 });
        }
      });
    }

    void mount();
    return () => {
      disposed = true;
      app?.destroy(true);
      if (hostRef.current) hostRef.current.innerHTML = "";
    };
  }, []);

  return <div ref={hostRef} className="mx-auto mb-5 h-[120px] w-full max-w-[360px]" aria-hidden="true" />;
}
