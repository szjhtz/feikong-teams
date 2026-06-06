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
