import { useEffect } from "react";
import { Provider } from "react-redux";
import { get } from "@/api/client";
import { listAgents } from "@/api/agents";
import { appActions, chatActions } from "@/app/store";
import { store } from "@/app/store";
import { useAppDispatch, useAppSelector } from "@/app/hooks";
import { AppShell } from "@/components/layout/AppShell";
import { ChatPage } from "@/features/chat/ChatPage";
import { ConfigPanel } from "@/features/config/ConfigPanel";
import { FileManager } from "@/features/files/FileManager";
import { LoginPage } from "@/features/auth/LoginPage";
import { PreviewPage } from "@/features/preview/PreviewPage";
import { SchedulePanel } from "@/features/schedules/SchedulePanel";
import { ShareManagerPanel } from "@/features/share/ShareManagerPanel";
import { SharePage } from "@/features/share/SharePage";
import { SkillPanel } from "@/features/skills/SkillPanel";
import { loadSessions } from "@/features/sessions/sessionThunks";
import { chatSessionIDFromPath, panelFromPath } from "@/lib/navigation";
import type { AgentInfo, VersionInfo } from "@/types/api";

export function App() {
  return (
    <Provider store={store}>
      <Root />
    </Provider>
  );
}

function Root() {
  const path = location.pathname;
  if (path === "/login") return <LoginPage />;
  if (path.startsWith("/p/")) return <PreviewPage />;
  if (path.startsWith("/s/")) return <SharePage />;
  return (
    <AppShell>
      <Workspace />
    </AppShell>
  );
}

function Workspace() {
  const dispatch = useAppDispatch();
  const activePanel = useAppSelector((state) => state.app.activePanel);

  useEffect(() => {
    const syncRoute = () => {
      const panel = panelFromPath(location.pathname);
      dispatch(appActions.setActivePanel(panel));
      if (panel === "chat") {
        const sessionID = chatSessionIDFromPath(location.pathname);
        dispatch(chatActions.setActiveSession(sessionID));
        if (!sessionID) dispatch(chatActions.clearMessages());
      }
    };

    syncRoute();
    void dispatch(loadSessions());
    void get<VersionInfo>("/api/fkteams/version").then((version) => dispatch(appActions.setVersion(version))).catch(() => undefined);
    void listAgents()
      .then((result) => {
        const agents = Array.isArray(result) ? result : result.agents || [];
        dispatch(appActions.setAgents(agents as AgentInfo[]));
      })
      .catch(() => undefined);
    const onAuthExpired = () => dispatch(appActions.setAuthExpired(true));
    window.addEventListener("popstate", syncRoute);
    window.addEventListener("fkteams:auth-expired", onAuthExpired);
    return () => {
      window.removeEventListener("popstate", syncRoute);
      window.removeEventListener("fkteams:auth-expired", onAuthExpired);
    };
  }, [dispatch]);

  switch (activePanel) {
    case "config":
      return <ConfigPanel />;
    case "files":
      return <FileManager />;
    case "schedules":
      return <SchedulePanel />;
    case "shares":
      return <ShareManagerPanel />;
    case "skills":
      return <SkillPanel />;
    case "chat":
    default:
      return <ChatPage />;
  }
}
