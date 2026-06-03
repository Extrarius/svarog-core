---
name: svarog-mcp-tool
description: Add a new MCP tool to svarog-core internal/mcp server (mark3labs/mcp-go) with rich descriptions, registration in Build(), and optional HTTP/stdio smoke. Use when extending the custom MCP server with a new tool, resource, or prompt pattern.
---

# svarog-mcp-tool

How to add a **tool** to the svarog MCP server in this repo.

## Layout

| Path | Role |
|------|------|
| [`internal/mcp/server.go`](../../../internal/mcp/server.go) | `Build()`, `registerTools`, resources, prompts |
| [`internal/mcp/deps.go`](../../../internal/mcp/deps.go) | `Deps` (pgx pool, HTTP client, config) |
| [`internal/mcp/config.go`](../../../internal/mcp/config.go) | envconfig |
| [`cmd/mcp-stdio`](../../../cmd/mcp-stdio/) | STDIO transport |
| [`cmd/mcp-http`](../../../cmd/mcp-http/) | Streamable HTTP `/mcp` |

## Steps

1. **Define the tool** in `registerTools()`:

```go
tool := mcp.NewTool("my_tool",
    mcp.WithDescription(`Long description for the agent: when to use, what it returns, side effects, limits.`),
    mcp.WithString("param", mcp.Required(), mcp.Description("...")),
)
s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // req.RequireString("param") etc.
    return mcp.NewToolResultText("..."), nil
})
```

2. **Implement logic** in the same file or a new `internal/mcp/my_tool.go` helper.
3. **Update resource** `svarog://service/info` payload in `registerResources` if the tool is user-facing.
4. **Env**: document new vars in [`.env.example`](../../../.env.example) if needed.
5. **Docs**: add row to [`doc/MCP.md`](../../../doc/MCP.md).

## Description quality (UX for agents)

- First paragraph: **when** to call the tool.
- Per-argument `mcp.Description`: format, examples, defaults.
- State read-only vs write, external API vs Postgres vs svarog HTTP.

## Verify

```bash
go vet ./...
make run-mcp-stdio   # Cursor: svarog-stdio in .cursor/mcp.json
# or
make up              # mcp-http on :8000/mcp
```

Reload MCP in Cursor after code changes.

## References

- [`doc/MCP.md`](../../../doc/MCP.md)
- Existing tools: `weather_forecast`, `list_users`, `list_active_sessions`, `register_user`
