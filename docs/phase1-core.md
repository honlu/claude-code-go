# Phase 1: 核心引擎实现

## 概述

Phase 1 完成了 Claude Code Go 版本的核心引擎实现，使得项目成为一个**可运行的最小原型**。

## 完成组件

### 1. API 客户端 (`internal/api/`)

完整实现了与 Anthropic API 的交互：

- **Provider 支持**: Direct, AWS Bedrock, Azure Foundry, Google Vertex
- **Message API**: `CreateMessage` 和 `CreateMessageStream`
- **SSE 流式处理**: 支持 `content_block_delta` 事件
- **认证**: API Key, Bearer Token, 自定义 Header

**关键文件:**

| 文件 | 说明 |
|------|------|
| `client.go` | API 客户端实现 |
| `streaming.go` | 流式处理 |
| `types.go` | 类型定义 |
| `provider.go` | Provider 检测 |

### 2. 工具执行引擎 (`internal/tools/execution.go`)

核心工具调用引擎：

```go
type ToolExecutor struct {
    registry   *Registry
    timeout    time.Duration
    semaphore  chan struct{}
    mu         sync.RWMutex
}
```

**功能:**

- 工具注册表管理
- 并发控制（最大 10 并发）
- 超时控制（默认 5 分钟）
- 权限检查
- 输入验证（JSON Schema）
- 进度报告

**关键方法:**

```go
func (e *ToolExecutor) Execute(ctx context.Context, toolName string, args map[string]any, tc ToolContext) (*ToolResult, error)
func (e *ToolExecutor) ExecuteWithProgress(ctx context.Context, toolName string, args map[string]any, tc ToolContext, onProgress ToolProgress) (*ToolResult, error)
func (e *ToolExecutor) ValidateInput(tool Tool, args map[string]any) error
func (e *ToolExecutor) ExecuteMultiple(ctx context.Context, calls []ToolCall, tc ToolContext) ([]*ToolResult, error)
```

### 3. 权限系统 (`internal/permissions/`)

完整的权限检查框架：

```go
type PermissionEngine struct {
    alwaysAllow  []PermissionRule
    alwaysDeny   []PermissionRule
    alwaysAsk    []PermissionRule
    rules        []PermissionRule
    mode         PermissionMode
}
```

**权限模式:**

```go
const (
    ModeAccept   PermissionMode = "accept"   // 自动接受
    ModeDeny     PermissionMode = "deny"     // 自动拒绝
    ModePrompt   PermissionMode = "prompt"    // 询问用户
    ModeApprove  PermissionMode = "approve"  // 单次批准
)
```

**内置规则:**

| 规则 | 说明 |
|------|------|
| `ToolNameRule` | 按工具名匹配 |
| `PathRule` | 按文件路径匹配 |
| `ReadOnlyRule` | 只读模式拒绝写入 |
| `SafePathsRule` | 安全路径限制 |
| `BashCommandRule` | Bash 命令限制 |
| `CompositeRule` | 组合多个规则 |

**使用示例:**

```go
engine := permissions.NewPermissionEngine()
engine.SetMode(permissions.ModePrompt)

// 添加规则
engine.AddAlwaysAllow(permissions.NewToolNameRule("Read", true, "allow read"))

// 检查权限
result, _ := engine.Check(ctx, "Bash", args, tc)
```

### 4. 查询引擎 (`internal/query/engine.go`)

主查询循环引擎：

```go
type QueryEngine struct {
    apiClient    *api.Client
    toolExecutor *tools.ToolExecutor
    permissions  *permissions.PermissionEngine
    messages     []api.Message
    model        string
    maxIterations int
}
```

**功能:**

- 用户消息管理
- API 调用（普通和流式）
- 完整工具调用循环（max 20 iterations）
- 对话历史管理

**使用示例:**

```go
engine := query.NewQueryEngine(
    query.WithModel("claude-sonnet-4-6"),
    query.WithAPIKey("sk-..."),
)

// 单次查询
err := engine.Query(ctx, "帮我创建 hello.go")

// 流式查询
err := engine.QueryStream(ctx, "你好", func(chunk string) error {
    fmt.Print(chunk)
    return nil
})
```

### 5. 交互式 REPL (`internal/cli/repl.go`)

命令行交互界面：

```go
type REPL struct {
    engine    *QueryEngine
    reader    *bufio.Reader
    history   []string
    model     string
}
```

**支持命令:**

| 命令 | 说明 |
|------|------|
| `/exit`, `/quit` | 退出 REPL |
| `/clear` | 清空对话历史 |
| `/model [name]` | 显示或切换模型 |
| `/history` | 查看命令历史 |
| `/help` | 显示帮助 |

## 文件结构

```
claude-code-go/
├── internal/
│   ├── api/
│   │   ├── client.go      # API 客户端
│   │   ├── streaming.go    # 流式处理
│   │   ├── types.go        # 类型定义
│   │   └── provider.go     # Provider 检测
│   ├── tools/
│   │   ├── registry.go     # 工具注册表
│   │   ├── execution.go    # 工具执行引擎
│   │   ├── tool.go         # 工具接口
│   │   └── ...             # 内置工具实现
│   ├── permissions/
│   │   ├── engine.go       # 权限引擎
│   │   └── rules.go        # 权限规则
│   ├── query/
│   │   └── engine.go       # 查询引擎
│   ├── cli/
│   │   ├── root.go         # CLI 入口
│   │   └── repl.go         # REPL 实现
│   └── ...
└── docs/
    ├── README.md
    └── phase1-core.md
```

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

## 完成状态

Phase 1 核心引擎已完整实现，所有 Phase 1-6 均已完成。

### 已实现功能

- ✅ 完整工具调用循环（max 20 iterations）
- ✅ 权限检查框架（ModePrompt/ModeAccept/ModeDeny）
- ✅ API 错误处理基础
- ✅ 6 个内置工具（Bash, Read, Edit, Write, Grep, Glob）
- ✅ 流式响应处理
- ✅ 查询引擎和 REPL
- ✅ 单元测试覆盖
