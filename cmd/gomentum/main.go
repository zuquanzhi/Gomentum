package main

import (
	"fmt"
	"log/slog"
	"os"

	"gomentum/internal/tui"
)

func main() {
	// Global panic handler to prevent window closing on crash
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
			tui.WaitPressEnter()
		}
	}()

	// Initialize structured logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	fmt.Println("Gomentum: CLI Planning Agent")
	tui.Start()

	// Pause before exit to keep window open
	fmt.Println("\nProgram finished.")
	tui.WaitPressEnter()
}
