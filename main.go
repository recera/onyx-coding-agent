package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	graph "github.com/onyx/onyx-tui/graph_service"
)

// Message types for communication with the agent
type MessageType string

const (
	MsgInit         MessageType = "init"
	MsgChat         MessageType = "chat"
	MsgError        MessageType = "error"
	MsgResponse     MessageType = "response"
	MsgToolCall     MessageType = "tool_call"
	MsgStreamChunk  MessageType = "stream_chunk"
	MsgRunCypher    MessageType = "run_cypher"
	MsgCypherResult MessageType = "cypher_result"
)

type AgentMessage struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// App states
type AppState int

const (
	StateAPIKey AppState = iota
	StateChat
)

// Chat message for display
type ChatMessage struct {
	Role      string // "user", "assistant", "system", "tool"
	Content   string
	Timestamp time.Time
	IsError   bool
}

// Tool call information
type ToolCallInfo struct {
	ToolName string                 `json:"toolName"`
	Args     map[string]interface{} `json:"args"`
}

// Model for our TUI application
type Model struct {
	state        AppState
	apiKeyInput  textinput.Model
	chatInput    textarea.Model
	viewport     viewport.Model
	messages     []ChatMessage
	agentProcess *exec.Cmd
	agentStdin   io.WriteCloser
	agentStdout  io.ReadCloser
	agentReady   bool
	width        int
	height       int
	err          error
	isProcessing bool
	graphResult  *graph.BuildGraphResult // Store the graph database connection
	workDir      string                  // Store the working directory
}

// Styles
var (
	appStyle = lipgloss.NewStyle().
			Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			PaddingBottom(1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	userMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#86EFAC")).
			Bold(true)

	assistantMsgStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#93C5FD"))

	toolMsgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FDE047")).
			Italic(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	inputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)
)

// Messages for the update loop
type agentResponseMsg struct {
	message AgentMessage
}

type errMsg struct {
	err error
}

func initialModel() Model {
	// API Key input
	ti := textinput.New()
	ti.Placeholder = "sk-..."
	ti.Focus()
	ti.CharLimit = 200 // Increased to handle longer keys
	ti.Width = 80      // Increased width to show more of the key
	ti.EchoMode = textinput.EchoPassword

	// Chat input
	ta := textarea.New()
	ta.Placeholder = "Type your message..."
	ta.CharLimit = 500
	ta.SetWidth(60)
	ta.SetHeight(3)

	// Viewport for messages
	vp := viewport.New(80, 20)
	vp.SetContent("Welcome to Onyx AI TUI!\n\nPlease enter your OpenAI API key to begin.")

	workDir := os.Getenv("ONYX_WORK_DIR")
	if workDir == "" {
		if wd, err := os.Getwd(); err == nil {
			workDir = wd
		} else {
			workDir = "."
		}
	}

	return Model{
		state:       StateAPIKey,
		apiKeyInput: ti,
		chatInput:   ta,
		viewport:    vp,
		messages:    []ChatMessage{},
		agentReady:  false,
		workDir:     workDir,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

type agentStartedMsg struct {
	process *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
}

type graphBuiltMsg struct {
	result *graph.BuildGraphResult
	err    error
}

type cypherResultMsg struct {
	requestID string
	result    string
	err       error
}

func (m Model) startAgent(apiKey string) tea.Cmd {
	return func() tea.Msg {
		// Check if agent directory exists
		if _, err := os.Stat("./agent"); os.IsNotExist(err) {
			return errMsg{err: fmt.Errorf("agent directory not found. Please run from the onyx-tui directory")}
		}

		// Start the TypeScript agent process
		cmd := exec.Command("npm", "start")
		cmd.Dir = "./agent"

		// Set up pipes
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to create stdin pipe: %w", err)}
		}

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return errMsg{err: fmt.Errorf("failed to create stdout pipe: %w", err)}
		}

		// Redirect stderr to a file for debugging
		errFile, err := os.Create("agent-error.log")
		if err == nil {
			cmd.Stderr = errFile
		}

		// Start the process
		if err := cmd.Start(); err != nil {
			return errMsg{err: fmt.Errorf("failed to start agent: %w", err)}
		}

		// Send initialization message
		initMsg := AgentMessage{
			Type: MsgInit,
			Data: json.RawMessage(fmt.Sprintf(`{"apiKey":"%s"}`, apiKey)),
		}

		msgBytes, _ := json.Marshal(initMsg)
		stdin.Write(msgBytes)
		stdin.Write([]byte("\n"))

		// Return message with the process info
		return agentStartedMsg{
			process: cmd,
			stdin:   stdin,
			stdout:  stdout,
		}
	}
}

func (m Model) listenToAgent() tea.Cmd {
	return func() tea.Msg {
		if m.agentStdout == nil {
			return errMsg{err: fmt.Errorf("agent stdout is nil")}
		}

		scanner := bufio.NewScanner(m.agentStdout)
		// Process ONE message and return it
		// The Update function will call listenToAgent again to continue
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			// Log raw output for debugging
			log.Printf("Agent output: %s", line)

			var msg AgentMessage
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				// Log non-JSON lines for debugging
				log.Printf("Non-JSON output from agent: %s", line)
				continue
			}

			// Return this message and let Update reschedule listening
			return agentResponseMsg{message: msg}
		}

		if err := scanner.Err(); err != nil {
			return errMsg{err: fmt.Errorf("agent stream error: %w", err)}
		}

		// Agent closed the stream
		return errMsg{err: fmt.Errorf("agent stream closed unexpectedly")}
	}
}

