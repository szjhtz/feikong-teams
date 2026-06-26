import { useState } from "react";
import { LogIn } from "lucide-react";
import { login } from "@/api/auth";
import { setAuthToken } from "@/lib/storage";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Panel, PanelBody, PanelHeader } from "@/components/ui/panel";

export function LoginPage() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  async function submit(event: React.FormEvent) {
    event.preventDefault();
    setError("");
    try {
      const result = await login(username, password);
      setAuthToken(result.token || "");
      location.href = "/chat";
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-muted/40 p-4">
      <Panel className="w-full max-w-sm">
        <PanelHeader>
          <div className="flex items-center gap-3">
            <img className="h-9 w-9" src="/assets/fkteams-logo.svg" alt="" />
            <div>
              <div className="font-semibold">非空小队</div>
              <div className="text-sm text-muted-foreground">登录后继续使用</div>
            </div>
          </div>
        </PanelHeader>
        <PanelBody>
          <form className="space-y-3" onSubmit={submit}>
            <Input value={username} onChange={(event) => setUsername(event.target.value)} placeholder="用户名" autoComplete="username" />
            <Input value={password} onChange={(event) => setPassword(event.target.value)} placeholder="密码" type="password" autoComplete="current-password" />
            {error ? <div className="text-sm text-destructive">{error}</div> : null}
            <Button className="w-full" type="submit">
              <LogIn className="h-4 w-4" />
              登录
            </Button>
          </form>
        </PanelBody>
      </Panel>
    </div>
  );
}
