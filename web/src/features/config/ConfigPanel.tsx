import { Save, RefreshCcw } from "lucide-react";
import { useEffect, useState } from "react";
import { getConfig, getToolCatalog, saveConfig } from "@/api/config";
import { configActions, appActions } from "@/app/store";
import { useAppDispatch, useAppSelector } from "@/app/hooks";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Panel, PanelBody, PanelHeader } from "@/components/ui/panel";

export function ConfigPanel() {
  const dispatch = useAppDispatch();
  const config = useAppSelector((state) => state.config.value);
  const tools = useAppSelector((state) => state.config.tools);
  const [draft, setDraft] = useState("");

  async function load() {
    const [cfg, catalog] = await Promise.all([getConfig(), getToolCatalog().catch(() => [])]);
    dispatch(configActions.setConfig(cfg));
    dispatch(configActions.setTools(catalog));
    setDraft(JSON.stringify(cfg, null, 2));
  }

  async function save() {
    await saveConfig(JSON.parse(draft));
    dispatch(appActions.showToast("配置已保存"));
    await load();
  }

  useEffect(() => {
    void load();
  }, []);

  return (
    <div className="h-full overflow-auto p-6">
      <div className="mx-auto grid max-w-6xl gap-4 lg:grid-cols-[1fr_360px]">
        <Panel>
          <PanelHeader className="flex items-center justify-between">
            <div>
              <div className="font-semibold">系统配置</div>
              <div className="text-sm text-muted-foreground">JSON 编辑器覆盖当前配置，后续可继续拆成结构化表单。</div>
            </div>
            <div className="flex gap-2">
              <Button variant="outline" onClick={load}>
                <RefreshCcw className="h-4 w-4" />
                重新加载
              </Button>
              <Button onClick={save}>
                <Save className="h-4 w-4" />
                保存
              </Button>
            </div>
          </PanelHeader>
          <PanelBody>
            <Textarea className="min-h-[620px] font-mono text-xs" value={draft} onChange={(event) => setDraft(event.target.value)} />
          </PanelBody>
        </Panel>
        <Panel>
          <PanelHeader>
            <div className="font-semibold">工具目录</div>
            <div className="text-sm text-muted-foreground">{tools.length} 个工具组</div>
          </PanelHeader>
          <PanelBody className="space-y-3">
            {tools.map((tool) => (
              <div key={tool.name} className="rounded-md border p-3">
                <div className="flex items-center justify-between gap-2">
                  <div className="font-medium">{tool.display_name || tool.name}</div>
                  <Badge>{tool.category || "tool"}</Badge>
                </div>
                <div className="mt-1 text-xs text-muted-foreground">{tool.description}</div>
                {tool.included_tools?.length ? (
                  <div className="mt-2 flex flex-wrap gap-1">
                    {tool.included_tools.map((name) => (
                      <Badge key={name}>{name}</Badge>
                    ))}
                  </div>
                ) : null}
              </div>
            ))}
            {!config ? <div className="text-sm text-muted-foreground">加载中...</div> : null}
          </PanelBody>
        </Panel>
      </div>
    </div>
  );
}
