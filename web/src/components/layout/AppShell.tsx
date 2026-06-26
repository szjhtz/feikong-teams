import { Menu } from "lucide-react";
import { appActions } from "@/app/store";
import { useAppDispatch, useAppSelector } from "@/app/hooks";
import { Button } from "@/components/ui/button";
import { Sidebar } from "./Sidebar";

export function AppShell({ children }: { children: React.ReactNode }) {
  const dispatch = useAppDispatch();
  const connectionState = useAppSelector((state) => state.chat.connectionState);
  const activePanel = useAppSelector((state) => state.app.activePanel);
  const toast = useAppSelector((state) => state.app.toast);
  const statusText = useAppSelector((state) => state.chat.statusText);
  const title = {
    chat: "对话",
    files: "文件",
    schedules: "任务",
    skills: "技能",
    config: "配置",
  }[activePanel];

  return (
    <div className="flex h-screen overflow-hidden bg-background/95 text-foreground">
      <Sidebar />
      <main className="flex min-w-0 flex-1 flex-col">
        <header className="sketch-rule flex h-14 items-center justify-between border-b bg-card/70 px-5 backdrop-blur">
          <div className="flex items-center gap-2">
            <Button
              className="md:hidden"
              size="icon"
              variant="ghost"
              aria-label="打开导航"
              onClick={() => dispatch(appActions.setSidebarOpen(true))}
            >
              <Menu className="h-4 w-4" />
            </Button>
            <div>
              <div className="text-base font-semibold">{title}</div>
            </div>
          </div>
          <div className="flex items-center gap-3 text-xs text-muted-foreground">
            {statusText ? <span className="hidden max-w-72 truncate md:inline">{statusText}</span> : null}
            <span
              className={
                connectionState === "connected"
                  ? "h-2.5 w-2.5 rounded-full bg-emerald-500 shadow-[0_0_0_3px_hsl(152_70%_45%/0.12)]"
                  : "h-2.5 w-2.5 rounded-full bg-amber-500 shadow-[0_0_0_3px_hsl(38_90%_45%/0.12)]"
              }
            />
            {connectionState === "connected" ? "已连接" : connectionState === "connecting" ? "连接中" : "未连接"}
          </div>
        </header>
        <div className="min-h-0 flex-1 overflow-hidden">{children}</div>
      </main>
      {toast ? (
        <div className="sketch-surface fixed bottom-4 right-4 z-50 rounded-md px-4 py-3 text-sm">
          {toast}
        </div>
      ) : null}
    </div>
  );
}
