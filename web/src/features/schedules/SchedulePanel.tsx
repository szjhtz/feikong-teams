import { Ban, RefreshCcw } from "lucide-react";
import { useEffect, useState } from "react";
import { cancelSchedule, getScheduleHistory, getScheduleResult, listSchedules } from "@/api/schedules";
import { schedulesActions } from "@/app/store";
import { useAppDispatch, useAppSelector } from "@/app/hooks";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Panel, PanelBody, PanelHeader } from "@/components/ui/panel";
import { renderMarkdown } from "@/lib/markdown";
import { formatTime, shortID } from "@/lib/format";

export function SchedulePanel() {
  const dispatch = useAppDispatch();
  const tasks = useAppSelector((state) => state.schedules.items);
  const [selected, setSelected] = useState<string>("");
  const [detail, setDetail] = useState("");

  async function load() {
    const result = await listSchedules();
    dispatch(schedulesActions.setSchedules(result.tasks || []));
  }

  async function showDetail(id: string) {
    setSelected(id);
    const [result, history] = await Promise.all([getScheduleResult(id).catch(() => undefined), getScheduleHistory(id).catch(() => undefined)]);
    setDetail([result?.result || result?.content || "", JSON.stringify(history?.entries || [], null, 2)].filter(Boolean).join("\n\n"));
  }

  useEffect(() => {
    void load();
  }, []);

  return (
    <div className="grid h-full grid-cols-[420px_1fr] overflow-hidden">
      <div className="border-r p-4">
        <div className="mb-3 flex items-center justify-between">
          <div className="font-semibold">定时任务</div>
          <Button variant="outline" size="sm" onClick={load}>
            <RefreshCcw className="h-4 w-4" />
            刷新
          </Button>
        </div>
        <div className="space-y-2">
          {tasks.map((task) => (
            <button key={task.id} className="w-full rounded-md border p-3 text-left hover:bg-accent" onClick={() => showDetail(task.id)}>
              <div className="flex items-center justify-between gap-2">
                <span className="font-medium">{shortID(task.id)}</span>
                <Badge>{task.status}</Badge>
              </div>
              <div className="mt-2 line-clamp-2 text-sm text-muted-foreground">{task.task}</div>
              <div className="mt-2 text-xs text-muted-foreground">{task.cron_expr || "once"} · {formatTime(task.next_run_at)}</div>
            </button>
          ))}
        </div>
      </div>
      <div className="overflow-auto p-6">
        {selected ? (
          <Panel>
            <PanelHeader className="flex items-center justify-between">
              <div>任务详情：{shortID(selected)}</div>
              <Button variant="destructive" size="sm" onClick={() => cancelSchedule(selected).then(load)}>
                <Ban className="h-4 w-4" />
                取消
              </Button>
            </PanelHeader>
            <PanelBody>
              <div className="prose prose-sm max-w-none" dangerouslySetInnerHTML={{ __html: renderMarkdown(detail || "暂无结果") }} />
            </PanelBody>
          </Panel>
        ) : (
          <div className="text-sm text-muted-foreground">选择一个任务查看结果与历史。</div>
        )}
      </div>
    </div>
  );
}
