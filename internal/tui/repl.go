package tui

import (
	"bufio"
	"fmt"
	"gomentum/internal/agent"
	"gomentum/internal/config"
	"gomentum/internal/mcp"
	"gomentum/internal/planner"
	"log/slog"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gen2brain/beeep"
)

// Start launches the Bubble Tea TUI for Gomentum
func Start() {
	// Helper to pause before exit
	waitExit := func() {
		fmt.Println("\nPress Enter to exit (or wait 30 seconds)...")

		// Force a small sleep to prevent immediate skipping if there's buffered input
		time.Sleep(500 * time.Millisecond)

		done := make(chan struct{})
		go func() {
			_, err := bufio.NewReader(os.Stdin).ReadString('\n')
			if err != nil {
				// If reading fails (e.g. no stdin), wait for the timeout
				return
			}
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(30 * time.Second):
		}
		os.Exit(1)
	}

	// Load Config
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		cwd, _ := os.Getwd()
		exe, _ := os.Executable()
		slog.Error("Failed to load config", "error", err, "cwd", cwd, "exe", exe)
		fmt.Printf("\nError loading config.yaml: %v\n", err)
		fmt.Printf("Current Working Directory: %s\n", cwd)
		fmt.Printf("Executable Path: %s\n", exe)
		fmt.Println("Please ensure 'config.yaml' exists in the Current Working Directory.")
		waitExit()
	}

	// Initialize Planner
	p, err := planner.NewPlanner(cfg.Database.Path)
	if err != nil {
		slog.Error("Failed to initialize planner", "error", err)
		fmt.Printf("\nError initializing database: %v\n", err)
		waitExit()
	}
	defer p.Close()

	// Initialize MCP Server
	ms := mcp.NewServer(p)

	// Initialize Agent
	ag, err := agent.NewAgent(cfg, ms, p)
	if err != nil {
		slog.Error("Failed to initialize agent", "error", err)
		fmt.Printf("\nError initializing agent: %v\n", err)
		fmt.Println("Please check your configuration (API Key, etc).")
		waitExit()
	}

	// Start background reminder
	go startReminder(p)

	// Start Bubble Tea Program
	// Note: WithAltScreen might cause issues if the terminal closes immediately after exit.
	// But for a TUI app, it's standard.
	prog := tea.NewProgram(InitialModel(cfg, p, ag), tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		waitExit()
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

// handleInput is removed as it is now handled by the Bubble Tea model
