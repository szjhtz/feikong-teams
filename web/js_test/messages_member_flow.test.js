const assert = require("node:assert/strict");
const test = require("node:test");

global.FKTeamsChat = function () {};
require("../js/messages.js");

function newChatWithRecordedMigrations() {
  const chat = Object.create(FKTeamsChat.prototype);
  return chat;
}

test("member tool flow key uses only canonical ref", () => {
  const chat = newChatWithRecordedMigrations();
  const entry = {};

  const key = chat.resolveMemberToolFlowKey(
    entry,
    {
      tool_call_ref: "tool_call:member-tool-call",
      tool_call_id: "member-tool-call",
      tool_call_index: 0,
      tool_name: "member_echo",
    },
    {
      id: "member-tool-call",
      index: 0,
      name: "member_echo",
    },
    0,
  );

  assert.equal(key, "ref:tool_call:member-tool-call");
});

test("member tool flow key rejects index-only event", () => {
  const chat = newChatWithRecordedMigrations();
  const entry = {};

  const key = chat.resolveMemberToolFlowKey(
    entry,
    {
      tool_call_index: 0,
      tool_name: "member_echo",
    },
    null,
    0,
  );

  assert.equal(key, "");
});

test("tool call normalization does not spread top-level identity into array calls", () => {
  const chat = newChatWithRecordedMigrations();

  const calls = chat.normalizeToolCallsForEvent({
    tool_call_id: "call-1",
    tool_call_ref: "ref-1",
    tool_call_index: 2,
    tool_name: "member_echo",
    tool_display_name: "Echo",
    tool_kind: "tool",
    tool_calls: [{
      id: "call-1",
      index: 2,
      name: "member_echo",
      arguments: "{\"text\":\"hello\"}",
    }],
  });

  assert.equal(calls.length, 1);
  assert.deepEqual(calls[0], {
    id: "call-1",
    ref: "",
    index: 2,
    name: "member_echo",
    display_name: "Echo",
    kind: "tool",
    target: "",
    arguments: "{\"text\":\"hello\"}",
  });
});

test("tool key rejects non-canonical span identity", () => {
  const chat = newChatWithRecordedMigrations();

  const calls = chat.normalizeToolCallsForEvent({
    type: "tool_update",
    span_id: "span-tool-1",
    tool_name: "member_echo",
    delta_kind: "tool_result",
    delta: "chunk",
  });

  assert.equal(calls.length, 1);
  assert.equal(chat.toolCallKey(calls[0], 0), "");
});

test("member tool flow rejects id-only event even when name alias exists", () => {
  const chat = newChatWithRecordedMigrations();
  const entry = {};

  const key = chat.resolveMemberToolFlowKey(
    entry,
    {
      tool_call_id: "member-tool-call",
      tool_name: "member_echo",
    },
    {
      id: "member-tool-call",
      name: "member_echo",
    },
    0,
  );

  assert.equal(key, "");
});

test("role tool message delta is routed as tool result update", () => {
  const chat = Object.create(FKTeamsChat.prototype);
  let routed = null;
  chat.handleCoreToolUpdate = (event) => {
    routed = event;
  };

  chat.handleCoreMessageDelta({
    role: "tool",
    tool_call_id: "call-1",
    tool_call_ref: "tool_call:call-1",
    tool_name: "member_echo",
    content: "stream chunk",
  });

  assert.equal(routed.tool_call_id, "call-1");
  assert.equal(routed.tool_call_ref, "tool_call:call-1");
  assert.equal(routed.delta_kind, "tool_result");
  assert.equal(routed.content, "stream chunk");
});

test("role tool message end is routed as tool result end", () => {
  const chat = Object.create(FKTeamsChat.prototype);
  let routed = null;
  chat.handleCoreToolEnd = (event) => {
    routed = event;
  };

  chat.handleCoreMessageEnd({
    role: "tool",
    tool_call_id: "call-1",
    tool_call_ref: "tool_call:call-1",
    tool_name: "member_echo",
    content: "done",
  });

  assert.equal(routed.tool_call_id, "call-1");
  assert.equal(routed.tool_call_ref, "tool_call:call-1");
  assert.equal(routed.content, "done");
});

test("dispatch task handling does not assume the first tool call", () => {
  const chat = Object.create(FKTeamsChat.prototype);
  chat.isMemberRunEvent = () => false;
  chat.messagesContainer = { querySelectorAll: () => [] };
  chat.scrollToBottom = () => {};

  chat.handleToolCalls({
    tool_calls: [
      { name: "other_tool", arguments: "{\"x\":1}" },
      { name: "dispatch_tasks", arguments: "{\"tasks\":[{\"title\":\"task\"}]}" },
    ],
  });

  assert.deepEqual(chat._pendingDispatchTasks, [{ title: "task" }]);
});
