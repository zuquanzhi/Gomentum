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
	"time"

	"github.com/gen2brain/beeep"
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

	// Start background reminder
	go startReminder(p)

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

func startReminder(p *planner.Planner) {
	// Check every 10 seconds for better responsiveness
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Find tasks that are due now (or past due)
		// We pass 0 duration because we want to trigger exactly at StartTime,
		// not 15 minutes before.
		tasks, err := p.GetUpcomingTasks(0)
		if err != nil {
			continue
		}

		for _, t := range tasks {
			// Print notification
			// Use \r to overwrite the prompt, then reprint prompt
			fmt.Printf("\n\n🔔 [REMINDER] Task '%s' starts at %s!\n", t.Title, t.StartTime.Local().Format("15:04"))
			if t.Description != "" {
				fmt.Printf("   %s\n", t.Description)
			}
			fmt.Print("\nGomentum > ")

			// Send system notification
			msg := fmt.Sprintf("Time: %s\n%s", t.StartTime.Local().Format("15:04"), t.Description)
			if err := beeep.Notify("Gomentum Reminder", msg, ""); err != nil {
				fmt.Printf("\n[System Notification Failed]: %v\n", err)
			}

			// Mark as reminded
			_ = p.MarkAsReminded(t.ID)
		}
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
