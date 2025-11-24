# Gomentum

Gomentum is a CLI-based planning agent powered by LLMs.

It acts as an intelligent task decomposer rather than a simple todo list. By leveraging the **Google adk-go** (Agent Development Kit) and **Model Context Protocol (MCP)**, it connects a reasoning engine with local execution tools to break down high-level goals into actionable schedules.

This project serves as a practical implementation of:
- **Google adk-go**: Cutting-edge framework for building AI agents in Go.
- **Model Context Protocol (MCP)**: Standardized interface for LLM tools.
- **Go Concurrency Patterns**: Background task management.
- **Clean Architecture**: Maintainable CLI structure.

## Status

**WIP** - Pre-alpha.

## Features

- **Natural Language Interface**: REPL-based interaction.
- **Agentic Workflow**: Built with `adk-go` to handle complex reasoning and task decomposition.
- **MCP Native**: Strictly follows the Model Context Protocol specification.
- **Flexible LLM Support**: Compatible with Gemini, OpenAI, or any OpenAI-compatible endpoint.
- **Concurrency**: Background timers and notifications using Go routines.

## Tech Stack

- **Go 1.23+**
- **Google adk-go** (Agent Framework)
- **mark3labs/mcp-go** (MCP Implementation)

## Project Structure

```text
gomentum/
├── cmd/
│   └── gomentum/        # Entry point
├── internal/
│   ├── agent/           # LLM integration
│   ├── mcp/             # MCP server & tool definitions
│   ├── planner/         # Core domain logic
│   └── tui/             # Terminal UI / REPL
├── pkg/                 # Shared libraries
└── README.md
```

## Quick Start

1.  **Clone & Init**
    ```bash
    git clone https://github.com/zuquanzhi/gomentum.git
    cd gomentum
    go mod tidy
    ```

2.  **Config**
    Configure your LLM provider. You can use Gemini, OpenAI, or any custom endpoint.

    **Example: DeepSeek**
    ```bash
    # Linux/macOS
    export LLM_API_KEY="sk-..."
    export LLM_BASE_URL="https://api.deepseek.com/v1"
    export LLM_MODEL="deepseek-chat"
    
    # Windows PowerShell
    $env:LLM_API_KEY="sk-..."
    $env:LLM_BASE_URL="https://api.deepseek.com/v1"
    $env:LLM_MODEL="deepseek-chat"
    ```

    **Example: Gemini**
    ```bash
    # Linux/macOS
    export LLM_API_KEY="your_gemini_key"
    # LLM_BASE_URL is not needed for Gemini default
    
    # Windows PowerShell
    $env:LLM_API_KEY="your_gemini_key"
    ```

3.  **Run**
    ```bash
    go run cmd/gomentum/main.go
    ```

## Todo

### Phase 1: Foundation
- [x] Initialize Go module and directory structure.
- [x] Implement basic REPL loop (read-eval-print).
- [ ] Integrate `adk-go` for agent reasoning.

### Phase 2: Core Logic (MCP)
- [ ] Define MCP server structure.
- [ ] Implement `list_tools` and `call_tool` handlers.
- [ ] Create `Planner` struct for in-memory task management.
- [ ] Implement `add_task` and `list_tasks` tools.

### Phase 3: Concurrency & IO
- [ ] Implement background worker for task reminders (Goroutines).
- [ ] Add file persistence (save/load tasks to JSON/Markdown).
- [ ] Handle graceful shutdown and context cancellation.

### Phase 4: Refinement
- [ ] Improve prompt engineering for task decomposition.
- [ ] Add CLI flags for configuration.
- [ ] Write unit tests for core logic.

## License

MIT

