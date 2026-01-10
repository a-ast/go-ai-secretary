package agent

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/mcptoolset"
	"google.golang.org/genai"
)

type Config struct {
	ModelName   string
	NotionURL   string
	NotionToken string
	Verbose     bool
}

type App struct {
	agent agent.Agent
	sess  session.Service
	run   *runner.Runner

	appName string
	userID  string

	verbose bool
}

func NewApp(ctx context.Context, cfg Config) (*App, error) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return nil, errors.New("GOOGLE_API_KEY is not set")
	}

	m, err := gemini.NewModel(ctx, cfg.ModelName, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, fmt.Errorf("init gemini model: %w", err)
	}

	toolset, err := newNotionMCPToolset(cfg)
	if err != nil {
		return nil, err
	}

	ag, err := llmagent.New(llmagent.Config{
		Name:        "notion_todo_agent",
		Description: "Read-only assistant that reads tasks from Notion and gives actionable advice.",
		Model:       m,
		Instruction: instruction(),
		Toolsets:    []tool.Toolset{toolset},
	})
	if err != nil {
		return nil, fmt.Errorf("init llmagent: %w", err)
	}

	sess := session.InMemoryService()
	run, err := runner.New(runner.Config{
		AppName:        "notion-todo-agent",
		Agent:          ag,
		SessionService: sess,
	})
	if err != nil {
		return nil, fmt.Errorf("init runner: %w", err)
	}

	return &App{
		agent:   ag,
		sess:    sess,
		run:     run,
		appName: "notion-todo-agent",
		userID:  "local-user",
		verbose: cfg.Verbose,
	}, nil
}

func (a *App) Ask(ctx context.Context, sessionID, userText string) (string, error) {
	userText = strings.TrimSpace(userText)
	if userText == "" {
		return "", nil
	}

	if err := a.ensureSession(ctx, sessionID); err != nil {
		return "", err
	}

	input := &genai.Content{
		Role:  genai.RoleUser,
		Parts: []*genai.Part{{Text: userText}},
	}

	seq := a.run.Run(ctx, a.userID, sessionID, input, agent.RunConfig{})

	var final *genai.Content
	for ev, err := range seq {
		if err != nil {
			return "", err
		}
		if ev == nil {
			continue
		}
		if a.verbose {
			printEvent(ev)
		}
		if ev.IsFinalResponse() {
			final = ev.Content
			break
		}
	}
	return strings.TrimSpace(contentText(final)), nil
}

func (a *App) ensureSession(ctx context.Context, sessionID string) error {
	_, err := a.sess.Get(ctx, &session.GetRequest{
		AppName:   a.appName,
		UserID:    a.userID,
		SessionID: sessionID,
	})
	if err == nil {
		return nil
	}
	_, err = a.sess.Create(ctx, &session.CreateRequest{
		AppName:   a.appName,
		UserID:    a.userID,
		SessionID: sessionID,
	})
	return err
}

func newNotionMCPToolset(cfg Config) (tool.Toolset, error) {
	var transport mcp.Transport
	token := strings.TrimSpace(cfg.NotionToken)
	httpClient := defaultHTTPClient(token)
	if httpClient == nil {
		return nil, errors.New("notion_token is required; pass it via -notion_token")
	}

	transport = &mcp.StreamableClientTransport{
		Endpoint:   cfg.NotionURL,
		HTTPClient: httpClient,
	}

	return mcptoolset.New(mcptoolset.Config{
		Transport:  transport,
		ToolFilter: isReadOnlyNotionTool,
	})
}

// isReadOnlyNotionTool best-effort filter that allows only read/query tools.
// Different Notion MCP servers expose slightly different tool names; we match by patterns.
func isReadOnlyNotionTool(ctx agent.ReadonlyContext, t tool.Tool) bool {
	_ = ctx
	n := strings.ToLower(t.Name())

	// Hard deny list.
	denySubstrings := []string{
		"delete", "update", "patch", "create", "append", "move", "archive", "restore", "set-", "write",
	}
	for _, d := range denySubstrings {
		if strings.Contains(n, d) {
			return false
		}
	}

	// Explicit allows (cover both notion-* and openapi-ish names).
	allowSubstrings := []string{
		// "notion-search",
		"notion-fetch",
		"search",            // e.g. post-search / search
		"retrieve",          // retrieve-a-page, retrieve-a-block, retrieve-a-data-source
		"get-",              // get-page, get-block-children, get-self
		"query-data-source", // notion task DB query
		"database-query",    // older servers
		"list",              // list-users, list-block-children
	}

	for _, a := range allowSubstrings {
		if strings.Contains(n, a) {
			return true
		}
	}
	return false
}

func instruction() string {
	return strings.TrimSpace(`You are an assistant for TODO tasks.

You have access to tools that can READ data from Notion (tasks, pages, databases). You MUST NOT modify Notion: no create, update, delete, move, archive, append, or any write operations.

When the user asks about their TODOs:
- Use notion_fetch to access https://www.notion.so/a-ast/b30702979243411fb71737dec7d10fa0?v=de9f6aecb5c947169eab5e1208a1fa6e
- Summarize the current state, list all fields of a TODO item
- List all tasks 
- Analyze all tasks together and look for patterns, for potential connections between tasks
- If information is missing, ask a single clarifying question

If the user requests any change in Notion, refuse politely and explain that you are read-only.`)
}

func contentText(c *genai.Content) string {
	if c == nil {
		return ""
	}
	var b strings.Builder
	for _, p := range c.Parts {
		if p == nil {
			continue
		}
		if p.Text != "" {
			b.WriteString(p.Text)
		}
	}
	return b.String()
}

func printEvent(ev *session.Event) {
	if ev == nil {
		return
	}
	if ev.Content != nil {
		fmt.Fprintf(os.Stderr, "[event] author=%s role=%s partial=%v final=%v\n", ev.Author, ev.Content.Role, ev.Partial, ev.IsFinalResponse())
	}
}

func defaultHTTPClient(token string) *http.Client {
	if token == "" {
		return nil
	}

	return &http.Client{
		Transport: &authTransport{
			base:  http.DefaultTransport,
			token: token,
		},
	}
}

type authTransport struct {
	base  http.RoundTripper
	token string
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cloned := req.Clone(req.Context())
	cloned.Header.Set("Authorization", "Bearer "+t.token)
	return t.base.RoundTrip(cloned)
}
