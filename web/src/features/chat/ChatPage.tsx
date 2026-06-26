import { useEffect } from "react";
import { useAppDispatch, useAppSelector } from "@/app/hooks";
import { chatActions } from "@/app/store";
import { loadSessionDetail } from "@/features/sessions/sessionThunks";
import { MessageList } from "./MessageList";
import { QueuePanel } from "./QueuePanel";
import { ChatInput } from "./ChatInput";

export function ChatPage() {
  const dispatch = useAppDispatch();
  const activeSessionID = useAppSelector((state) => state.chat.activeSessionID);
  const runningSessionID = useAppSelector((state) => state.chat.runningSessionID);
  const isProcessing = useAppSelector((state) => state.chat.isProcessing);

  useEffect(() => {
    if (activeSessionID && !(isProcessing && runningSessionID === activeSessionID)) {
      void dispatch(loadSessionDetail(activeSessionID));
    }
  }, [activeSessionID, runningSessionID, isProcessing, dispatch]);

  useEffect(() => {
    dispatch(chatActions.setConnectionState("connected"));
  }, [dispatch]);

  return (
    <div className="flex h-full flex-col">
      <MessageList />
      <QueuePanel />
      <ChatInput />
    </div>
  );
}
