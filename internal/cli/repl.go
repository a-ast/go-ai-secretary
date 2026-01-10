package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	agentpkg "github.com/a-ast/go-ai-assistant/internal/agent"
)

/*
* Implements a Read-Eval-Print Loop (REPL) for the AI assistant.
 */
func RunREPL(ctx context.Context, app *agentpkg.App, sessionID string) error {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			return scanner.Err()
		}
		line := strings.TrimSpace(scanner.Text())
		switch line {
		case "", "\\":
			continue
		case "/exit", "/quit", "/q":
			return nil
		case "/help":
			printHelp()
			continue
		}
		answer, err := app.Ask(ctx, sessionID, line)
		if err != nil {
			return err
		}
		if answer != "" {
			fmt.Println(answer)
		}
	}
}

func printHelp() {
	fmt.Println("Commands:")
	fmt.Println("  /help   show this help")
	fmt.Println("  /exit   quit")
	fmt.Println("Tips:")
	fmt.Println("  Ask: 'what are my tasks for today?'")
}
