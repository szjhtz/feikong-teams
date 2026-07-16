# 会话分享 API

会话分享用于把已保存的聊天历史发布为公开链接。管理接口需要普通 Web/API 认证；公开访问接口用于分享页读取内容。

## POST /api/fkteams/session-shares

创建会话分享。

**请求 Body**：

```json
{
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "password": "optional",
  "expires_in": 604800,
  "allow_tool_details": false
}
```

| 字段 | 类型 | 必填 | 说明 |
| ---- | ---- | ---- | ---- |
| `session_id` | string | 是 | 会话 ID |
| `password` | string | 否 | 访问密码 |
| `expires_in` | int | 否 | 过期秒数，默认 7 天，最长 90 天；小于 0 表示永不过期 |
| `allow_tool_details` | bool | 否 | 是否在分享内容中保留工具参数和结果详情 |

**成功响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "share-id",
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "会话标题",
    "has_password": true,
    "allow_tool_details": false,
    "message_count": 8,
    "expires_at": 1760000000,
    "created_at": 1750000000
  }
}
```

**失败响应**：

| 状态码 | message | 说明 |
| ------ | ------- | ---- |
| 400 | `invalid request body` | 请求体解析失败 |
| 400 | `invalid session ID` | 会话 ID 不合法 |
| 400 | `session has no shareable messages` | 会话没有可分享消息 |
| 404 | `session history not found` | 会话历史不存在 |
| 500 | `failed to read session history` | 读取历史失败 |
| 500 | `failed to create share` | 生成分享 ID 失败 |
| 500 | `failed to process password` | 密码处理失败 |

---

## GET /api/fkteams/session-shares

列出未过期的会话分享，按创建时间倒序。

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": "share-id",
      "session_id": "session-id",
      "title": "会话标题",
      "has_password": false,
      "allow_tool_details": false,
      "message_count": 8,
      "expires_at": 1760000000,
      "created_at": 1750000000,
      "last_accessed_at": 1750000100
    }
  ]
}
```

---

## DELETE /api/fkteams/session-shares/:shareID

删除会话分享。

**成功响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "share deleted"
  }
}
```

**失败响应**：

| 状态码 | message | 说明 |
| ------ | ------- | ---- |
| 400 | `missing share ID` | 缺少分享 ID |
| 404 | `share not found` | 分享不存在 |

---

## GET /api/fkteams/public/session-shares/:shareID/info

获取公开分享基础信息，不返回消息内容。

**成功响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "share-id",
    "title": "会话标题",
    "has_password": true,
    "message_count": 8,
    "expires_at": 1760000000,
    "created_at": 1750000000,
    "allow_tool_details": false
  }
}
```

**失败响应**：

| 状态码 | message | 说明 |
| ------ | ------- | ---- |
| 404 | `share not found` | 分享不存在 |
| 410 | `share expired` | 分享已过期 |

---

## POST /api/fkteams/public/session-shares/:shareID/access

访问公开分享内容。无密码分享可以提交空 JSON；有密码分享需提交密码。

**请求 Body**：

```json
{
  "password": "optional"
}
```

**成功响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "share-id",
    "title": "会话标题",
    "messages": [],
    "message_count": 8,
    "expires_at": 1760000000,
    "created_at": 1750000000,
    "allow_tool_details": false
  }
}
```

**密码错误或未提供密码**：

```json
{
  "code": 1,
  "message": "password required",
  "data": {
    "require_password": true
  }
}
```

**失败响应**：

| 状态码 | message | 说明 |
| ------ | ------- | ---- |
| 401 | `password required` | 未提供密码 |
| 401 | `invalid password` | 密码错误 |
| 404 | `share not found` | 分享不存在 |
| 410 | `share expired` | 分享已过期 |
| 410 | `shared session unavailable` | 分享对应历史不可用 |
| 429 | `too many authentication attempts` | 尝试次数过多，响应包含 `Retry-After` |
