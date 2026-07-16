import type { ChatEvent } from "@/types/events";

export function stableChatEventIdentity(event: ChatEvent) {
  if (event.event_id) return `event:${event.event_id}`;
  if (event.run_id && event.sequence !== undefined) return `run:${event.run_id}:${event.sequence}`;
  return undefined;
}

export function appendBufferedChatEvent(events: ChatEvent[], event: ChatEvent, messageKey: string) {
  const previous = events[events.length - 1];
  if (previous && canCompactDelta(previous, event, messageKey)) {
    mergeDelta(previous, event);
    return;
  }
  events.push({ ...event });
}

function canCompactDelta(previous: ChatEvent, event: ChatEvent, messageKey: string) {
  if (event.type !== "assistant_text_delta" && event.type !== "assistant_reasoning_delta") return false;
  if (previous.type !== event.type) return false;
  return bufferedMessageKey(previous) === messageKey
    && previous.delta_kind === event.delta_kind
    && previous.role === event.role
    && previous.block_id === event.block_id
    && previous.block_type === event.block_type
    && previous.member_call_id === event.member_call_id
    && previous.parent_tool_call_id === event.parent_tool_call_id;
}

function bufferedMessageKey(event: ChatEvent) {
  if (event.message_id) return event.message_id;
  if (event.stream_id && event.delta_kind) {
    const suffix = `:${event.delta_kind}`;
    if (event.stream_id.endsWith(suffix)) return event.stream_id.slice(0, -suffix.length);
  }
  if (event.turn_id) return `${event.turn_id}:${event.agent_name || "assistant"}`;
  if (event.run_id) return `${event.run_id}:${event.agent_name || "assistant"}`;
  return event.stream_id || event.agent_name || "assistant";
}

function mergeDelta(target: ChatEvent, source: ChatEvent) {
  if (source.content) target.content = `${target.content || ""}${source.content}`;
  if (source.reasoning_content) {
    target.reasoning_content = `${target.reasoning_content || ""}${source.reasoning_content}`;
  }
  target.stream_event_id = source.stream_event_id ?? target.stream_event_id;
  target.sequence = source.sequence ?? target.sequence;
  target.chunk_index = source.chunk_index ?? target.chunk_index;
}
