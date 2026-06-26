import { Paperclip, Send, Square } from "lucide-react";
import { useState } from "react";
import { useAppDispatch, useAppSelector } from "@/app/hooks";
import { chatActions } from "@/app/store";
import { startStream, stopStream } from "@/api/chat";
import { subscribeStream } from "@/api/stream";
import { readJSON, storageKeys, writeJSON } from "@/lib/storage";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";

export function ChatInput() {
  const dispatch = useAppDispatch();
  const sessionID = useAppSelector((state) => state.chat.activeSessionID);
  const runningSessionID = useAppSelector((state) => state.chat.runningSessionID);
  const mode = useAppSelector((state) => state.chat.mode);
  const currentAgent = useAppSelector((state) => state.chat.currentAgent);
  const isProcessing = useAppSelector((state) => state.chat.isProcessing);
  const [value, setValue] = useState("");
  const [composing, setComposing] = useState(false);

  async function submit() {
    const message = value.trim();
    if (!message || isProcessing) return;
    setValue("");
    dispatch(chatActions.setError(undefined));
    dispatch(chatActions.appendUserMessage({ id: `user-${Date.now()}`, content: message }));
    dispatch(chatActions.setProcessing(true));
    try {
      const result = await startStream({
        session_id: sessionID || undefined,
        message,
        mode,
        agent_name: currentAgent || undefined,
      });
      dispatch(chatActions.setActiveSession(result.session_id));
      dispatch(chatActions.setRunningSession(result.session_id));
      resetOffset(result.session_id);
      void subscribe(result.session_id, 0);
    } catch (error) {
      dispatch(chatActions.setError(error instanceof Error ? error.message : String(error)));
      dispatch(chatActions.setProcessing(false));
    }
  }

  async function subscribe(id: string, initialOffset?: number) {
    const offsets = readJSON<Record<string, number>>(storageKeys.streamOffsets, {});
    const offset = initialOffset ?? offsets[id] ?? 0;
    await subscribeStream(id, offset, (event) => {
      dispatch(chatActions.receiveEvent(event));
      if (event.stream_event_id !== undefined) {
        offsets[id] = Number(event.stream_event_id) + 1;
        writeJSON(storageKeys.streamOffsets, offsets);
      }
    }).catch((error) => {
      dispatch(chatActions.setError(error instanceof Error ? error.message : String(error)));
      dispatch(chatActions.setProcessing(false));
    });
  }

  async function stop() {
    const id = runningSessionID || sessionID;
    if (!id) return;
    try {
      await stopStream(id);
    } catch (error) {
      dispatch(chatActions.setError(error instanceof Error ? error.message : String(error)));
    } finally {
      dispatch(chatActions.setProcessing(false));
    }
  }

  return (
    <div className="sketch-rule border-t bg-card/55 p-4 backdrop-blur">
      <div className="mx-auto flex max-w-5xl gap-3">
        <Button variant="outline" size="icon" aria-label="添加附件">
          <Paperclip className="h-4 w-4" />
        </Button>
        <Textarea
          value={value}
          onChange={(event) => setValue(event.target.value)}
          onCompositionStart={() => setComposing(true)}
          onCompositionEnd={() => setComposing(false)}
          onKeyDown={(event) => {
            if (event.key === "Enter" && !event.shiftKey && !composing) {
              event.preventDefault();
              void submit();
            }
          }}
          className="min-h-12 flex-1 resize-none text-base"
          placeholder="输入任务，使用 # 引用文件，@ 指定智能体。"
        />
        {isProcessing ? (
          <Button variant="destructive" onClick={stop}>
            <Square className="h-4 w-4" />
            取消
          </Button>
        ) : (
          <Button onClick={submit}>
            <Send className="h-4 w-4" />
            发送
          </Button>
        )}
      </div>
    </div>
  );
}

function resetOffset(sessionID: string) {
  const offsets = readJSON<Record<string, number>>(storageKeys.streamOffsets, {});
  delete offsets[sessionID];
  writeJSON(storageKeys.streamOffsets, offsets);
}
