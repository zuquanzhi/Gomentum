package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	gmcp "gomentum/internal/mcp"

	"github.com/mark3labs/mcp-go/mcp"
	openai "github.com/sashabaranov/go-openai"
)

// Agent defines the interface for our planning agent
type Agent interface {
	// Chat sends a message to the agent and returns the response
	Chat(ctx context.Context, prompt string) (string, error)
}

// OpenAIAgent implements Agent for OpenAI-compatible APIs (e.g., DeepSeek)
type OpenAIAgent struct {
	client    *openai.Client
	model     string
	mcpServer *gmcp.Server
	history   []openai.ChatCompletionMessage
}

// NewAgent creates a new agent based on environment variables
func NewAgent(mcpServer *gmcp.Server) (Agent, error) {
	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("LLM_API_KEY is not set")
	}

	baseURL := os.Getenv("LLM_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1" // Default to DeepSeek for now
	}

	model := os.Getenv("LLM_MODEL")
	if model == "" {
		model = "deepseek-chat"
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL

	client := openai.NewClientWithConfig(config)

	return &OpenAIAgent{
		client:    client,
		model:     model,
		mcpServer: mcpServer,
		history: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are Gomentum, a helpful planning assistant.", // Placeholder, updated in Chat
			},
		},
	}, nil
}

// Chat implements the Agent interface
func (a *OpenAIAgent) Chat(ctx context.Context, prompt string) (string, error) {
	// Update system prompt with current time
	if len(a.history) > 0 && a.history[0].Role == openai.ChatMessageRoleSystem {
		now := time.Now()
		a.history[0].Content = fmt.Sprintf("You are Gomentum, a helpful planning assistant. The current local time is %s. When scheduling tasks, use this time as reference. IMPORTANT: When calling tools with start_time or end_time, you MUST use RFC3339 format with the SAME timezone offset as the current time (e.g. if current time is +08:00, use +08:00). Do not convert to UTC. If the user provides a relative time (like 'tomorrow', 'next Monday'), calculate the absolute date and EXECUTE the tool immediately. Do not ask for confirmation unless the time is ambiguous. Be concise.", now.Format(time.RFC3339))
	}

	// Add user message to history
	a.history = append(a.history, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})

	// Prepare tools
	tools := a.getOpenAITools()

	// Loop to handle tool calls
	for {
		resp, err := a.client.CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model:    a.model,
				Messages: a.history,
				Tools:    tools,
			},
		)

		if err != nil {
			return "", err
		}

		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("no response from LLM")
		}

		msg := resp.Choices[0].Message
		a.history = append(a.history, msg)

		// If there are no tool calls, we are done
		if len(msg.ToolCalls) == 0 {
			return msg.Content, nil
		}

		// Handle tool calls
		for _, toolCall := range msg.ToolCalls {
			fmt.Printf("\n[Agent] Calling tool: %s\n", toolCall.Function.Name)

			var args map[string]interface{}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				return "", fmt.Errorf("failed to parse tool arguments: %v", err)
			}

			result, err := a.mcpServer.CallTool(ctx, toolCall.Function.Name, args)
			content := ""
			if err != nil {
				content = fmt.Sprintf("Error: %v", err)
			} else {
				// MCP result can be text or image, we assume text for now
				// The result content is a list of Content objects
				for _, c := range result.Content {
					if textContent, ok := c.(mcp.TextContent); ok {
						content += textContent.Text + "\n"
					}
				}
			}

			a.history = append(a.history, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    content,
				ToolCallID: toolCall.ID,
			})
		}
		// Loop continues to send tool results back to LLM
	}
}

func (a *OpenAIAgent) getOpenAITools() []openai.Tool {
	mcpTools := a.mcpServer.GetTools()
	var tools []openai.Tool

	for _, t := range mcpTools {
		tools = append(tools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.InputSchema,
			},
		})
	}
	return tools
}
