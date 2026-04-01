# Claude Code Go

Claude Code 的 Go 语言实现版本。

## 项目概述

这是一个从 TypeScript 版本（Claude Code 2.1.88）重写的 Go 版本，目标是用 Go 实现一个功能等效的 CLI 工具。

## 目录结构

```
claude-code-go/
├── cmd/claude/           # CLI 入口
├── internal/
│   ├── api/             # Anthropic API 客户端
│   ├── assistant/        # Assistant 主逻辑
│   ├── bootstrap/        # 全局状态管理
│   ├── bridge/           # 平台桥接
│   ├── buddy/           # Buddy Agent
│   ├── cli/              # CLI 命令和 REPL
│   ├── commands/          # 命令系统
│   ├── context/           # 上下文管理
│   ├── coordinator/       # 多 Agent 协调
│   ├── hooks/             # 生命周期钩子
│   ├── keybindings/       # 键盘快捷键
│   ├── mcp/              # MCP 客户端
│   ├── memdir/           # 内存目录
│   ├── migrations/        # 数据库迁移
│   ├── permissions/       # 权限系统
│   ├── plugins/           # 插件系统
│   ├── query/            # 查询引擎
│   ├── remote/           # 远程客户端
│   ├── server/           # HTTP 服务器
│   ├── services/         # 后台服务
│   ├── skills/           # 技能系统
│   ├── state/            # 状态管理
│   ├── tasks/            # 任务管理
│   ├── tools/            # 工具系统
│   └── upstreamproxy/     # 上游代理
├── pkg/
│   ├── constants/        # 常量定义
│   └── utils/            # 工具函数
└── docs/                # 文档
```

## 模块覆盖

| TS 模块 | Go 实现 | 状态 |
|---------|---------|------|
| assistant | `internal/assistant/` | ✅ |
| bootstrap | `internal/bootstrap/` | ✅ |
| bridge | `internal/bridge/` | ✅ |
| buddy | `internal/buddy/` | ✅ |
| cli | `internal/cli/` | ✅ |
| commands | `internal/commands/` | ✅ |
| context | `internal/context/` | ✅ |
| coordinator | `internal/coordinator/` | ✅ |
| hooks | `internal/hooks/` | ✅ |
| keybindings | `internal/keybindings/` | ✅ |
| mcp | `internal/mcp/` | ✅ |
| memdir | `internal/memdir/` | ✅ |
| migrations | `internal/migrations/` | ✅ |
| plugins | `internal/plugins/` | ✅ |
| query | `internal/query/` | ✅ |
| remote | `internal/remote/` | ✅ |
| schemas | `internal/api/types.go` | ✅ |
| server | `internal/server/` | ✅ |
| services | `internal/services/` | ✅ |
| skills | `internal/skills/` | ✅ |
| state | `internal/state/` | ✅ |
| tasks | `internal/tasks/` | ✅ |
| tools | `internal/tools/` | ✅ |
| types | `internal/api/types.go` | ✅ |
| upstreamproxy | `internal/upstreamproxy/` | ✅ |
| voice | `internal/voice/` | ✅ |
| utils | `pkg/utils/` | ✅ |
| constants | `pkg/constants/` | ✅ |

## 快速开始

### 构建

```bash
go build -o claude-go ./cmd/claude
```

### 测试

```bash
go test ./...
```

### 运行

```bash
# 列出可用工具
./claude-go tools

# 版本信息
./claude-go version

# 交互模式
./claude-go agent

# 单次查询
./claude-go agent "帮我创建 hello.go"
```

## 核心模块

### Tools (工具系统)

工具是 Claude Code 执行操作的核心组件。

| 工具 | 说明 | 状态 |
|------|------|------|
| Bash | 执行 shell 命令 | ✅ |
| Read | 读取文件 | ✅ |
| Edit | 编辑文件（old_string → new_string） | ✅ |
| Write | 创建/覆盖文件 | ✅ |
| Grep | 正则搜索 | ✅ |
| Glob | 文件匹配 | ✅ |

### API 客户端

支持多种部署方式：

- **Direct API**: 使用 `ANTHROPIC_API_KEY`
- **AWS Bedrock**: 使用 `CLAUDE_CODE_USE_BEDROCK`
- **Azure Foundry**: 使用 `CLAUDE_CODE_USE_FOUNDRY`
- **Google Vertex**: 使用 `CLAUDE_CODE_USE_VERTEX`

### 查询引擎

主查询循环引擎，支持：

- 用户消息管理
- API 调用（普通和流式）
- 完整工具调用循环（max 20 iterations）
- 对话历史管理

### 权限系统

完整的权限检查框架：

- **ModePrompt**: 询问用户（默认）
- **ModeAccept**: 自动接受
- **ModeDeny**: 自动拒绝
- **ModeApprove**: 单次批准

## 测试覆盖

| 包 | 测试数 | 状态 |
|----|--------|------|
| `internal/api` | - | ✅ |
| `internal/bootstrap` | - | ✅ |
| `internal/tools` | - | ✅ |
| `internal/query` | 10 | ✅ |
| `internal/permissions` | 16 | ✅ |
| `internal/coordinator` | - | ✅ |
| `internal/mcp` | - | ✅ |

## 文档

- [Phase 1 核心引擎实现](phase1-core.md)
- [系统架构](architecture.md)
- [API 类型参考](api-types.md)
- [实现状态](implementation-status.md)

## License

本项目仅用于研究目的。Claude Code 版权归 Anthropic 所有。
