package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gomentum/internal/planner"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server wraps the MCP server and the Planner
type Server struct {
	mcpServer *server.MCPServer
	planner   *planner.Planner
}

// NewServer creates a new MCP server instance
func NewServer(p *planner.Planner) *Server {
	s := server.NewMCPServer(
		"Gomentum Planner",
		"0.1.0",
	)

	srv := &Server{
		mcpServer: s,
		planner:   p,
	}

	srv.registerTools()
	return srv
}

func (s *Server) registerTools() {
	// Tool: current_time
	s.mcpServer.AddTool(mcp.NewTool("current_time",
		mcp.WithDescription("Return the current local time in RFC3339 format with timezone offset"),
	), s.handleCurrentTime)

	// Tool: add_task
	s.mcpServer.AddTool(mcp.NewTool("add_task",
		mcp.WithDescription("Add a new task to the schedule"),
		mcp.WithString("title", mcp.Required(), mcp.Description("The title of the task")),
		mcp.WithString("description", mcp.Description("Detailed description of the task")),
		mcp.WithString("start_time", mcp.Required(), mcp.Description("Start time in RFC3339 format (e.g. 2023-10-01T14:00:00Z)")),
		mcp.WithString("end_time", mcp.Required(), mcp.Description("End time in RFC3339 format")),
	), s.handleAddTask)

	// Tool: list_tasks
	s.mcpServer.AddTool(mcp.NewTool("list_tasks",
		mcp.WithDescription("List all scheduled tasks"),
	), s.handleListTasks)

	// Tool: export_tasks
	s.mcpServer.AddTool(mcp.NewTool("export_tasks",
		mcp.WithDescription("Export scheduled tasks to a markdown file"),
		mcp.WithString("filename", mcp.Description("The filename to save to (default: plan.md)")),
	), s.handleExportTasks)

	// Tool: update_task
	s.mcpServer.AddTool(mcp.NewTool("update_task",
		mcp.WithDescription("Update an existing task"),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("The ID of the task to update")),
		mcp.WithString("title", mcp.Description("The new title of the task")),
		mcp.WithString("description", mcp.Description("The new description")),
		mcp.WithString("start_time", mcp.Description("The new start time (RFC3339)")),
		mcp.WithString("end_time", mcp.Description("The new end time (RFC3339)")),
		mcp.WithString("status", mcp.Description("The new status (pending, completed, in_progress)")),
	), s.handleUpdateTask)

	// Tool: delete_task
	s.mcpServer.AddTool(mcp.NewTool("delete_task",
		mcp.WithDescription("Delete a task by ID"),
		mcp.WithNumber("id", mcp.Required(), mcp.Description("The ID of the task to delete")),
	), s.handleDeleteTask)
}

func (s *Server) handleCurrentTime(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	now := time.Now()
	payload := fmt.Sprintf(`{"local_time":"%s"}`, now.Format(time.RFC3339))
	return mcp.NewToolResultText(payload), nil
}

func (s *Server) handleAddTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	title, _ := args["title"].(string)
	desc, _ := args["description"].(string)
	startStr, _ := args["start_time"].(string)
	endStr, _ := args["end_time"].(string)

	startTime, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid start_time format: %v", err)), nil
	}

	endTime, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid end_time format: %v", err)), nil
	}

	// Check for overlap
	allowOverlap, _ := args["allow_overlap"].(bool)
	if !allowOverlap {
		conflict, err := s.planner.CheckOverlap(startTime, endTime, 0)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to check overlap: %v", err)), nil
		}
		if conflict != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Time conflict with existing task: '%s' (ID: %d) from %s to %s. Set allow_overlap=true to force.",
				conflict.Title, conflict.ID, conflict.StartTime.Format("15:04"), conflict.EndTime.Format("15:04"))), nil
		}
	}

	task, err := s.planner.AddTask(title, desc, startTime, endTime)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to add task: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Task added: ID=%d, Title=%s", task.ID, task.Title)), nil
}

func (s *Server) handleListTasks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tasks, err := s.planner.ListTasks()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list tasks: %v", err)), nil
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal tasks: %v", err)), nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

func (s *Server) handleExportTasks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := request.Params.Arguments.(map[string]interface{})
	filename, _ := args["filename"].(string)
	if filename == "" {
		filename = "plan.md"
	}

	if err := s.planner.ExportToMarkdown(filename); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to export tasks: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Tasks exported to %s", filename)), nil
}

