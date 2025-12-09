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

// WaitPressEnter pauses execution to allow user to read output before window closes
func WaitPressEnter() {
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
}

// Start launches the Bubble Tea TUI for Gomentum
func Start() {
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
		WaitPressEnter()
		os.Exit(1)
	}

	// Initialize Planner
	p, err := planner.NewPlanner(cfg.Database.Path)
	if err != nil {
		slog.Error("Failed to initialize planner", "error", err)
		fmt.Printf("\nError initializing database: %v\n", err)
		WaitPressEnter()
		os.Exit(1)
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
		WaitPressEnter()
		os.Exit(1)
	}

	// Start background reminder
	go startReminder(p)

	// Start Bubble Tea Program
	// Note: WithAltScreen might cause issues if the terminal closes immediately after exit.
	// But for a TUI app, it's standard.
	prog := tea.NewProgram(InitialModel(cfg, p, ag), tea.WithAltScreen())
	if _, err := prog.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		WaitPressEnter()
		os.Exit(1)
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
			// Send system notification
			msg := fmt.Sprintf("Time: %s\n%s", t.StartTime.Local().Format("15:04"), t.Description)
			if err := beeep.Notify("Gomentum Reminder", msg, ""); err != nil {
				// Silently fail or log to file if needed, but don't print to stdout
				slog.Error("System notification failed", "error", err)
			}

			// Mark as reminded
			_ = p.MarkAsReminded(t.ID)
		}
	}
}
