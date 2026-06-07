# 流式任务 API

流式任务 API 将任务执行与前端连接解耦。任务在服务端后台独立运行，所有输出事件缓存在内存中。前端通过 SSE（Server-Sent Events）订阅事件流，断线重连后可从断点继续接收，实现无损续接。

## 核心流程

```
1. POST /stream/start          → 启动后台任务，返回 session_id
2. POST /stream/steer          → 向运行中的任务注入转向消息
3. GET  /stream/queue/:id      → 查询未执行队列
4. PATCH/DELETE/POST move      → 修改、删除、排序未执行队列项
5. GET  /stream/subscribe/:id  → SSE 订阅事件流（支持断线重连）
6. POST /stream/stop/:id       → 停止正在运行的任务
7. GET  /stream/status/:id     → 查询任务状态
8. GET  /stream/events/:id     → 一次性拉取已缓冲事件
9. POST /stream/approval       → 提交 HITL 审批决定
10. POST /stream/ask-response  → 提交交互式提问的回答
```

## 接口详情

### 启动任务

```
POST /api/fkteams/stream/start
```

**请求体**：

```json
{
  "session_id": "可选，不提供则自动生成 UUID",
  "message": "用户消息",
  "mode": "team",
  "agent_name": "可选，指定单个智能体",
  "contents": []
}
```

| 字段         | 类型   | 必填 | 说明                                                |
| ------------ | ------ | ---- | --------------------------------------------------- |
| `session_id` | string | 否   | 会话 ID，不提供则自动生成                           |
| `message`    | string | 条件 | 文本消息（与 `contents` 二选一）                    |
| `mode`       | string | 否   | 工作模式：`team`/`roundtable`/`custom`/`deep`，兼容旧值 `supervisor` |
| `agent_name` | string | 否   | 指定单个智能体名称                                  |
| `contents`   | array  | 条件 | 多模态内容（与 `message` 二选一），格式同聊天 API   |

**成功响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "session_id": "abc-123",
    "status": "processing",
    "message": "task started"
  }
}
```

如果同一 `session_id` 已有运行中的任务，`/stream/start` 会把消息作为 follow-up 排队；当前 Agent 正常停止后继续处理该消息，不会取消当前任务。

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "session_id": "abc-123",
    "status": "queued",
    "message": "message queued",
    "queue_kind": "follow_up",
    "queued_count": 1
  }
}
```

**错误码**：

| 状态码 | 说明            |
| ------ | --------------- |
| 400    | 参数错误        |
| 500    | Runner 创建失败 |

---

### 转向任务

```
POST /api/fkteams/stream/steer
```

向运行中的任务发送 steering 消息。消息会在当前模型输出结束、工具调用完成后，于下一次模型调用前注入上下文；不会中断正在输出的 token，也不会强杀正在执行的工具。

**请求体**：

```json
{
  "session_id": "abc-123",
  "message": "停止当前方向，优先检查这个问题",
  "contents": []
}
```

**成功响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "session_id": "abc-123",
    "status": "queued",
    "message": "steering queued",
    "queue_kind": "steering",
    "queued_count": 1
  }
}
```

| 状态码 | 说明                         |
| ------ | ---------------------------- |
| 400    | 参数错误                     |
| 404    | 该会话没有正在运行的流式任务 |

---

### 管理未执行队列

运行中的 `follow_up` 与 `steering` 都会进入未执行队列。每个队列项包含稳定 `id`、`kind`、`text`、`display_text`、`created_at`，前端可在尚未消费前编辑、删除、调整顺序或切换 `kind`。排序只在同类队列内生效：`steering` 仍由 `SteeringSource` 在下一次模型调用前消费，`follow_up` 在当前任务完成后继续执行。

```
GET /api/fkteams/stream/queue/:sessionID
PATCH /api/fkteams/stream/queue/:sessionID/:queueID
DELETE /api/fkteams/stream/queue/:sessionID/:queueID
POST /api/fkteams/stream/queue/:sessionID/:queueID/kind
POST /api/fkteams/stream/queue/:sessionID/:queueID/move
```

编辑请求体：

```json
{
  "message": "新的队列内容",
  "contents": []
}
```

移动请求体：

```json
{
  "direction": "up"
}
```

切换类型请求体：

```json
{
  "kind": "steering"
}
```

成功响应会返回最新 `queue` 快照；服务端也会推送 `queue_updated` 事件。

---

### 订阅事件流

```
GET /api/fkteams/stream/subscribe/:sessionID
```

SSE 长连接，持续推送事件直到任务完成或客户端断开。

**断线重连**：

- **方式一**：浏览器 `EventSource` 自动携带 `Last-Event-ID` 请求头
- **方式二**：手动指定 `?offset=N` query 参数

每个 SSE 事件格式：

```
id: 42
data: {"type":"stream_chunk","agent_name":"coder","content":"...","session_id":"abc-123"}
```

**事件类型**：

| type                | 说明         |
| ------------------- | ------------ |
| `processing_start`  | 任务开始     |
| `user_message`      | 用户消息；运行中排队时包含 `queued` / `queue_id` / `queue_kind` / `queued_count` |
| `queue_updated`     | 未执行队列快照，包含 `queue` / `queued_count` |
| `stream_chunk`      | 文本片段     |
| `reasoning_chunk`   | 推理内容片段 |
| `tool_calls`        | 工具调用     |
| `tool_result`       | 工具结果     |
| `action`            | 动作事件     |
| `approval_required` | 需要审批     |
| `error`             | 错误         |
| `cancelled`         | 任务已取消   |
| `processing_end`    | 任务完成     |

**前端示例**：

```javascript
const es = new EventSource("/api/fkteams/stream/subscribe/abc-123");

