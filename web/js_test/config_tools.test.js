const assert = require("node:assert/strict");
const test = require("node:test");

global.FKTeamsChat = function () {};
require("../js/config.js");

test("tool catalog normalizes rich and legacy responses", () => {
  const chat = Object.create(FKTeamsChat.prototype);
  const catalog = chat.normalizeToolCatalog([
    "file",
    {
      name: "search",
      display_name: "网络搜索",
      description: "检索互联网信息",
      category: "研究",
      builtin: true,
      included_tools: ["search"],
    },
  ]);

  assert.deepEqual(catalog.map((tool) => tool.name), ["file", "search"]);
  assert.equal(catalog[0].display_name, "file");
  assert.equal(catalog[0].builtin, true);
  assert.equal(catalog[1].display_name, "网络搜索");
  assert.equal(catalog[1].description, "检索互联网信息");
});

test("agent tool catalog merges enabled mcp servers without duplicates", () => {
  const chat = Object.create(FKTeamsChat.prototype);
  chat._toolCatalog = chat.normalizeToolCatalog([
    { name: "file", description: "文件工具" },
    { name: "mcp-demo", description: "来自接口" },
  ]);
  chat._configData = {
    custom: {
      mcp_servers: [
        { name: "demo", desc: "来自配置", enabled: true },
        { name: "disabled", desc: "禁用服务", enabled: false },
        { name: "fresh", desc: "新服务", enabled: true },
      ],
    },
  };

  const names = chat.getAgentToolCatalog().map((tool) => tool.name);

  assert.deepEqual(names, ["file", "mcp-demo", "mcp-fresh"]);
});
