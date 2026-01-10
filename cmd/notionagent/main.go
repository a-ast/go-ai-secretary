package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	agentpkg "github.com/a-ast/go-ai-assistant/internal/agent"
	"github.com/a-ast/go-ai-assistant/internal/cli"
)

func main() {
	const defaultNotionURL = "https://mcp.notion.com/mcp"
	var (
		modelName   = flag.String("model", "gemini-2.5-flash", "Gemini model name, e.g. gemini-2.5-flash")
		notionToken = flag.String("notion_token", "", "Notion access token (see docs)")
		sessionID   = flag.String("session", "default", "Session ID (keeps chat history in-memory per session)")
		verbose     = flag.Bool("v", false, "Verbose: print tool activity")
	)
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Ctrl+C handling.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	app, err := agentpkg.NewApp(ctx, agentpkg.Config{
		ModelName:   *modelName,
		NotionURL:   defaultNotionURL,
		NotionToken: *notionToken,
		Verbose:     *verbose,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "init error:", err)
		os.Exit(1)
	}

	fmt.Println("AI Assistant for current tasks. Type your question. Commands: /help, /exit")
	if err := cli.RunREPL(ctx, app, *sessionID); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
