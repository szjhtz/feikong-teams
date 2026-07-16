import { describe, expect, test } from "bun:test";
import type { ChatEvent } from "@/types/events";
import { appendBufferedChatEvent, stableChatEventIdentity } from "./eventBuffer";

describe("chat event buffer", () => {
  test("uses stable protocol identities for replay deduplication", () => {
    expect(stableChatEventIdentity({ type: "system_notice", event_id: "event-1" })).toBe("event:event-1");
    expect(stableChatEventIdentity({ type: "system_notice", run_id: "run-1", sequence: 2 })).toBe("run:run-1:2");
    expect(stableChatEventIdentity({ type: "system_notice", sequence: 2 })).toBeUndefined();
  });

  test("compacts adjacent text deltas without mutating the source events", () => {
    const events: ChatEvent[] = [];
    const first: ChatEvent = {
      type: "assistant_text_delta",
      event_id: "event-1",
      message_id: "message-1",
      content: "你",
      stream_event_id: 4,
      sequence: 4,
    };
    const second: ChatEvent = {
      type: "assistant_text_delta",
      event_id: "event-2",
      message_id: "message-1",
      content: "好",
      stream_event_id: 5,
      sequence: 5,
    };

    appendBufferedChatEvent(events, first, "message-1");
    appendBufferedChatEvent(events, second, "message-1");

    expect(events).toHaveLength(1);
    expect(events[0].content).toBe("你好");
    expect(events[0].stream_event_id).toBe(5);
    expect(events[0].event_id).toBe("event-1");
    expect(first.content).toBe("你");
    expect(second.content).toBe("好");
  });

  test("preserves boundaries between messages and content blocks", () => {
    const events: ChatEvent[] = [];
    appendBufferedChatEvent(events, {
      type: "assistant_reasoning_delta",
      message_id: "message-1",
      block_id: "reasoning-1",
      content: "a",
    }, "message-1");
    appendBufferedChatEvent(events, {
      type: "assistant_reasoning_delta",
      message_id: "message-1",
      block_id: "reasoning-2",
      content: "b",
    }, "message-1");
    appendBufferedChatEvent(events, {
      type: "assistant_reasoning_delta",
      message_id: "message-2",
      block_id: "reasoning-2",
      content: "c",
    }, "message-2");

    expect(events.map((event) => event.content)).toEqual(["a", "b", "c"]);
  });
});
