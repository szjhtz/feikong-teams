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
    <div className="flex min-h-screen items-center justify-center bg-muted/40 px-4 py-8 sm:px-6">
      <Panel className="w-full max-w-md">
        <PanelHeader className="px-6 py-5 sm:px-7 sm:py-6">
          <div className="flex items-center gap-4">
            <img className="h-12 w-12 shrink-0 drop-shadow-sm" src="/assets/fkteams-logo.svg" alt="" />
            <div>
              <div className="text-2xl font-semibold">非空小队</div>
              <div className="mt-1 text-base text-muted-foreground">登录后继续使用</div>
            </div>
          </div>
        </PanelHeader>
        <PanelBody className="px-6 py-6 sm:px-7">
          <form className="space-y-4" onSubmit={submit}>
            <Input className="h-12 px-4 text-base" value={username} onChange={(event) => setUsername(event.target.value)} placeholder="用户名" autoComplete="username" />
            <Input className="h-12 px-4 text-base" value={password} onChange={(event) => setPassword(event.target.value)} placeholder="密码" type="password" autoComplete="current-password" />
            {error ? <div className="text-sm text-destructive">{error}</div> : null}
            <Button className="h-12 w-full text-base" type="submit">
              <LogIn className="h-5 w-5" />
              登录
            </Button>
          </form>
        </PanelBody>
      </Panel>
    </div>
  );
}
