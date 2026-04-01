# Claude Code Go 架构文档

## 概述

Claude Code Go 是一个用 Go 语言实现的 CLI 工具，旨在提供与 TypeScript 版本（Claude Code 2.1.88）功能等效的体验。

## 系统架构

```
┌─────────────────────────────────────────────────────────────┐
│                        CLI Layer                             │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────────┐   │
│  │  agent  │  │  tools  │  │ version │  │    help     │   │
│  └────┬────┘  └────┬────┘  └────┬────┘  └──────┬──────┘   │
└───────┼────────────┼────────────┼───────────────┼──────────┘
        │            │            │               │
        ▼            ▼            ▼               ▼
┌─────────────────────────────────────────────────────────────┐
│                     Core Engine                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                   Query Engine                       │   │
│  │  ┌─────────┐  ┌────────────┐  ┌────────────────┐  │   │
│  │  │ Message │  │ Tool Loop   │  │ Stream Handler │  │   │
│  │  │ Manager │  │ (max 20 it) │  │                │  │   │
│  │  └────┬────┘  └─────┬──────┘  └────────────────┘  │   │
│  │       │             │                             │   │
│  │       ▼             ▼                             │   │
│  │  ┌────────────────────────────────────────────┐   │   │
│  │  │           Tool Executor                     │   │   │
│  │  │  ┌──────────┐ ┌──────────┐ ┌───────────┐  │   │   │
│  │  │  │Registry │ │ Semaphore │ │ Timeout   │  │   │   │
│  │  │  └──────────┘ └──────────┘ └───────────┘  │   │   │
│  │  └────────────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────┘   │
│                            │                               │
│  ┌─────────────────────────┼─────────────────────────────┐ │
│  │              Permission Engine                        │ │
│  │  ┌────────────┐  ┌────────────┐  ┌───────────────┐  │ │
│  │  │AlwaysAllow │  │ AlwaysDeny │  │     Rules     │  │ │
│  │  └────────────┘  └────────────┘  └───────────────┘  │ │
│  └──────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   External Services                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────┐    │
│  │ Anthropic   │  │    MCP      │  │   File System   │    │
│  │   API       │  │   Servers   │  │   (Tools)       │    │
│  └─────────────┘  └─────────────┘  └─────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

## 核心组件

### 1. QueryEngine

查询引擎是系统的核心，负责：

1. 维护对话消息历史
2. 调用 Anthropic API
3. 处理工具调用循环
4. 管理流式响应

**流程:**

```
用户输入
    │
    ▼
┌────────────────────────────────────┐
│ Add User Message to History        │
└────────────────────────────────────┘
    │
    ▼
┌────────────────────────────────────┐
│ Build API Request                  │
│ - Messages from history            │
│ - Tools from registry             │
│ - Model configuration              │
└────────────────────────────────────┘
    │
    ▼
┌────────────────────────────────────┐
│ Call API (CreateMessage/Stream)    │
└────────────────────────────────────┘
    │
    ├────────────────────┐
    ▼                    ▼
┌─────────┐        ┌──────────┐
│ Text    │        │Tool Call │
│ Response│        │ Detected │
└────┬────┘        └────┬─────┘
     │                  │
     ▼                  ▼
┌──────────────────────────────┐
│ Store in Message History     │
└──────────────────────────────┘
    │
    ▼
┌────────────────────────────────────┐
│ Execute Tools (if any)            │
│ - Permission Check                │
│ - Tool Call                      │
│ - Format Result                  │
└────────────────────────────────────┘
    │
    ▼
┌────────────────────────────────────┐
│ Add Tool Result to History         │
└────────────────────────────────────┘
    │
    │ Loop until no more tool calls
    │
    ▼
返回结果给用户
```

### 2. ToolExecutor

工具执行器负责：

- 工具注册和查找
- 并发控制（最大 10 并发）
- 超时管理（默认 5 分钟）
- 输入验证
- 进度回调

### 3. PermissionEngine

权限引擎提供：

- 多模式支持（accept/deny/prompt/approve）
- 可插拔规则系统
- 规则优先级：alwaysDeny > alwaysAllow > mode > alwaysAsk > rules

### 4. ToolRegistry

工具注册表维护：

- 所有可用工具的注册表
- 工具查找（by name）
- 工具列表（for API）

## 数据流

### 非流式查询

```
QueryEngine.Query(input)
    │
    ▼
