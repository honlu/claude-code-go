# API 类型参考

## 概览

本文档描述 Claude Code Go 使用的 API 类型，与 Anthropic API 对应。

## Message

对话消息。

```go
type Message struct {
    Role    string `json:"role"`      // "user" or "assistant"
    Content any    `json:"content"`   // string or []ContentBlock
}
```

**示例:**

```json
{
    "role": "user",
    "content": "帮我创建 hello.go"
}
```

## ContentBlock

内容块，可以是文本或工具调用。

```go
type ContentBlock struct {
    Type         string         `json:"type"`            // "text" or "tool_use"
    Text         string         `json:"text,omitempty"`
    ToolUse      *ToolUse       `json:"tool_use,omitempty"`
    Id           string         `json:"id,omitempty"`
    Name         string         `json:"name,omitempty"`
    Input        map[string]any `json:"input,omitempty"`
    ToolUseId    string         `json:"tool_use_id,omitempty"`
    Content      string         `json:"content,omitempty"`
    IsError      bool           `json:"is_error,omitempty"`
}
```

**示例 (text):**

```json
{
    "type": "text",
    "text": "好的，我来创建文件。"
}
```

**示例 (tool_use):**

```json
{
    "type": "tool_use",
    "id": "toolu_01A2B3C4D5E6F7",
    "name": "Write",
    "input": {
        "path": "hello.go",
        "content": "package main\n\nfunc main() {\n    println(\"Hello, World!\")\n}"
    }
}
```

## ToolUse

工具使用请求。

```go
type ToolUse struct {
    Type      string         `json:"type"`
    Name      string         `json:"name"`
    Input     map[string]any `json:"input"`
    Id        string         `json:"id"`
}
```

## Tool

工具定义。

```go
type Tool struct {
    Name        string       `json:"name"`
    Description string       `json:"description,omitempty"`
    InputSchema *InputSchema `json:"input_schema,omitempty"`
}
```

**示例:**

```json
{
    "name": "Write",
    "description": "Create or overwrite a file with content",
    "input_schema": {
        "type": "object",
        "properties": {
            "path": {
                "type": "string",
                "description": "Path to the file"
            },
            "content": {
                "type": "string",
                "description": "Content to write"
            }
        },
        "required": ["path", "content"]
    }
}
```

## InputSchema

工具输入模式。

```go
type InputSchema struct {
    Type       string         `json:"type,omitempty"`
    Properties map[string]any `json:"properties,omitempty"`
    Required   []string       `json:"required,omitempty"`
}
```

## MessageRequest

创建消息请求。

```go
type MessageRequest struct {
    Model       string    `json:"model"`
    Messages    []Message `json:"messages"`
    System      string    `json:"system,omitempty"`
    MaxTokens   int       `json:"max_tokens"`
    Tools       []Tool    `json:"tools,omitempty"`
    Stream      bool      `json:"stream,omitempty"`
    Temperature float64   `json:"temperature,omitempty"`
    TopP        float64   `json:"top_p,omitempty"`
    TopK        int       `json:"top_k,omitempty"`
}
```

**示例:**

```json
{
    "model": "claude-sonnet-4-6",
    "messages": [
        {"role": "user", "content": "你好"}
    ],
    "max_tokens": 4096,
    "tools": [
        {
            "name": "Write",
            "description": "Create or overwrite a file",
            "input_schema": {
                "type": "object",
                "properties": {
                    "path": {"type": "string"},
                    "content": {"type": "string"}
                },
                "required": ["path", "content"]
            }
        }
    ]
}
```

## MessageResponse

创建消息响应。

```go
type MessageResponse struct {
    Id         string         `json:"id"`
    Type       string         `json:"type"`
    Role       string         `json:"role"`
    Content    []ContentBlock `json:"content"`
    Model      string         `json:"model"`
    StopReason string         `json:"stop_reason"`
    StopSequence interface{}   `json:"stop_sequence"`
    Usage      MessageUsage   `json:"usage"`
}
```

**StopReason 值:**

- `end_turn` - 对话结束
- `tool_use` - 模型请求调用工具
- `max_tokens` - 达到最大 token 限制

## MessageUsage

Token 使用统计。

```go
type MessageUsage struct {
    InputTokens  int `json:"input_tokens"`
    OutputTokens int `json:"output_tokens"`
}
```

## StreamEvent

流式事件。

```go
type StreamEvent struct {
    Type string `json:"type"`
    Data any    `json:"data"`
}
```

**事件类型:**

| 类型 | 说明 | Data 类型 |
|------|------|-----------|
| `content_block_start` | 内容块开始 | `map[string]any` |
| `content_block_delta` | 内容块增量 | `map[string]any` |
| `message_delta` | 消息增量 | `map[string]any` |
| `message_stop` | 消息结束 | - |

### content_block_delta

```json
{
    "type": "content_block_delta",
    "index": 0,
    "delta": {
        "type": "text_delta",
        "text": "好的，"
    }
}
```

### message_delta

```json
{
    "type": "message_delta",
    "delta": {
        "stop_reason": "tool_use"
    },
    "usage": {
        "input_tokens": 100,
        "output_tokens": 50
    }
}
```

## ErrorResponse

API 错误响应。

```go
type ErrorResponse struct {
    Type  string       `json:"type"`
    Error ErrorDetail  `json:"error"`
}

type ErrorDetail struct {
    Type    string `json:"type"`
    Message string `json:"message"`
    Code    string `json:"code,omitempty"`
}
```

**错误类型:**

- `authentication_error` - 认证失败
- `invalid_request_error` - 请求无效
- `rate_limit_error` - 速率限制
- `api_error` - 通用 API 错误

## Provider

API 提供者枚举。

```go
type Provider int

const (
    ProviderDirect Provider = iota  // 直接调用 Anthropic API
    ProviderBedrock                // AWS Bedrock
    ProviderFoundry                // Azure Foundry
    ProviderVertex                 // Google Vertex
)
```

## 模型信息

```go
type ModelInfo struct {
    Name         string
    DisplayName  string
    InputTokens  int
    OutputTokens int
    SupportsVision bool
    SupportsTools  bool
}
```

**默认模型:**

| 模型 | 输入 Token | 输出 Token | 工具支持 |
|------|-----------|-----------|----------|
| claude-opus-4-5 | 200000 | 8192 | ✅ |
| claude-sonnet-4-6 | 200000 | 8192 | ✅ |
| claude-haiku-4-5 | 200000 | 8192 | ✅ |

## 定价信息

```go
type PricingInfo struct {
    InputCostPerToken  float64
    OutputCostPerToken float64
    Currency          string
}
```

**默认定价 (USD):**

| 模型 | 输入 ($/1K) | 输出 ($/1K) |
|------|------------|------------|
| claude-opus-4-5 | $0.015 | $0.075 |
| claude-sonnet-4-6 | $0.003 | $0.015 |
| claude-haiku-4-5 | $0.0008 | $0.004 |
