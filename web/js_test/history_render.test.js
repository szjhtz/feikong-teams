const assert = require("node:assert/strict");
const test = require("node:test");

global.FKTeamsChat = function () {};
require("../js/history.js");

function fakeAssistantMessage(body) {
  return {
    querySelector(selector) {
      if (selector === ".message-body") return body;
      return null;
    },
  };
}

function fakeMessageBody() {
  return {
    html: "",
    appendChild() {},
    prepend() {},
    querySelector() {
      return null;
    },
    setAttribute() {},
    set innerHTML(value) {
      this.html = value;
    },
    get innerHTML() {
      return this.html;
    },
  };
}

test("history action splits assistant text timeline", () => {
  const chat = Object.create(FKTeamsChat.prototype);
  const bodies = [];
  const timeline = [];

  chat.createAssistantMessage = () => {
    const body = fakeMessageBody();
    bodies.push(body);
    timeline.push({ type: "message", body });
    return fakeAssistantMessage(body);
  };
  chat.renderSingleAction = (action) => {
    timeline.push({ type: "action", action });
  };
  chat.renderMarkdown = (content) => content;

  chat.renderHistoryAgentMessage({
    agent_name: "coordinator",
    events: [
      { type: "text", content: "before" },
      {
        type: "action",
        action: { action_type: "approval_required", content: "approval" },
      },
      { type: "text", content: "after" },
    ],
  });

  assert.equal(bodies.length, 2);
  assert.deepEqual(timeline.map((item) => item.type), [
    "message",
    "action",
    "message",
  ]);
  assert.equal(bodies[0].html, "before");
  assert.equal(bodies[1].html, "after");
});

test("sidebar history shows loading before debounced fetch", () => {
  const chat = Object.create(FKTeamsChat.prototype);
  let debounceCalled = false;
  const classes = new Set();
  chat.sidebarSessionList = {
    innerHTML: '<div class="sidebar-session-empty">暂无会话记录</div>',
    classList: {
      add(name) { classes.add(name); },
      remove(name) { classes.delete(name); },
      contains(name) { return classes.has(name); },
    },
  };
  chat.debounce = () => {
    debounceCalled = true;
  };

  chat.loadSidebarHistory();

  assert.equal(debounceCalled, true);
  assert.equal(chat.sidebarSessionList.classList.contains("loading"), true);
  assert.match(chat.sidebarSessionList.innerHTML, /sidebar-session-loading/);
  assert.match(chat.sidebarSessionList.innerHTML, /加载中/);
});

test("sidebar session render clears loading layout", () => {
  const chat = Object.create(FKTeamsChat.prototype);
  const classes = new Set(["loading"]);
  chat.sidebarSessionList = {
    innerHTML: "",
    appendChild() {},
    querySelectorAll() { return []; },
    classList: {
      add(name) { classes.add(name); },
      remove(name) { classes.delete(name); },
      contains(name) { return classes.has(name); },
    },
  };
  chat._sidebarMenuOutsideBound = true;
  chat.escapeHtml = (value) => String(value || "");
  chat.formatTime = () => "";

  chat.renderSidebarSessions([]);

  assert.equal(chat.sidebarSessionList.classList.contains("loading"), false);
});
