package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"time"

	"gomentum/internal/tui"
)

func main() {
	// Global panic handler to prevent window closing on crash
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
			fmt.Println("Press Enter to exit (or wait 30s)...")

			done := make(chan struct{})
			go func() {
				bufio.NewReader(os.Stdin).ReadString('\n')
				close(done)
			}()

			select {
			case <-done:
			case <-time.After(30 * time.Second):
			}
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
	fmt.Println("\nProgram finished. Press Enter to close window...")

	done := make(chan struct{})
	go func() {
		bufio.NewReader(os.Stdin).ReadString('\n')
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
	}
}
