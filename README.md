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

## Roadmap

### Phase 1: Foundation (Completed)
- [x] **Core Agent**: REPL loop with LLM integration (DeepSeek/OpenAI).
- [x] **MCP Implementation**: Full Model Context Protocol server with tools (`add_task`, `list_tasks`, etc.).
- [x] **Persistence**: SQLite database for robust task storage.
- [x] **Concurrency**: Background worker for task monitoring.
- [x] **Notifications**: Cross-platform system notifications (Toast/Banner).
- [x] **Conflict Detection**: Smart scheduling with overlap warnings.

### Phase 2: Backend Engineering (In Progress)
Transforming Gomentum into a production-grade backend system.

#### 1. Architecture Upgrade (Clean Architecture)
- [ ] **Refactor**: Decouple UI/CLI from business logic.
- [ ] **Layers**: Implement strict Handler -> Service -> Repository layering.
- [ ] **Dependency Injection**: Improve testability and modularity.

#### 2. Storage & Data Integrity
- [ ] **PostgreSQL**: Migrate from SQLite to Dockerized PostgreSQL.
- [ ] **Migrations**: Implement schema version control (e.g., `golang-migrate`).
- [ ] **ORM/Builder**: Switch to `sqlx` or `GORM` for better query management.

#### 3. Distributed Systems & Performance
- [ ] **Redis**: Introduce Redis for caching and task queuing.
- [ ] **Asynq**: Replace polling loop with a distributed task queue (`hibiken/asynq`) for precise delayed execution.
- [ ] **Scalability**: Decouple the scheduler from the executor.

#### 4. API & Interface
- [ ] **RESTful API**: Expose functionality via HTTP (Gin/Echo).
- [ ] **gRPC**: Implement high-performance RPC endpoints.
- [ ] **Headless Mode**: Run as a background daemon/service.

#### 5. Observability (DevOps)
- [ ] **Structured Logging**: Implement `uber-go/zap` for JSON logs.
- [ ] **Metrics**: Expose Prometheus metrics (task counts, latency, queue depth).
- [ ] **Tracing**: Add OpenTelemetry for request tracing.

#### 6. Infrastructure
- [ ] **Docker**: Containerize the application, DB, and Redis.
- [ ] **CI/CD**: GitHub Actions for automated testing and linting.

## License

MIT

