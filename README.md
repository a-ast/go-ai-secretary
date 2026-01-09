# AI Assistant for tasks (CLI edition)

This is an AI Agent that should help with everyday tasks:
* Checking and analyzing current tasks
* Checking emails and connecting them to tasks (to do)


## How to setup

1. Export env vars

```
export GOOGLE_API_KEY="..."
```

2. Authorize Notion MCP server

Run:
```
npx @modelcontextprotocol/inspector
```

In the Inspector UI select:

* Transport type: Streamable HTTP
* URL: https://mcp.notion.com/mcp

Click Open Auth Settings → choose Quick OAuth Flow → log in to Notion and approve

Follow the “OAuth Flow Progress” steps until it says Authentication complete

Copy the access_token value


## How to run

```
go run ./cmd/notionagent -notion_token={access_token}
```

## TO DO

1. Remove notion from names
2. Make notion to be only one of possible tools
3. Simplify authentication



