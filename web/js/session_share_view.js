(function () {
  const state = {
    shareID: decodeURIComponent(location.pathname.split("/").filter(Boolean).pop() || ""),
    info: null,
  };

  const els = {
    title: document.getElementById("share-title"),
    meta: document.getElementById("share-meta"),
    content: document.getElementById("share-content"),
    passwordCard: document.getElementById("share-password-card"),
    passwordInput: document.getElementById("share-password-input"),
    passwordSubmit: document.getElementById("share-password-submit"),
    passwordError: document.getElementById("share-password-error"),
  };

  function escapeHtml(text) {
    const div = document.createElement("div");
    div.textContent = text || "";
    return div.innerHTML;
  }

  function renderMarkdown(text) {
    if (window.marked && typeof window.marked.parse === "function") {
      return window.marked.parse(text || "");
    }
    return escapeHtml(text || "").replace(/\n/g, "<br>");
  }

  function formatUnixTime(value) {
    const unix = Number(value);
    if (!Number.isFinite(unix) || unix <= 0) return "";
    return new Date(unix * 1000).toLocaleString("zh-CN", {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    });
  }

  function setError(message) {
    if (!els.content) return;
    els.passwordCard.style.display = "none";
    els.content.style.display = "";
    els.content.innerHTML = `<div class="share-empty">${escapeHtml(message)}</div>`;
  }

  function renderMeta(data) {
    const items = [];
    if (data.message_count !== undefined) items.push(`${data.message_count || 0} 条消息`);
    items.push(data.expires_at ? `有效期至 ${formatUnixTime(data.expires_at)}` : "永不过期");
    if (data.has_password !== undefined) items.push(data.has_password ? "需要密码" : "无需密码");
    items.push(data.allow_tool_details ? "包含工具详情" : "仅对话内容");
    els.meta.innerHTML = items.map((item) => `<span>${escapeHtml(item)}</span>`).join("");
  }

  async function loadInfo() {
    if (!state.shareID) {
      setError("分享链接无效");
      return;
    }
    try {
      const resp = await fetch(`/api/fkteams/public/session-shares/${encodeURIComponent(state.shareID)}/info`);
      const data = await resp.json();
      if (data.code !== 0) {
        setError(data.message === "share expired" ? "分享链接已过期" : "分享链接不存在或已失效");
        return;
      }
      state.info = data.data;
      els.title.textContent = state.info.title || "会话分享";
      renderMeta(state.info);
      if (state.info.has_password) {
        els.content.style.display = "none";
        els.passwordCard.style.display = "";
        setTimeout(() => els.passwordInput?.focus(), 50);
        return;
      }
      accessShare("");
    } catch (err) {
      console.error("load share info error:", err);
      setError("加载分享信息失败");
    }
  }

  async function accessShare(password) {
    if (els.passwordSubmit) els.passwordSubmit.disabled = true;
    if (els.passwordError) els.passwordError.textContent = "";
    try {
      const resp = await fetch(`/api/fkteams/public/session-shares/${encodeURIComponent(state.shareID)}/access`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ password: password || "" }),
      });
      const data = await resp.json();
      if (data.code !== 0) {
        if (resp.status === 401) {
          els.passwordError.textContent = "密码不正确";
          return;
        }
        setError(data.message === "share expired" ? "分享链接已过期" : "分享内容不可访问");
        return;
      }
      els.passwordCard.style.display = "none";
      els.content.style.display = "";
      els.title.textContent = data.data.title || state.info?.title || "会话分享";
      renderMeta({ ...(state.info || {}), ...data.data });
      renderMessages(data.data.messages || []);
    } catch (err) {
      console.error("access share error:", err);
      setError("加载分享内容失败");
    } finally {
      if (els.passwordSubmit) els.passwordSubmit.disabled = false;
    }
  }

  function renderMessages(messages) {
    if (!messages.length) {
      els.content.innerHTML = '<div class="share-empty">这个分享暂无会话内容</div>';
      return;
    }
    els.content.innerHTML = buildMessageBlocks(messages).map(renderBlock).join("");
  }

  function buildMessageBlocks(messages) {
    const blocks = [];
    const renderedMemberIndexes = new Set();

    for (let index = 0; index < messages.length; index++) {
      const msg = messages[index];
      if (isMemberMessage(msg)) {
        if (renderedMemberIndexes.has(index)) continue;
        const group = [];
        while (
          index < messages.length &&
          isMemberMessage(messages[index]) &&
          !renderedMemberIndexes.has(index)
        ) {
          group.push(messages[index]);
          renderedMemberIndexes.add(index);
          index++;
        }
        index--;
        blocks.push({ type: "members", messages: group });
        continue;
      }

      const inserted = agentMessageWithMemberInsert(msg, messages, renderedMemberIndexes);
      if (inserted) {
        blocks.push(...inserted);
        continue;
      }
      blocks.push({ type: "message", message: msg });
    }
    return blocks;
  }

  function agentMessageWithMemberInsert(msg, messages, renderedMemberIndexes) {
    const refs = agentToolRefs(msg);
    if (refs.ids.size === 0 && refs.names.size === 0) return null;

    const members = [];
    (messages || []).forEach((candidate, index) => {
      if (renderedMemberIndexes.has(index)) return;
      if (!isMemberMessage(candidate)) return;
      if (!memberMatchesRefs(candidate, refs)) return;
      members.push({ msg: candidate, index });
    });
    if (members.length === 0) return null;

    const events = msg?.events || [];
    let lastAgentToolIndex = -1;
    events.forEach((event, index) => {
      if (event.type !== "tool_call" || !isAgentTool(event.tool_call)) return;
      if (!toolInRefs(event.tool_call, refs)) return;
      lastAgentToolIndex = index;
    });
    if (lastAgentToolIndex < 0) return null;

    members.forEach((item) => renderedMemberIndexes.add(item.index));
    const blocks = [];
    const before = events.slice(0, lastAgentToolIndex + 1);
    const after = events.slice(lastAgentToolIndex + 1);
    if (before.length > 0) blocks.push({ type: "message", message: { ...msg, events: before } });
    blocks.push({ type: "members", messages: members.map((item) => item.msg) });
    if (after.length > 0) blocks.push({ type: "message", message: { ...msg, events: after } });
    return blocks;
  }

  function renderBlock(block) {
    if (block.type === "members") return renderMemberGroup(block.messages);
    return renderMessage(block.message);
  }

  function isLegacyMemberMessage(msg) {
    const name = msg?.agent_name || "";
    return /^ask_fkagent_[A-Za-z0-9_-]+$/.test(name);
  }

  function isMemberMessage(msg) {
    return !!(
      msg &&
      (msg.is_member_event ||
        msg.member_call_id ||
        msg.member_tool_name ||
        msg.member_name ||
        isLegacyMemberMessage(msg))
    );
  }

  function isAgentTool(tool) {
    if (!tool) return false;
    const kind = String(tool.kind || "").toLowerCase();
    const name = tool.name || "";
    return kind === "agent" || /^ask_fkagent_[A-Za-z0-9_-]+$/.test(name);
  }

  function agentToolRefs(msg) {
    const refs = { ids: new Set(), names: new Set() };
    (msg?.events || []).forEach((event) => {
      if (event.type !== "tool_call" || !isAgentTool(event.tool_call)) return;
      if (event.tool_call.id) refs.ids.add(event.tool_call.id);
      if (event.tool_call.name) refs.names.add(event.tool_call.name);
    });
    return refs;
  }

  function toolInRefs(tool, refs) {
    if (!tool || !refs) return false;
    return !!(
      (tool.id && refs.ids.has(tool.id)) ||
      (tool.name && refs.names.has(tool.name))
    );
  }

  function memberMatchesRefs(msg, refs) {
    if (!msg || !refs) return false;
    return !!(
      (msg.member_call_id && refs.ids.has(msg.member_call_id)) ||
      (msg.member_tool_name && refs.names.has(msg.member_tool_name)) ||
      (msg.agent_name && refs.names.has(msg.agent_name))
    );
  }

  function memberLabel(msg) {
    if (msg.member_name) return msg.member_name;
    const raw = isLegacyMemberMessage(msg)
      ? (msg.agent_name || "").slice(4)
      : msg.agent_name || "Member";
    return raw
      .split(/[_-]+/)
      .filter(Boolean)
      .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
      .join(" ") || "Member";
  }

  function memberKey(msg, index) {
    return msg.member_call_id || msg.member_tool_name || memberLabel(msg) || `member-${index}`;
  }

  function renderMemberGroup(messages) {
    const cards = buildMemberCards(messages);
    const title = cards.length > 1 ? "成员并行任务" : "成员任务";
    return `
      <section class="share-member-panel">
        <div class="share-member-header">
          <span>${escapeHtml(title)}</span>
          <span>${cards.length} 个成员</span>
        </div>
        <div class="share-member-list">
          ${cards.map(renderMemberCard).join("")}
        </div>
      </section>
    `;
  }

  function buildMemberCards(messages) {
    const map = new Map();
    messages.forEach((msg, index) => {
      const key = memberKey(msg, index);
      let card = map.get(key);
      if (!card) {
        card = {
          key,
          label: memberLabel(msg),
          messages: [],
          hasError: false,
        };
        map.set(key, card);
      }
      card.messages.push(msg);
      if ((msg.events || []).some((event) => event.type === "error")) {
        card.hasError = true;
      }
    });
    return Array.from(map.values());
  }

  function renderMemberCard(card) {
    const summary = memberSummary(card.messages);
    return `
      <details class="share-member-card">
        <summary>
          <span class="share-member-dot ${card.hasError ? "error" : ""}"></span>
          <span class="share-member-name">${escapeHtml(card.label)}</span>
          <span class="share-member-summary">${escapeHtml(summary)}</span>
          <span class="share-member-status ${card.hasError ? "error" : ""}">${card.hasError ? "失败" : "完成"}</span>
        </summary>
        <div class="share-member-detail">
          ${card.messages.map(renderMemberMessage).join("")}
        </div>
      </details>
    `;
  }

  function memberSummary(messages) {
    for (const msg of messages) {
      for (const event of msg.events || []) {
        if ((event.type === "text" || event.type === "reasoning") && event.content) {
          return compactText(event.content, 72);
        }
        if (event.type === "tool_call" && event.tool_call) {
          return event.tool_call.display_name || event.tool_call.name || "工具调用";
        }
        if (event.type === "action" && event.action?.content) {
          return compactText(event.action.content, 72);
        }
      }
    }
    return `${messages.length} 条成员消息`;
  }

  function compactText(text, maxLen) {
    const value = String(text || "").replace(/\s+/g, " ").trim();
    const chars = Array.from(value);
    if (chars.length <= maxLen) return value;
    return chars.slice(0, maxLen).join("") + "...";
  }

  function renderMemberMessage(msg) {
    const time = msg.start_time ? new Date(msg.start_time).toLocaleString("zh-CN") : "";
    const events = (msg.events || []).map(renderEvent).join("");
    return `
      <div class="share-member-message">
        ${time ? `<div class="share-member-time">${escapeHtml(time)}</div>` : ""}
        ${events || '<div class="share-event">无内容</div>'}
      </div>
    `;
  }

  function renderMessage(msg) {
    const agent = msg.member_name || msg.agent_name || "成员";
    const time = msg.start_time ? new Date(msg.start_time).toLocaleString("zh-CN") : "";
    const events = (msg.events || []).map(renderEvent).join("");
    const roleClass = msg.agent_name === "用户" ? " user" : "";
    return `
      <article class="share-message${roleClass}">
        <div class="share-message-head">
          <span class="share-agent">${escapeHtml(agent)}</span>
          <span class="share-time">${escapeHtml(time)}</span>
        </div>
        ${events || '<div class="share-event">无内容</div>'}
      </article>
    `;
  }

  function renderEvent(event) {
    if (!event) return "";
    if (event.type === "text") {
      return `<div class="share-event markdown-body">${renderMarkdown(event.content || "")}</div>`;
    }
    if (event.type === "reasoning") {
      return `<div class="share-event reasoning">${escapeHtml(event.content || "")}</div>`;
    }
    if (event.type === "tool_call" && event.tool_call) {
      const tool = event.tool_call;
      const name = tool.display_name || tool.name || "工具调用";
      const args = tool.arguments ? `<pre>${escapeHtml(tool.arguments)}</pre>` : "";
      const result = tool.result ? `<pre>${escapeHtml(tool.result)}</pre>` : "";
      return `<div class="share-event tool"><strong>${escapeHtml(name)}</strong>${args}${result}</div>`;
    }
    if (event.type === "action" && event.action) {
      const action = event.action.action_type ? `[${event.action.action_type}] ` : "";
      return `<div class="share-event action">${escapeHtml(action + (event.action.content || ""))}</div>`;
    }
    if (event.type === "error") {
      return `<div class="share-event error">${escapeHtml(event.content || "执行失败")}</div>`;
    }
    return "";
  }

  if (els.passwordSubmit) {
    els.passwordSubmit.addEventListener("click", () => accessShare(els.passwordInput.value));
  }
  if (els.passwordInput) {
    els.passwordInput.addEventListener("keydown", (e) => {
      if (e.key === "Enter") accessShare(els.passwordInput.value);
    });
  }

  loadInfo();
})();