func (s *Server) handleUpdateTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	idFloat, ok := args["id"].(float64)
	if !ok {
		return mcp.NewToolResultError("Task ID is required and must be a number"), nil
	}
	id := int(idFloat)

	// Get existing task
	task, err := s.planner.GetTask(id)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to find task: %v", err)), nil
	}

	// Update fields if provided
	if title, ok := args["title"].(string); ok && title != "" {
		task.Title = title
	}
	if desc, ok := args["description"].(string); ok {
		task.Description = desc
	}
	if status, ok := args["status"].(string); ok && status != "" {
		task.Status = status
	}
	if startStr, ok := args["start_time"].(string); ok && startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			task.StartTime = t
		}
	}
	if endStr, ok := args["end_time"].(string); ok && endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			task.EndTime = t
		}
	}

	// Check for overlap
	allowOverlap, _ := args["allow_overlap"].(bool)
	if !allowOverlap {
		conflict, err := s.planner.CheckOverlap(task.StartTime, task.EndTime, task.ID)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to check overlap: %v", err)), nil
		}
		if conflict != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Time conflict with existing task: '%s' (ID: %d) from %s to %s. Set allow_overlap=true to force.",
				conflict.Title, conflict.ID, conflict.StartTime.Format("15:04"), conflict.EndTime.Format("15:04"))), nil
		}
	}

	if err := s.planner.UpdateTask(task); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update task: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Task %d updated successfully", id)), nil
}

func (s *Server) handleDeleteTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	idFloat, ok := args["id"].(float64)
	if !ok {
		return mcp.NewToolResultError("Task ID is required and must be a number"), nil
	}
	id := int(idFloat)

	if err := s.planner.DeleteTask(id); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete task: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Task %d deleted successfully", id)), nil
}

// GetTools returns the list of tool definitions (helper for the Agent)
// In a real MCP setup, the client would discover these via the protocol.
// Here we expose them directly to bridge to the OpenAI Agent.
func (s *Server) GetTools() []mcp.Tool {
	// Accessing the internal tools map is not directly exposed by the high-level server struct in some versions,
	// but let's see if we can reconstruct them or if we need to store them separately.
	// For now, let's just return the definitions we know we added.
	// Ideally, we should ask the mcpServer.

	// Since mark3labs/mcp-go server might not expose a simple "GetTools" list for local consumption easily without reflection or private access,
	// we will manually reconstruct the definitions for the Agent to consume.

	return []mcp.Tool{
		mcp.NewTool("current_time",
			mcp.WithDescription("Return the current local time in RFC3339 format with timezone offset"),
		),
		mcp.NewTool("add_task",
			mcp.WithDescription("Add a new task to the schedule"),
			mcp.WithString("title", mcp.Required(), mcp.Description("The title of the task")),
			mcp.WithString("description", mcp.Description("Detailed description of the task")),
			mcp.WithString("start_time", mcp.Required(), mcp.Description("Start time in RFC3339 format (e.g. 2023-10-01T14:00:00Z)")),
			mcp.WithString("end_time", mcp.Required(), mcp.Description("End time in RFC3339 format")),
			mcp.WithBoolean("allow_overlap", mcp.Description("Set to true to allow scheduling even if there is a conflict")),
		),
		mcp.NewTool("list_tasks",
			mcp.WithDescription("List all scheduled tasks"),
		),
		mcp.NewTool("export_tasks",
			mcp.WithDescription("Export scheduled tasks to a markdown file"),
			mcp.WithString("filename", mcp.Description("The filename to save to (default: plan.md)")),
		),
		mcp.NewTool("update_task",
			mcp.WithDescription("Update an existing task"),
			mcp.WithNumber("id", mcp.Required(), mcp.Description("The ID of the task to update")),
			mcp.WithString("title", mcp.Description("The new title of the task")),
			mcp.WithString("description", mcp.Description("The new description")),
			mcp.WithString("start_time", mcp.Description("The new start time (RFC3339)")),
			mcp.WithString("end_time", mcp.Description("The new end time (RFC3339)")),
			mcp.WithString("status", mcp.Description("The new status (pending, completed, in_progress)")),
			mcp.WithBoolean("allow_overlap", mcp.Description("Set to true to allow scheduling even if there is a conflict")),
		),
		mcp.NewTool("delete_task",
			mcp.WithDescription("Delete a task by ID"),
			mcp.WithNumber("id", mcp.Required(), mcp.Description("The ID of the task to delete")),
		),
	}
}

// CallTool directly calls a tool (helper for the Agent)
func (s *Server) CallTool(ctx context.Context, name string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	// We need to construct a CallToolRequest
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	}

	// We need to route this to the handler.
	// The mcpServer handles JSON-RPC, but we want to call the handler directly.
	// Let's switch on name for now since we are bridging locally.

	switch name {
	case "current_time":
		return s.handleCurrentTime(ctx, req)
	case "add_task":
		return s.handleAddTask(ctx, req)
	case "list_tasks":
		return s.handleListTasks(ctx, req)
	case "export_tasks":
		return s.handleExportTasks(ctx, req)
	case "update_task":
		return s.handleUpdateTask(ctx, req)
	case "delete_task":
		return s.handleDeleteTask(ctx, req)
	default:
		return nil, fmt.Errorf("tool not found: %s", name)
	}
}
