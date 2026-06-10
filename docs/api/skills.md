# 技能管理 API

## GET /api/fkteams/skills

列出本地已安装技能。

**成功响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "skills": [],
    "total": 0
  }
}
```

`skills` 元素结构来自 `commands/skill.LocalSkillInfo`。

---

## GET /api/fkteams/skills/search

搜索技能市场。

**Query 参数**：

| 参数 | 类型 | 必填 | 默认 | 说明 |
| ---- | ---- | ---- | ---- | ---- |
| `q` | string | 是 | - | 搜索关键词 |
| `page` | int | 否 | `1` | 页码，必须大于 0 |
| `size` | int | 否 | `20` | 每页数量，最大 50 |
| `sort` | string | 否 | `downloads` | 排序字段 |
| `order` | string | 否 | `desc` | 排序方向 |

**成功响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "skills": [],
    "total": 0,
    "page": 1,
    "size": 20
  }
}
```

**失败响应**：

| 状态码 | message | 说明 |
| ------ | ------- | ---- |
| 400 | `keyword is required` | 缺少 `q` |
| 503 | `no skill provider available` | 无可用技能市场提供者 |
| 500 | 错误详情 | 搜索失败 |

---

## POST /api/fkteams/skills/install

从技能市场安装技能。

**请求 Body**：

```json
{
  "slug": "skill-slug"
}
```

**成功响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "slug": "skill-slug",
    "message": "skill installed"
  }
}
```

**失败响应**：

| 状态码 | message | 说明 |
| ------ | ------- | ---- |
| 400 | `slug is required` | 缺少 slug |
| 503 | `no skill provider available` | 无可用技能市场提供者 |
| 500 | 错误详情 | 安装失败 |

---

## DELETE /api/fkteams/skills/:slug

删除本地已安装技能。

**成功响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "slug": "skill-slug",
    "message": "skill removed"
  }
}
```

---

## GET /api/fkteams/skills/:slug/files

列出技能目录文件。

**Query 参数**：

| 参数 | 类型 | 必填 | 说明 |
| ---- | ---- | ---- | ---- |
| `path` | string | 否 | 技能目录内的相对路径 |

**成功响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "slug": "skill-slug",
    "files": []
  }
}
```

**失败响应**：

| 状态码 | message | 说明 |
| ------ | ------- | ---- |
| 400 | `slug is required` | 缺少 slug |
| 404 | 错误详情 | 技能或路径不存在 |

---

## GET /api/fkteams/skills/:slug/file

读取技能文件内容。

**Query 参数**：

| 参数 | 类型 | 必填 | 说明 |
| ---- | ---- | ---- | ---- |
| `path` | string | 是 | 技能目录内文件路径 |

**成功响应**：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "slug": "skill-slug",
    "path": "SKILL.md",
    "content": "# Skill"
  }
}
```

**失败响应**：

| 状态码 | message | 说明 |
| ------ | ------- | ---- |
| 400 | `slug and path are required` | 参数缺失 |
| 404 | 错误详情 | 技能或文件不存在 |
