import type { ChatEvent } from "@/types/events";

export const maxFallbackEventKeys = 2048;
const fallbackEventKeyPruneCount = 512;

export interface ChatEventDedupState {
  seenEventKeys: Record<string, true>;
  seenEventKeyOrder: string[];
  seenStreamEventID?: number;
}

export function stableChatEventIdentity(event: ChatEvent) {
  if (event.event_id) return `event:${event.event_id}`;
  if (event.run_id && event.sequence !== undefined) return `run:${event.run_id}:${event.sequence}`;
  return undefined;
}

// rememberChatEvent 使用流序号水位线去重，避免为每个增量事件长期保存字符串键。
export function rememberChatEvent(state: ChatEventDedupState, event: ChatEvent) {
  const identity = stableChatEventIdentity(event);
  const streamEventID = event.stream_event_id;
  if (typeof streamEventID === "number" && Number.isSafeInteger(streamEventID) && streamEventID >= 0) {
    if (identity && state.seenEventKeys[identity]) return false;
    if (state.seenStreamEventID !== undefined && streamEventID <= state.seenStreamEventID) return false;
    state.seenStreamEventID = streamEventID;
    return true;
  }
  if (!identity) return true;
  if (state.seenEventKeys[identity]) return false;
  state.seenEventKeys[identity] = true;
  state.seenEventKeyOrder.push(identity);
  if (state.seenEventKeyOrder.length > maxFallbackEventKeys) {
    const expired = state.seenEventKeyOrder.splice(0, fallbackEventKeyPruneCount);
    for (const key of expired) delete state.seenEventKeys[key];
  }
  return true;
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
