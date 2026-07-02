# 配置指南

配置文件位于 `~/.fkteams/config/config.toml`，可通过下面命令生成示例：

```bash
fkteams generate config
```

## 模型池

`[[models]]` 是全局模型池。其他配置只通过 `id` 引用模型，不直接重复填写 `base_url`、`api_key` 等连接参数。

```toml
[[models]]
id = "main"
name = "主力模型"
use_for = ["chat", "agent"]
provider = "openai"
base_url = "https://api.openai.com/v1"
api_key = "your_api_key"
model = "gpt-5"

[[models]]
id = "fast"
name = "快速模型"
use_for = ["title", "summary"]
provider = "deepseek"
base_url = "https://api.deepseek.com/v1"
api_key = "your_deepseek_key"
model = "deepseek-chat"
```

字段说明：

| 字段 | 说明 |
| ---- | ---- |
| `id` | 稳定引用 ID，智能体、圆桌和 API 都用它引用模型 |
| `name` | 展示名称，可随时调整 |
| `use_for` | 默认用途，可选 `chat`、`agent`、`title`、`summary` |
| `provider` | 模型提供商，如 `openai`、`deepseek`、`claude`、`ollama`、`ark`、`gemini`、`qwen`、`openrouter`、`copilot` |
| `base_url` | 模型服务地址 |
| `api_key` | 模型服务密钥 |
| `model` | 上游真实模型名 |
| `extra_headers` | 额外 HTTP 请求头，格式为 `Key:Value,Other:Value` |

`use_for` 不能在多个模型中重复配置。必须有一个模型包含 `chat`；`agent`、`title`、`summary` 未配置时会回退到 `chat` 模型。

## 服务与认证

```toml
[server]
host = "127.0.0.1"
port = 23456
log_level = "info"
allow_origins = ["http://localhost:5173", "http://127.0.0.1:5173"]

[server.auth]
enabled = false
username = "admin"
password = "your_password"
secret = "your_jwt_secret"
```

## OpenAI 兼容 API

```toml
[openai_api]
api_keys = ["sk-fkteams-your-secret-key"]
```

客户端使用 `http://<host>:<port>/v1` 作为 Base URL，`model` 字段填写本地模型 `id`。

## 内置智能体

```toml
[agents]
researcher = true
assistant = true
analyst = false

[agents.ssh_visitor]
enabled = false
host = "ip:port"
username = "your_ssh_user"
password = "your_ssh_password"
```

## 圆桌讨论

圆桌成员顺序由 `[[roundtable.members]]` 数组顺序决定，不再单独配置序号。

```toml
[roundtable]
max_iterations = 2

[[roundtable.members]]
id = "logic"
name = "逻辑分析"
description = "擅长结构化推理和反驳"
model_id = "main"

[[roundtable.members]]
id = "creative"
name = "创意视角"
description = "擅长发散思考和补充方案"
model_id = "fast"
```

## 自定义智能体

```toml
[custom.moderator]
id = "moderator"
name = "协调者"
description = "负责协调成员协作"
prompt = "你是一个公正的协调者，负责根据任务需求合理分配给团队成员。"
model_id = "main"
tools = []

[[custom.agents]]
id = "frontend"
name = "前端开发专家"
description = "专注于前端开发"
prompt = """你是一个专业的前端开发工程师。
你擅长 React、TypeScript、CSS 和前端工程化。
"""
model_id = "main"
tools = ["file", "command", "search"]
```

字段说明：

| 字段 | 说明 |
| ---- | ---- |
| `id` | 稳定智能体 ID，用于 `@` 引用、命令参数和工具标识 |
| `name` | 展示名称 |
| `description` | 能力描述 |
| `prompt` | 系统提示词 |
| `model_id` | 引用 `[[models]].id` |
| `tools` | 可用工具列表，可包含内置工具和 `mcp-<server_id>` |

## MCP 服务

```toml
[[custom.mcp_servers]]
id = "filesystem"
name = "文件系统"
description = "文件系统操作工具"
enabled = true
timeout = "30s"
url = "http://127.0.0.1:3000/mcp"
transport = "http"

[[custom.mcp_servers]]
id = "postgres"
name = "PostgreSQL"
description = "数据库查询工具"
enabled = true
timeout = "30s"
command = "npx"
args = ["-y", "@modelcontextprotocol/server-postgres"]
env = { DATABASE_URL = "postgresql://localhost/mydb" }
transport = "stdio"
```

MCP 工具名为 `mcp-<server_id>`，例如上面的 `filesystem` 服务在智能体 `tools` 中写作 `mcp-filesystem`。

## 消息通道

通道 `mode` 只表示运行模式。需要绑定单个智能体时使用 `mode = "agent"` 和 `agent_id`。

```toml
[channels.qq]
enabled = false
app_id = "your_app_id"
app_secret = "your_app_secret"
sandbox = true
mode = "team"
agent_id = ""

[channels.discord]
enabled = false
token = "your_discord_bot_token"
allow_from = ""
mode = "team"
agent_id = ""

[channels.weixin]
enabled = false
base_url = "https://ilinkai.weixin.qq.com"
cred_path = "channels/weixin/credentials.json"
log_level = "info"
allow_from = ""
mode = "team"
agent_id = ""
```

## 数据目录与环境变量

默认应用目录为 `~/.fkteams`，可通过 `FEIKONG_APP_DIR` 覆盖。常用子目录包括 `workspace`、`sessions`、`scheduler`、`history`、`config`、`log`、`share` 和 `runtime`。

| 变量名 | 说明 | 默认值 |
| ------ | ---- | ------ |
| `FEIKONG_APP_DIR` | 应用数据目录 | `~/.fkteams` |
| `FEIKONG_PROXY_URL` | 代理地址 | - |
| `FEIKONG_MAX_ITERATIONS` | 智能体最大迭代次数，`0` 或 `-1` 表示不限制 | `60` |
