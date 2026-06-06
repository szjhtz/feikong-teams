const assert = require("node:assert/strict");
const test = require("node:test");

global.FKTeamsChat = function () {};
require("../js/messages.js");

function newChatWithRecordedMigrations() {
  const chat = Object.create(FKTeamsChat.prototype);
  chat.migrations = [];
  chat.migrateMemberToolFlow = function (_entry, fromKey, toKey) {
    if (fromKey && toKey && fromKey !== toKey) this.migrations.push([fromKey, toKey]);
  };
  return chat;
}

test("member tool flow key resolves idx and id aliases to final ref key", () => {
  const chat = newChatWithRecordedMigrations();
  const entry = { toolFlowKeyByName: { member_echo: "idx:0" } };

  const key = chat.resolveMemberToolFlowKey(
    entry,
    {
      tool_call_ref: "tool|member|idx:0",
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

  assert.equal(key, "ref:tool|member|idx:0");
  assert.deepEqual(chat.migrations, [
    ["id:member-tool-call", "ref:tool|member|idx:0"],
    ["idx:0", "ref:tool|member|idx:0"],
    ["fallback:0", "ref:tool|member|idx:0"],
  ]);
});

test("member tool flow key resolves index-only event without inventing another card", () => {
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

  assert.equal(key, "idx:0");
  assert.deepEqual(chat.migrations, [["fallback:0", "idx:0"]]);
});

test("tool call normalization merges top-level identity into array calls", () => {
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
    ref: "ref-1",
    index: 2,
    name: "member_echo",
    display_name: "Echo",
    kind: "tool",
    target: "",
    arguments: "{\"text\":\"hello\"}",
  });
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
