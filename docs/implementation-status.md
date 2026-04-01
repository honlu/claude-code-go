# Claude Code Go 实现状态

**更新时间**: 2026-04-01

## 总览

Go 版本已完成核心模块实现，可作为 CLI 工具运行。以下是各模块的详细实现状态。

## 已完成模块 ✅

| 模块 | 文件 | 状态 | 说明 |
|------|------|------|------|
| API Client | `internal/api/` | ✅ | 支持 Direct/Bedrock/Foundry/Vertex |
| Tools | `internal/tools/` | ✅ | Bash/Read/Edit/Write/Grep/Glob |
| Permissions | `internal/permissions/` | ✅ | 规则引擎 + ModePrompt/Accept/Deny |
| Query Engine | `internal/query/` | ✅ | 工具调用循环 (max 20 iterations) |
| Coordinator | `internal/coordinator/` | ✅ | 多 Agent 协调 |
| MCP | `internal/mcp/` | ✅ | stdio/SSE/HTTP/WebSocket 传输 |
| CLI/REPL | `internal/cli/` | ✅ | 交互式 REPL |
| Bootstrap | `internal/bootstrap/` | ✅ | 全局状态管理 |

## 已实现但需完善 ⚠️

### assistant (`internal/assistant/`)
**状态**: 基础实现完成
**需要完善**:
- [ ] Companion sprite 动画
- [ ] Buddy notification 系统
- [ ] Assistant mode 特定功能

### bridge (`internal/bridge/`)
**状态**: 基础实现完成
**需要完善**:
- [ ] CCR (Cloud Code Remote) WebSocket 连接
- [ ] Environments API 集成
- [ ] Session spawning 管理
- [ ] Bridge logger + QR code UI
- [ ] Ingress message routing

### commands (`internal/commands/`)
**状态**: 基础命令注册完成
**需要完善**:
- [ ] `/init` - 项目初始化命令
- [ ] `/install` - 安装更新命令
- [ ] `/commit` - Git commit 工作流
- [ ] `/review` - Code review 功能
- [ ] `/insights` - Session 分析
- [ ] `/bridge-kick` - Debug 命令
- [ ] 100+ slash commands (TS 版本有完整实现)

### context (`internal/context/`)
**状态**: Go 实现完成 (TS 版本为 React hooks)
**说明**: Go 版本不需要 React context

### hooks (`internal/hooks/`)
**状态**: 基础 hooks 实现完成
**说明**: React hooks 不适用于 Go CLI

### keybindings (`internal/keybindings/`)
**状态**: 基础解析器完成
**需要完善**:
- [ ] Chord 支持 (多键组合)
- [ ] User bindings override
- [ ] KeybindingProvider (React)
- [ ] 默认快捷键绑定

### memdir (`internal/memdir/`)
**状态**: 基础内存目录完成
**需要完善**:
- [ ] loadMemoryPrompt()
- [ ] buildMemoryPrompt() - 4 种 memory type
- [ ] findRelevantMemories() - Sonnet 相关性选择
- [ ] scanMemoryFiles()
- [ ] team memory 支持
- [ ] auto-memory 持久化

### migrations (`internal/migrations/`)
**状态**: 基础迁移框架完成
**需要完善**:
- [ ] migrateLegacyOpusToCurrent
- [ ] migrateAutoUpdatesToSettings
- [ ] migrateBypassPermissionsAcceptedToSettings
- [ ] migrateEnableAllProjectMcpServersToSettings
- [ ] migrateFennecToOpus / migrateSonnet1mToSonnet45 等模型迁移

### plugins (`internal/plugins/`)
**状态**: 基础 plugin loader 完成
**需要完善**:
- [ ] Plugin 安装/卸载
- [ ] Plugin 市场集成
- [ ] 插件生命周期管理

### remote (`internal/remote/`)
**状态**: 基础 HTTP client 完成
**需要完善**:
- [ ] RemoteSessionManager - CCR WebSocket
- [ ] SessionsWebSocket - 重连 + exponential backoff
- [ ] SDK message adapter
- [ ] Permission bridging

### server (`internal/server/`)
**状态**: 基础 HTTP server 完成
**需要完善**:
- [ ] Direct-connect session 管理
- [ ] WebSocket 通信
- [ ] Session lifecycle (starting/running/detached/stopping/stopped)

### services (`internal/services/`)
**状态**: 基础 services 完成
**需要完善**:
- [ ] `api/claude.ts` - Anthropic API 封装
- [ ] `api/bootstrap.ts` - Bootstrap API
- [ ] `api/sessionIngress.ts` - Session ingress
- [ ] `api/usage.ts` - Usage tracking
- [ ] `api/withRetry.ts` - Retry logic
- [ ] `compact/` - Conversation compaction
- [ ] `lsp/` - LSP client/server
- [ ] `analytics/` - GrowthBook/Datadog
- [ ] `voiceStreamSTT.ts` - Voice transcription

