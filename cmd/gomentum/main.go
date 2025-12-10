package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

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

	// Initialize structured logging to file
	var logOutput io.Writer = os.Stdout // Default fallback

	homeDir, err := os.UserHomeDir()
	if err == nil {
		logDir := filepath.Join(homeDir, ".gomentum")
		if err := os.MkdirAll(logDir, 0755); err == nil {
			logPath := filepath.Join(logDir, "gomentum.log")
			if f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				// We don't defer f.Close() here because we want it open for the lifetime of the app
				// and OS will close it on exit.
				logOutput = f
			}
		}
	}

	// If we failed to open file, use io.Discard to avoid messing up TUI
	if logOutput == os.Stdout {
		logOutput = io.Discard
	}

	logger := slog.New(slog.NewTextHandler(logOutput, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	fmt.Println("Gomentum: CLI Planning Agent")
	tui.Start()

	// Pause before exit to keep window open
	fmt.Println("\nProgram finished.")
	tui.WaitPressEnter()
}