func (m Model) sendChatMessage(message string) tea.Cmd {
	return func() tea.Msg {
		if m.agentStdin == nil {
			return errMsg{err: fmt.Errorf("agent not initialized")}
		}

		chatMsg := AgentMessage{
			Type: MsgChat,
			Data: json.RawMessage(fmt.Sprintf(`{"message":"%s"}`, strings.ReplaceAll(message, "\"", "\\\""))),
		}

		msgBytes, _ := json.Marshal(chatMsg)
		m.agentStdin.Write(msgBytes)
		m.agentStdin.Write([]byte("\n"))

		return nil
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			// Clean shutdown
			if m.agentProcess != nil {
				m.agentProcess.Process.Kill()
			}
			if m.graphResult != nil {
				m.graphResult.Close()
			}
			return m, tea.Quit

		case tea.KeyEnter:
			if m.state == StateAPIKey && !m.isProcessing {
				apiKey := m.apiKeyInput.Value()
				if apiKey != "" {
					m.isProcessing = true
					cmds = append(cmds, m.startAgent(apiKey))
				}
			} else if m.state == StateChat && !strings.Contains(m.chatInput.Value(), "\n") {
				// Send message on Enter if not in multiline mode (no newlines present)
				message := strings.TrimSpace(m.chatInput.Value())
				if message != "" && m.agentReady && !m.isProcessing {
					m.messages = append(m.messages, ChatMessage{
						Role:      "user",
						Content:   message,
						Timestamp: time.Now(),
					})
					m.chatInput.Reset()
					m.isProcessing = true
					m.updateViewport()
					cmds = append(cmds, m.sendChatMessage(message))
				}
			}

		case tea.KeyCtrlS:
			// Send message with Ctrl+S in chat mode
			if m.state == StateChat {
				message := strings.TrimSpace(m.chatInput.Value())
				if message != "" && m.agentReady && !m.isProcessing {
					m.messages = append(m.messages, ChatMessage{
						Role:      "user",
						Content:   message,
						Timestamp: time.Now(),
					})
					m.chatInput.Reset()
					m.isProcessing = true
					m.updateViewport()
					cmds = append(cmds, m.sendChatMessage(message))
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update viewport size
		headerHeight := 6
		footerHeight := 8
		m.viewport.Width = m.width - 4
		m.viewport.Height = m.height - headerHeight - footerHeight

		// Update chat input width
		m.chatInput.SetWidth(m.width - 8)

		m.updateViewport()

	case agentStartedMsg:
		// Agent process started successfully
		m.agentProcess = msg.process
		m.agentStdin = msg.stdin
		m.agentStdout = msg.stdout

		// Start listening to agent output
		cmds = append(cmds, m.listenToAgent())

		// Start building the graph database
		cmds = append(cmds, m.buildGraph())

	case agentResponseMsg:
		// Continue listening
		cmds = append(cmds, m.listenToAgent())

		switch msg.message.Type {
		case MsgResponse:
			var respData struct {
				Status  string `json:"status"`
				Message string `json:"message"`
				Content string `json:"content"`
			}
			json.Unmarshal(msg.message.Data, &respData)

			if respData.Status == "initialized" {
				m.state = StateChat
				m.agentReady = true
				m.isProcessing = false
				m.chatInput.Focus()
				m.messages = append(m.messages, ChatMessage{
					Role:      "system",
					Content:   "✓ Agent initialized successfully! You can now start by sending a message.",
					Timestamp: time.Now(),
				})
			} else if respData.Content != "" {
				// Check if we already have this content from streaming
				// If the last message is from assistant with the same content, don't duplicate
				if len(m.messages) > 0 &&
					m.messages[len(m.messages)-1].Role == "assistant" &&
					m.messages[len(m.messages)-1].Content == respData.Content {
					// Already have this content from streaming, just update processing state
					m.isProcessing = false
				} else if len(m.messages) == 0 || m.messages[len(m.messages)-1].Role != "assistant" {
					// No assistant message yet, add it
					m.messages = append(m.messages, ChatMessage{
						Role:      "assistant",
						Content:   respData.Content,
						Timestamp: time.Now(),
					})
					m.isProcessing = false
				} else {
					// Just mark as done
					m.isProcessing = false
				}
			}
			m.updateViewport()

		case MsgToolCall:
			var toolData ToolCallInfo
			json.Unmarshal(msg.message.Data, &toolData)

			// Create a concise single-line format for tool calls
			// Format: "Tool: tool_name | param1: value1, param2: value2"
			toolMsg := fmt.Sprintf("🔧 Tool: %s", toolData.ToolName)

			// Add key parameters in a compact format
			if len(toolData.Args) > 0 {
				var params []string

				// Special handling for common tools to show most relevant info
				switch toolData.ToolName {
				case "read_file", "write_file", "edit_file":
					if filePath, ok := toolData.Args["filePath"].(string); ok {
						params = append(params, fmt.Sprintf("file: %s", filePath))
					}
				case "list_files":
					if dirPath, ok := toolData.Args["dirPath"].(string); ok {
						params = append(params, fmt.Sprintf("dir: %s", dirPath))
					}
					if recursive, ok := toolData.Args["recursive"].(bool); ok && recursive {
						params = append(params, "recursive")
					}
				case "search_files":
					if pattern, ok := toolData.Args["pattern"].(string); ok {
						params = append(params, fmt.Sprintf("pattern: \"%s\"", pattern))
					}
					if dir, ok := toolData.Args["directory"].(string); ok && dir != "." {
						params = append(params, fmt.Sprintf("in: %s", dir))
					}
				case "run_command":
					if cmd, ok := toolData.Args["command"].(string); ok {
						// Truncate long commands
						if len(cmd) > 50 {
							params = append(params, fmt.Sprintf("cmd: %s...", cmd[:50]))
						} else {
							params = append(params, fmt.Sprintf("cmd: %s", cmd))
						}
					}
				case "run_cypher":
					if query, ok := toolData.Args["query"].(string); ok {
						// Truncate long queries and format nicely
						if len(query) > 60 {
							params = append(params, fmt.Sprintf("query: %s...", query[:60]))
						} else {
							params = append(params, fmt.Sprintf("query: %s", query))
						}
					}
				case "web_search":
					if query, ok := toolData.Args["query"].(string); ok {
						params = append(params, fmt.Sprintf("query: \"%s\"", query))
					}
					if limit, ok := toolData.Args["limit"].(float64); ok && limit != 3 {
						params = append(params, fmt.Sprintf("limit: %d", int(limit)))
					}
					if scrape, ok := toolData.Args["scrape"].(bool); ok && scrape {
						params = append(params, "scrape: true")
					}
				case "url_extract":
					if url, ok := toolData.Args["url"].(string); ok {
						// Truncate long URLs
						if len(url) > 50 {
							params = append(params, fmt.Sprintf("url: %s...", url[:50]))
						} else {
							params = append(params, fmt.Sprintf("url: %s", url))
						}
					}
					if formats, ok := toolData.Args["formats"].([]interface{}); ok && len(formats) > 0 {
						formatStrs := make([]string, 0)
						for _, f := range formats {
							if fStr, ok := f.(string); ok {
								formatStrs = append(formatStrs, fStr)
							}
						}
						if len(formatStrs) > 0 {
							params = append(params, fmt.Sprintf("formats: [%s]", strings.Join(formatStrs, ",")))
						}
					}
				default:
					// Generic handling for unknown tools
					for key, value := range toolData.Args {
						var valueStr string
						switch v := value.(type) {
						case string:
							if len(v) > 30 {
								valueStr = fmt.Sprintf("\"%s...\"", v[:30])
							} else {
								valueStr = fmt.Sprintf("\"%s\"", v)
							}
						case bool:
							valueStr = fmt.Sprintf("%v", v)
						case float64:
							if v == float64(int(v)) {
								valueStr = fmt.Sprintf("%d", int(v))
							} else {
								valueStr = fmt.Sprintf("%v", v)
							}
						default:
							valueStr = fmt.Sprintf("%v", v)
						}
						params = append(params, fmt.Sprintf("%s: %s", key, valueStr))
					}
				}

				if len(params) > 0 {
					toolMsg += " | " + strings.Join(params, ", ")
				}
			}

			m.messages = append(m.messages, ChatMessage{
				Role:      "tool",
				Content:   toolMsg,
				Timestamp: time.Now(),
			})
			m.updateViewport()

		case MsgStreamChunk:
			var chunkData struct {
				Content string `json:"content"`
				Status  string `json:"status"`
			}
			json.Unmarshal(msg.message.Data, &chunkData)

			if chunkData.Status == "thinking" {
				// Show thinking indicator
				m.messages = append(m.messages, ChatMessage{
					Role:      "system",
					Content:   "💭 Thinking...",
					Timestamp: time.Now(),
				})
			} else if chunkData.Content != "" {
				// Update or append assistant message
				if len(m.messages) > 0 && m.messages[len(m.messages)-1].Role == "assistant" {
					m.messages[len(m.messages)-1].Content = chunkData.Content
				} else {
					m.messages = append(m.messages, ChatMessage{
						Role:      "assistant",
						Content:   chunkData.Content,
						Timestamp: time.Now(),
					})
				}
			}
			m.updateViewport()

		case MsgRunCypher:
			// Handle run_cypher request from agent
			var queryData struct {
				Query     string `json:"query"`
				RequestID string `json:"request_id"`
			}
			json.Unmarshal(msg.message.Data, &queryData)

			if m.graphResult != nil && m.graphResult.Database != nil {
				cmds = append(cmds, m.executeCypher(queryData.Query, queryData.RequestID))
			} else {
				// Send error response if graph is not ready
				response := AgentMessage{
					Type: MsgCypherResult,
					Data: json.RawMessage(fmt.Sprintf(`{"request_id":"%s","error":"Graph database not initialized"}`, queryData.RequestID)),
				}
				msgBytes, _ := json.Marshal(response)
				if m.agentStdin != nil {
					m.agentStdin.Write(msgBytes)
					m.agentStdin.Write([]byte("\n"))
				}
			}

		case MsgError:
			var errData struct {
				Message string `json:"message"`
				Details string `json:"details"`
			}
			json.Unmarshal(msg.message.Data, &errData)

			m.messages = append(m.messages, ChatMessage{
				Role:      "system",
				Content:   fmt.Sprintf("❌ Error: %s", errData.Message),
				Timestamp: time.Now(),
				IsError:   true,
			})
			m.isProcessing = false
			m.updateViewport()
		}

	case graphBuiltMsg:
		if msg.err != nil {
			m.messages = append(m.messages, ChatMessage{
				Role:      "system",
				Content:   fmt.Sprintf("⚠️ Graph database initialization failed: %s", msg.err.Error()),
				Timestamp: time.Now(),
				IsError:   true,
			})
		} else {
			m.graphResult = msg.result
			m.messages = append(m.messages, ChatMessage{
				Role: "system",
				Content: fmt.Sprintf("✓ Graph database initialized with %d files, %d functions, %d classes",
					msg.result.Stats.FilesCount, msg.result.Stats.FunctionsCount, msg.result.Stats.ClassesCount),
				Timestamp: time.Now(),
			})
		}
		m.updateViewport()

	case cypherResultMsg:
		// Send the Cypher result back to the agent
		var responseData map[string]interface{}
		if msg.err != nil {
			responseData = map[string]interface{}{
				"request_id": msg.requestID,
				"error":      msg.err.Error(),
			}
		} else {
			responseData = map[string]interface{}{
				"request_id": msg.requestID,
				"result":     msg.result,
			}
		}

		dataBytes, err := json.Marshal(responseData)
		if err != nil {
			log.Printf("Failed to marshal cypher result: %v", err)
			return m, nil
		}

		response := AgentMessage{
			Type: MsgCypherResult,
			Data: json.RawMessage(dataBytes),
		}

		msgBytes, err := json.Marshal(response)
		if err != nil {
			log.Printf("Failed to marshal response message: %v", err)
			return m, nil
		}

		if m.agentStdin != nil {
			m.agentStdin.Write(msgBytes)
			m.agentStdin.Write([]byte("\n"))
			log.Printf("Sent cypher_result to agent: %s", string(msgBytes))
		}

	case errMsg:
		m.err = msg.err
		m.isProcessing = false
		m.messages = append(m.messages, ChatMessage{
			Role:      "system",
			Content:   fmt.Sprintf("❌ System Error: %s", msg.err.Error()),
			Timestamp: time.Now(),
			IsError:   true,
		})
		m.updateViewport()
	}

	// Update sub-components
	if m.state == StateAPIKey {
		var cmd tea.Cmd
		m.apiKeyInput, cmd = m.apiKeyInput.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.state == StateChat {
		var cmd tea.Cmd
		m.chatInput, cmd = m.chatInput.Update(msg)
		cmds = append(cmds, cmd)

		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) updateViewport() {
	var content strings.Builder

	for _, msg := range m.messages {
		timestamp := msg.Timestamp.Format("15:04:05")

		var style lipgloss.Style
		var prefix string

		switch msg.Role {
		case "user":
			style = userMsgStyle
			prefix = "You"
		case "assistant":
			style = assistantMsgStyle
			prefix = "AI"
		case "tool":
			style = toolMsgStyle
			prefix = "Tool"
		case "system":
			if msg.IsError {
				style = errorStyle
			} else {
				style = statusStyle
			}
			prefix = "System"
		}

		// Format based on role
		if msg.Role == "tool" {
			// Tool messages get a compact single-line format
			content.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, toolMsgStyle.Render(msg.Content)))
		} else {
			// Regular messages with prefix and indentation
			content.WriteString(fmt.Sprintf("[%s] %s:\n", timestamp, style.Render(prefix)))

			// Wrap and indent message content
			lines := strings.Split(msg.Content, "\n")
			for _, line := range lines {
				content.WriteString(fmt.Sprintf("  %s\n", line))
			}
		}
		content.WriteString("\n")
	}

	m.viewport.SetContent(content.String())
	m.viewport.GotoBottom()
}

func (m Model) buildGraph() tea.Cmd {
	return func() tea.Msg {
		// Build the graph database
		dbPath := filepath.Join(m.workDir, ".onyx-graphdb")

		// Redirect stderr to discard KuzuDB's verbose parser errors
		// Save the original stderr
		origStderr := os.Stderr
		// Create a null writer to discard output
		devNull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		os.Stderr = devNull

		// Build the graph with stderr redirected
		result, err := graph.BuildGraph(graph.BuildGraphOptions{
			RepoPath:    m.workDir,
			DBPath:      dbPath,
			CleanupDB:   false, // Keep the database for reuse
			LoadEnvFile: false, // Don't load .env file
		})

		// Restore original stderr
		os.Stderr = origStderr
		devNull.Close()

		if err != nil {
			return graphBuiltMsg{err: err}
		}

		return graphBuiltMsg{result: result}
	}
}

func (m Model) executeCypher(query string, requestID string) tea.Cmd {
	return func() tea.Msg {
		// Log the query for debugging
		log.Printf("Cypher query: %s", query)

		if m.graphResult == nil || m.graphResult.Database == nil {
			return cypherResultMsg{
				requestID: requestID,
				err:       fmt.Errorf("Graph database not initialized"),
			}
		}

		// Execute the Cypher query
		result, err := m.graphResult.Database.ExecuteQuery(query)

		// Log the result for debugging
		if err != nil {
			log.Printf("Cypher error: %v", err)
		} else {
			log.Printf("Cypher result: %s", result)
		}

		return cypherResultMsg{
			requestID: requestID,
			result:    result,
			err:       err,
		}
	}
}

func (m Model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	var content string

	switch m.state {
	case StateAPIKey:
		title := titleStyle.Render("🚀 Onyx AI Coding TUI")

		prompt := "OpenAI API key:"
		if m.isProcessing {
			prompt = "Initializing agent..."
		}

		input := inputStyle.Render(m.apiKeyInput.View())

		help := helpStyle.Render("\nPress Enter to continue • Ctrl+C to quit")

		content = lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			prompt,
			input,
			help,
		)

		if m.err != nil {
			content += "\n\n" + errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
		}

	case StateChat:
		title := titleStyle.Render("💬 Onyx AI Assistant")

		status := statusStyle.Render(fmt.Sprintf("Connected • %d messages", len(m.messages)))
		if m.isProcessing {
			status = statusStyle.Render("Processing... 🔄")
		}

		header := lipgloss.JoinHorizontal(
			lipgloss.Left,
			title,
			strings.Repeat(" ", 10),
			status,
		)

		chatHistory := m.viewport.View()

		inputLabel := "Message:"
		if m.isProcessing {
			inputLabel = "Waiting for response..."
		}

		chatInputView := lipgloss.JoinVertical(
			lipgloss.Left,
			inputLabel,
			inputStyle.Render(m.chatInput.View()),
		)

		help := helpStyle.Render("Ctrl+S to send • Ctrl+C to quit")

		content = lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			chatHistory,
			chatInputView,
			help,
		)
	}

	return appStyle.Render(content)
}

func main() {
	// Set up logging - try to create log file but don't fail if we can't
	homeDir, _ := os.UserHomeDir()
	logPath := filepath.Join(homeDir, ".local", "lib", "onyx-tui", "onyx-tui.log")
	logFile, err := os.Create(logPath)
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	} else {
		// If we can't create the log file, just log to stderr
		log.SetOutput(os.Stderr)
	}

	// Create and run the TUI
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
