package tui

import (
	"bufio"
	"context"
	"fmt"
	"gomentum/internal/agent"
	"gomentum/internal/mcp"
	"gomentum/internal/planner"
	"os"
	"strings"
)

// Start launches the Read-Eval-Print Loop for Gomentum
func Start() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Welcome to Gomentum. Type 'exit' or 'quit' to leave.")

	// Initialize Planner
	p, err := planner.NewPlanner("gomentum.db")
	if err != nil {
		fmt.Printf("Error: Failed to initialize planner: %v\n", err)
		return
	}
	defer p.Close()

	// Initialize MCP Server
	ms := mcp.NewServer(p)

	// Initialize Agent
	ag, err := agent.NewAgent(ms)
	if err != nil {
		fmt.Printf("Error: Failed to initialize agent: %v\n", err)
		fmt.Println("Please set LLM_API_KEY environment variable.")
		return
	}

	for {
		fmt.Print("Gomentum > ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Bye!")
			break
		}

		handleInput(context.Background(), ag, input)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
	}
}

func handleInput(ctx context.Context, ag agent.Agent, input string) {
	fmt.Print("Thinking...")
	resp, err := ag.Chat(ctx, input)
	if err != nil {
		fmt.Printf("\rError: %v\n", err)
		return
	}
	fmt.Printf("\r%s\n", resp)
}
