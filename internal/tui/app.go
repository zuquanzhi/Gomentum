package tui

import (
	"context"
	"fmt"
	"strings"

	"gomentum/internal/agent"
	"gomentum/internal/config"
	"gomentum/internal/planner"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#04B575")).
				Render

	errorMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF0000")).
				Render
)

// Task Item for List
type taskItem struct {
	id          int
	title       string
	description string
	status      string
	startTime   string
}

func (t taskItem) Title() string       { return t.title }
func (t taskItem) Description() string { return fmt.Sprintf("[%s] %s", t.startTime, t.description) }
func (t taskItem) FilterValue() string { return t.title }

type errMsg error

type model struct {
	viewport    viewport.Model
	textarea    textarea.Model
	taskList    list.Model
	senderStyle lipgloss.Style
	err         error

	// App state
	cfg     *config.Config
	planner *planner.Planner
	agent   agent.Agent

	// Chat state
	messages    []string
	isThinking  bool
	currentResp string

	// Streaming
	sub chan string

	// Layout
	width  int
	height int
}

func InitialModel(cfg *config.Config, p *planner.Planner, ag agent.Agent) model {
	ta := textarea.New()
	ta.Placeholder = "Ask Gomentum to plan your day..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to Gomentum!
Type a message to start planning.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	// Initialize Task List
	items := []list.Item{}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Tasks"
	l.SetShowHelp(false)

	return model{
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		taskList:    l,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
		cfg:         cfg,
		planner:     p,
		agent:       ag,
		sub:         make(chan string),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.refreshTasks)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		lCmd  tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.taskList, lCmd = m.taskList.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Layout: 30% Sidebar, 70% Chat
		sidebarWidth := int(float64(msg.Width) * 0.3)
		chatWidth := msg.Width - sidebarWidth - 4 // Margins

		m.taskList.SetSize(sidebarWidth, msg.Height-2)

		m.textarea.SetWidth(chatWidth)
		m.viewport.Width = chatWidth
		m.viewport.Height = msg.Height - m.textarea.Height() - 4

		m.renderChat()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.isThinking {
				return m, nil
			}

			input := m.textarea.Value()
			if strings.TrimSpace(input) == "" {
				return m, nil
			}

			m.messages = append(m.messages, "**You**: "+input)
			m.renderChat()
			m.textarea.Reset()
			m.viewport.GotoBottom()

			m.isThinking = true
			m.currentResp = ""
			m.sub = make(chan string) // Reset channel

			// Start agent interaction
			return m, tea.Batch(
				m.startChat(input),
				waitForActivity(m.sub),
			)
		}

	// We handle custom messages here for streaming
	case tokenMsg:
		m.currentResp += string(msg)
		m.renderChat()
		m.viewport.GotoBottom()
		return m, waitForActivity(m.sub) // Wait for next token

	case finishMsg:
		m.isThinking = false
		m.messages = append(m.messages, "**Gomentum**: "+m.currentResp)
		m.currentResp = ""
		// Refresh tasks after agent is done, as it might have changed them
		return m, m.refreshTasks

	case errMsg:
		m.err = msg
		return m, nil

	case []list.Item:
		m.taskList.SetItems(msg)
	}

	return m, tea.Batch(tiCmd, vpCmd, lCmd)
}

func (m model) View() string {
	chatView := fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		appStyle.Render(m.taskList.View()),
		appStyle.Render(chatView),
	)
}

func (m *model) renderChat() {
	content := strings.Join(m.messages, "\n\n")
	if m.currentResp != "" {
		content += "\n\n**Gomentum**: " + m.currentResp
	}

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(m.viewport.Width),
	)
	str, err := renderer.Render(content)
	if err != nil {
		m.viewport.SetContent(content)
	} else {
		m.viewport.SetContent(str)
	}
}

func (m model) refreshTasks() tea.Msg {
	tasks, err := m.planner.ListTasks()
	if err != nil {
		return errMsg(err)
	}

	items := []list.Item{}
	for _, t := range tasks {
		items = append(items, taskItem{
			id:          t.ID,
			title:       t.Title,
			description: t.Description,
			status:      t.Status,
			startTime:   t.StartTime.Local().Format("15:04"),
		})
	}
	return items
}

// Custom messages
type tokenMsg string
type finishMsg struct{}
type errorMsg error

func waitForActivity(sub chan string) tea.Cmd {
	return func() tea.Msg {
		token, ok := <-sub
		if !ok {
			return finishMsg{}
		}
		return tokenMsg(token)
	}
}

func (m model) startChat(input string) tea.Cmd {
	return func() tea.Msg {
		go func() {
			_, err := m.agent.Chat(context.Background(), input, func(token string) {
				m.sub <- token
			})
			if err != nil {
				// We can't easily send error to channel if it expects string
				// For now, just log or send as text
				m.sub <- fmt.Sprintf("\nError: %v", err)
			}
			close(m.sub)
		}()
		return nil // The actual messages come via waitForActivity subscription
	}
}