API.CreateMessage(request)
    │
    ▼
QueryEngine.processResponse(response)
    │
    ├────────────────────────┐
    ▼                        ▼
Text Content            Tool Calls
    │                        │
    ▼                        ▼
Add to History      ToolExecutor.Execute()
    │                        │
    │                        ▼
    │                PermissionEngine.Check()
    │                        │
    │                        ▼
    │                Tool.Call()
    │                        │
    ▼                        ▼
[Done]              Add Tool Result to History
                         │
                         ▼
                   [Loop: Send result to API]
```

### 流式查询

```
QueryEngine.QueryStream(input, onChunk)
    │
    ▼
API.CreateMessageStream(request)
    │
    ▼
Event Loop:
    │
    ├─ "content_block_delta" ──► onChunk(text)
    │
    ├─ "content_block_start" ──► Detect tool_use blocks
    │
    └─ "message_stop" ──────────► Store final content
                                     │
                                     ▼
                              Execute any tool calls
                                     │
                                     ▼
                              [Loop or Done]
```

## 并发模型

### 工具执行并发

- 最大 10 个工具并发执行
- 通过 `semaphore chan struct{}` 实现
- 非并发安全的工具串行执行

```go
if !tool.IsConcurrencySafe(args) {
    select {
    case e.semaphore <- struct{}{}:
        defer func() { <-e.semaphore }()
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}
```

### Agent 并发

Coordinator 支持多 Agent 并行：

```go
for i := 0; i < agentCount; i++ {
    go func(id int) {
        agent, _ := c.SpawnAgent(ctx, "worker", "Task", tools)
        results <- agent
    }(i)
}
```

## 错误处理

### API 错误

```go
type APIError struct {
    Type    string  // "rate_limit_error", "invalid_request_error", etc.
    Message string
    Code    string
}
```

### 工具错误

```go
type ToolResult struct {
    Data          any
    Error         error
    NewMessages   []Message
    IsImage       bool
    PersistedFile string
}
```

## 配置

### 环境变量

| 变量 | 说明 |
|------|------|
| `ANTHROPIC_API_KEY` | API 密钥 |
| `ANTHROPIC_API_URL` | 自定义 API URL |
| `CLAUDE_CODE_USE_BEDROCK` | 使用 AWS Bedrock |
| `CLAUDE_CODE_USE_FOUNDRY` | 使用 Azure Foundry |
| `CLAUDE_CODE_USE_VERTEX` | 使用 Google Vertex |

### CLI 参数

```bash
./claude-go agent [message]     # 交互模式或单次查询
  --model, -m                   # 指定模型
  --print                       # 仅打印，不执行工具
  --permission-mode             # 权限模式
  --dangerously-skip-permissions # 跳过权限检查
```

## 扩展

### 添加新工具

1. 创建新文件 `internal/tools/mytool/mytool.go`
2. 实现 `Tool` 接口
3. 在 `init()` 中注册

```go
package mytool

type myTool struct{}

func NewTool() *myTool { return &myTool{} }

func (t *myTool) Name() string { return "MyTool" }
func (t *myTool) Description(_ map[string]any) string { return "My tool" }
func (t *myTool) InputSchema() any { return nil }
func (t *myTool) Call(ctx context.Context, args map[string]any, tc tools.ToolContext) (*tools.ToolResult, error) {
    return &tools.ToolResult{Data: "result"}, nil
}
func (t *myTool) IsConcurrencySafe(_ map[string]any) bool { return true }
func (t *myTool) IsReadOnly(_ map[string]any) bool { return true }
func (t *myTool) CheckPermissions(_ context.Context, _ map[string]any, _ tools.ToolContext) (tools.PermissionResult, error) {
    return tools.PermissionResult{Behavior: "allow"}, nil
}

func init() {
    tools.Register("MyTool", NewTool())
}
```

### 添加新权限规则

```go
type CustomRule struct{}

func (r *CustomRule) Name() string { return "CustomRule" }
func (r *CustomRule) Check(ctx context.Context, toolName string, args map[string]any) (bool, string) {
    // 自定义检查逻辑
    return true, ""
}

// 使用
engine.AddRule(&CustomRule{})
```
