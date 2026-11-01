# honi

a LLM helper for gentoo, she is a nice girl.

基于 OpenAI 兼容 Responses API 的本地 AI 助手，集成 TUI 交互式对话、单次问答、微信机器人和 MCP 工具调用。

## 特性

- **交互式 TUI**：基于 Bubbletea 的终端聊天界面，支持流式输出
- **单次问答**：`ask` 命令一行搞定，适合脚本调用
- **微信机器人**：`serve` 启动 HTTP 服务，通过微信收发消息
- **MCP 工具**：支持接入 MCP 协议的工具服务（如 AnySearch）
- **多 LLM Provider**：兼容 OpenAI 接口的厂商随意切换
- **配置自举**：首次运行自动生成默认配置和系统提示词

## 安装

从源码构建：

```sh
git clone <repo-url>
cd honi-agent
go build -o honi
```

## 快速开始

首次运行任意命令会自动在 `$XDG_CONFIG_HOME/honi/` 下生成默认配置：

```sh
./honi
```

编辑配置文件填入你的 API Key：

```toml
# ~/.config/honi/config.toml
debug = false
provider = "mimo"

[[LLM]]
  provider = "openAi"
  model = "mimo-v2.6-flash"
  base_url = "https://token-plan-cn.xiaomimimo.com/v1"
  api_key = "your-api-key"

[[MCP]]
  name = "anysearch"
  endpoint = "https://api.anysearch.com/mcp"
  api_key = "your-mcp-api-key"
```

系统提示词位于 `~/.config/honi/honi.md`，可自由编辑。

## 命令

| 命令 | 说明 |
|------|------|
| `honi` | 启动 TUI 交互式聊天 |
| `honi ask <message>` | 单次问答，输出到 stdout |
| `honi serve` | 启动 HTTP 守护进程（微信机器人，端口 `:12358`） |
| `honi login` | 登录微信账号 |
| `honi config` | 查看 / 设置配置 |

### 配置管理

```sh
# 查看当前配置
honi config -l

# 切换 provider
honi config --provider openAi
```

## 项目结构

```
honi-agent/
├── honi.go              # 入口
├── config/              # 配置加载与默认配置（embed）
├── internal/
│   ├── agent/           # Agent 核心：LLM 调用循环、Responses API
│   ├── api/             # HTTP 服务（Fiber）与微信 webhook
│   ├── auth/            # 微信登录凭证管理
│   ├── cmd/             # Cobra CLI 命令
│   ├── llm/responses/   # Responses API 工具与函数定义
│   ├── tea/             # Bubbletea TUI
│   └── tool/            # 工具实现（MCP、AnySearch、文件系统等）
└── common/              # 公共常量与工具函数
```

## 技术栈

- [Cobra](https://github.com/spf13/cobra) — CLI 框架
- [Bubbletea](https://github.com/charmbracelet/bubbletea) — TUI 框架
- [Fiber](https://gofiber.io/) — HTTP 服务
- [goai](https://github.com/zendev-sh/goai) — LLM 调用
- [go-sdk (MCP)](https://github.com/modelcontextprotocol/go-sdk) — MCP 协议

## License

见 [LICENSE](LICENSE)。
