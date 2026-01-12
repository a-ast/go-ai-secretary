# AI Assistant for Tasks - Copilot Instructions

## Project Overview
This is a Go CLI application that provides an AI-powered assistant for managing TODO tasks via Notion integration. It uses Google's ADK framework for LLM agents and MCP (Model Context Protocol) for tool-based interactions with Notion.

## Architecture
- **Entry Point**: `cmd/notionagent/main.go` - CLI flags, app initialization, REPL runner
- **Core Logic**: `internal/agent/app.go` - App struct with ADK agent, session service, and runner
- **CLI Interface**: `internal/cli/repl.go` - Read-Eval-Print Loop with commands (/help, /exit)
- **Tool Integration**: Notion MCP tools filtered to read-only operations only

## Key Patterns
- **Agent Framework**: Uses `google.golang.org/adk` for LLM agents, sessions, and tool execution
- **MCP Tools**: `mcptoolset.New()` with custom `isReadOnlyNotionTool()` filter denying write operations
- **Session Management**: In-memory sessions via `session.InMemoryService()` for chat history
- **Authentication**: Bearer token auth for Notion MCP via `authTransport`
- **Error Handling**: Early returns with wrapped errors using `fmt.Errorf("init %s: %w", ...)`

## Development Workflow
- **Build/Run**: `go run ./cmd/notionagent -notion_token={token} -session={id} -v`
- **Notion Auth**: Use MCP Inspector for OAuth flow to get access token
- **Environment**: Set `GOOGLE_API_KEY` env var for Gemini model access
- **Debugging**: Use `-v` flag to print tool activity and event streams

## Code Conventions
- **Imports**: Use `agentpkg` alias for internal/agent package
- **Struct Initialization**: Config structs with named fields (e.g., `agentpkg.Config{ModelName: ..., ...}`)
- **Context Usage**: Pass `context.Context` through all async operations
- **Tool Filtering**: Pattern-based allow/deny lists in `isReadOnlyNotionTool()`
- **Instructions**: Agent prompts defined as functions returning trimmed strings

## Integration Points
- **Notion MCP**: Streamable HTTP transport to `https://mcp.notion.com/mcp`
- **Google Gemini**: Model initialization with API key for LLM responses
- **Session Persistence**: In-memory only; no external storage

## Common Tasks
- Adding new tools: Extend `allowSubstrings` in `isReadOnlyNotionTool()` for new read operations
- Modifying agent behavior: Update `instruction()` function with new prompt text
- CLI commands: Add new `/command` cases in `RunREPL()` switch statement</content>
<parameter name="filePath">/Users/orso/Projects/go-ai-secretary/.github/copilot-instructions.md