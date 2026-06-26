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
  const title = {
    chat: "对话",
    files: "文件",
    schedules: "任务",
    skills: "技能",
    config: "配置",
  }[activePanel];

  return (
    <div className="flex h-screen overflow-hidden bg-background text-foreground">
      <Sidebar />
      <main className="flex min-w-0 flex-1 flex-col">
        <header className="flex h-14 items-center justify-between border-b px-4">
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
              <div className="text-sm font-semibold">{title}</div>
            </div>
          </div>
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <span
              className={
                connectionState === "connected"
                  ? "h-2 w-2 rounded-full bg-emerald-500"
                  : "h-2 w-2 rounded-full bg-amber-500"
              }
            />
            {connectionState === "connected" ? "已连接" : connectionState === "connecting" ? "连接中" : "未连接"}
          </div>
        </header>
        <div className="min-h-0 flex-1 overflow-hidden">{children}</div>
      </main>
      {toast ? (
        <div className="fixed bottom-4 right-4 z-50 rounded-md border bg-popover px-4 py-3 text-sm shadow-lg">
          {toast}
        </div>
      ) : null}
    </div>
  );
}
