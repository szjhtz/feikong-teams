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

  useEffect(() => {
    if (activeSessionID) void dispatch(loadSessionDetail(activeSessionID));
  }, [activeSessionID, dispatch]);

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