### skills (`internal/skills/`)
**状态**: 基础 skill 注册完成
**需要完善**:
- [ ] getSkillDirCommands() - filesystem skills 加载
- [ ] parseSkillFrontmatterFields() - YAML frontmatter
- [ ] discoverSkillDirsForPaths() - 动态 skill 发现
- [ ] activateConditionalSkillsForPaths() - 条件激活
- [ ] MCP skill builders

### state (`internal/state/`)
**状态**: 基础 state + history 完成
**需要完善**:
- [ ] history.jsonl 持久化
- [ ] paste store (大文件粘贴)
- [ ] lockfile 支持

### tasks (`internal/tasks/`)
**状态**: 基础 task 管理完成
**需要完善**:
- [ ] TaskState 完整状态机
- [ ] LocalShell / LocalAgent / RemoteAgent / InProcessTeammate
- [ ] Background task indicator
- [ ] Task lifecycle events

### upstreamproxy (`internal/upstreamproxy/`)
**状态**: 基础 proxy 完成
**需要完善**:
- [ ] UpstreamProxyRelay WebSocket relay
- [ ] HTTP CONNECT tunnel
- [ ] protobuf chunk encoding
- [ ] mTLS 支持
- [ ] getUpstreamProxyEnv()

### voice (`internal/voice/`)
**状态**: 基础 service 接口完成
**需要完善**:
- [ ] isVoiceModeEnabled() - GrowthBook kill-switch
- [ ] hasVoiceAuth() - OAuth token 检查
- [ ] Voice STT/TTS 集成

## 无需实现 (UI) ❌

这些是 React UI 组件，Go CLI 版本不需要：

| 模块 | 说明 |
|------|------|
| `components/` | React UI 组件 |
| `ink/` | React-based terminal UI |
| `screens/` | UI screens |
| `dialogLaunchers.tsx` | Dialog 组件 |

## 核心 TS 文件对应 Go 实现

| TS 文件 | Go 实现 | 状态 |
|---------|---------|------|
| `QueryEngine.ts` | `internal/query/engine.go` | ✅ |
| `query.ts` | `internal/query/engine.go` | ✅ |
| `Tool.ts` | `internal/tools/tool.go` | ✅ |
| `tools.ts` | `internal/tools/registry.go` | ✅ |
| `bootstrap/*` | `internal/bootstrap/` | ✅ |
| `coordinator/*` | `internal/coordinator/` | ✅ |
| `mcp/*` | `internal/mcp/` | ✅ |
| `replLauncher.tsx` | `internal/cli/repl.go` | ✅ |
| `permissions/*` | `internal/permissions/` | ✅ |
| `Task.ts` | `internal/tasks/task.go` | ⚠️ 基础 |
| `commands.ts` | `internal/commands/` | ⚠️ 基础 |
| `history.ts` | `internal/state/state.go` | ⚠️ 基础 |
| `skills/*` | `internal/skills/` | ⚠️ 基础 |
| `memdir/*` | `internal/memdir/` | ⚠️ 基础 |
| `server/*` | `internal/server/` | ⚠️ 基础 |
| `remote/*` | `internal/remote/` | ⚠️ 基础 |
| `bridge/*` | `internal/bridge/` | ⚠️ 基础 |
| `upstreamproxy/*` | `internal/upstreamproxy/` | ⚠️ 基础 |
| `voice/*` | `internal/voice/` | ⚠️ 基础 |
| `migrations/*` | `internal/migrations/` | ⚠️ 基础 |

## 下一步工作

### 优先级 1 (核心功能)
1. **commands** - 实现主要 slash 命令 (`/init`, `/commit`)
2. **server** - WebSocket session 管理
3. **remote** - CCR session 连接

### 优先级 2 (重要功能)
4. **memdir** - auto-memory 系统
5. **tasks** - 完整 task 状态机
6. **skills** - filesystem skill 加载

### 优先级 3 (增强功能)
7. **bridge** - 完整 CCR bridge
8. **services** - API/analytics/compact
9. **plugins** - 插件市场

## 测试状态

```
go build ./...     ✅
go test ./...      ✅
./claude-go tools  ✅
```

## 文件统计

| 目录 | Go 文件数 |
|------|-----------|
| `internal/` | 45+ |
| `pkg/` | 2 |
| `cmd/` | 1 |
| **总计** | 48+ |