es.onmessage = (e) => {
  const event = JSON.parse(e.data);
  switch (event.type) {
    case "stream_chunk":
      appendText(event.content);
      break;
    case "processing_end":
      es.close();
      break;
  }
};

// 断线后浏览器会自动重连并携带 Last-Event-ID，
// 服务端从断点继续推送，无需额外处理。
```

---

### 停止任务

```
POST /api/fkteams/stream/stop/:sessionID
```

无请求体。

**成功响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "session_id": "abc-123",
    "message": "task stop requested"
  }
}
```

| 状态码 | 说明               |
| ------ | ------------------ |
| 404    | 未找到该会话的任务 |
| 409    | 任务未在运行状态   |

---

### 查询任务状态

```
GET /api/fkteams/stream/status/:sessionID
```

**有活跃任务时**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "session_id": "abc-123",
    "status": "processing",
    "has_task": true,
    "mode": "team",
    "agent_name": "",
    "event_count": 42,
    "created_at": "2026-04-05T10:00:00Z"
  }
}
```

**无活跃任务但有会话记录时**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "session_id": "abc-123",
    "status": "completed",
    "has_task": false,
    "title": "用户问题标题",
    "created_at": "2026-04-05T10:00:00Z",
    "updated_at": "2026-04-05T10:05:00Z"
  }
}
```

`status` 取值：`processing` / `completed` / `error` / `cancelled` / `idle`

---

### 拉取已缓冲事件

```
GET /api/fkteams/stream/events/:sessionID?offset=0
```

一次性返回已缓冲的事件（非 SSE），适用于页面加载时快速获取历史。

| 参数     | 说明               |
| -------- | ------------------ |
| `offset` | 起始位置（默认 0） |

**响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "session_id": "abc-123",
    "status": "processing",
    "events": [
      { "id": 0, "data": { "type": "processing_start", "...": "..." } },
      { "id": 1, "data": { "type": "stream_chunk", "...": "..." } }
    ],
    "event_count": 42,
    "done": false
  }
}
```

---

### 提交审批

```
POST /api/fkteams/stream/approval
```

**请求体**：

```json
{
  "session_id": "abc-123",
  "decision": 1
}
```

| decision | 含义         |
| -------- | ------------ |
| 0        | 拒绝         |
| 1        | 允许（一次） |
| 2        | 允许（该项） |
| 3        | 全部允许     |

**成功响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "approval submitted"
  }
}
```

| 状态码 | 说明                 |
| ------ | -------------------- |
| 404    | 无运行中的任务       |
| 409    | 当前没有待审批的请求 |

---

### 提交交互式提问回答

```
POST /api/fkteams/stream/ask-response
```

当智能体通过 `ask_questions` 工具向用户提问时，前端通过此接口提交用户的回答。

**请求体**：

```json
{
  "session_id": "abc-123",
  "selected": ["选项1", "选项2"],
  "free_text": "用户的自由输入文本"
}
```

| 字段         | 类型     | 必填 | 说明               |
| ------------ | -------- | ---- | ------------------ |
| `session_id` | string   | ✓    | 会话 ID            |
| `selected`   | string[] | 否   | 用户选择的选项列表 |
| `free_text`  | string   | 否   | 用户的自由文本输入 |

**成功响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "response submitted"
  }
}
```

| 状态码 | 说明                     |
| ------ | ------------------------ |
| 404    | 无运行中的任务           |
| 409    | 当前没有待回答的提问请求 |

---

## 缓存机制

流式事件缓存**仅服务于运行中及刚完成的任务**：

| 状态                      | 数据来源 | 接口                           |
| ------------------------- | -------- | ------------------------------ |
| 任务运行中                | 内存缓存 | `/stream/subscribe` SSE 实时流 |
| 任务刚完成（5 分钟内）    | 内存缓存 | `/stream/events` 拉取缓冲      |
| 任务已完成（超过 5 分钟） | 历史文件 | `/sessions/:id` 加载历史       |

- 缓存在任务完成 5 分钟后自动释放内存
- 已完成任务的完整数据已通过历史系统持久化，前端应使用会话接口加载
- 同一 session 启动新任务时，旧缓存自动替换

## 典型使用场景

### 页面首次打开（多轮会话场景）

1. 调用 `GET /sessions/:id` 加载完整历史（所有已完成的轮次）并渲染
2. 调用 `GET /stream/status/:id` 检查是否有运行中的任务
3. 如果 `has_task=true && status=processing`：
   - 调用 `GET /stream/events/:id?offset=0` 获取当前轮次的已缓冲事件并渲染
   - 然后用 `GET /stream/subscribe/:id?offset=<last_id+1>` 接入实时流
4. 如果 `has_task=false`，仅展示历史，等待用户输入

### 断线重连

浏览器 `EventSource` 自动处理：断线时浏览器自动重连并发送 `Last-Event-ID`，服务端从下一个事件开始推送。前端无需额外代码。

### 退出页面后重新进入

1. 调用 `GET /sessions/:id` 加载完整会话历史 → 渲染已完成的所有轮次
2. 调用 `GET /stream/status/:id` → 若 `has_task=true, status=processing`：
   - 调用 `GET /stream/events/:id` 拉取当前轮次缓冲事件 → 渲染
   - 调用 `GET /stream/subscribe/:id?offset=<last_event_id+1>` → 续接实时流
3. 若 `has_task=false`，仅展示历史
