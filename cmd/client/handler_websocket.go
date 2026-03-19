package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"
)

// Styles ────────────────────────────────────────────────────────────────────

var (
	messageStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	inputStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	borderStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	connectStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Italic(true)
	disconnectStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Italic(true)
)

// Messages (events) ─────────────────────────────────────────────────────────

type msgReceived struct{ content string }
type wsError struct{ err error }

// Model ─────────────────────────────────────────────────────────────────────

type model struct {
	conn     *websocket.Conn
	messages []string
	input    string
	width    int
	height   int
}

func initialModel(conn *websocket.Conn) model {
	return model{conn: conn}
}

// waitForMessage écoute le websocket et envoie les messages au programme bubbletea
func waitForMessage(conn *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return wsError{err}
		}
		return msgReceived{content: string(msg)}
	}
}

// Init ──────────────────────────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	return waitForMessage(m.conn)
}

// Update ────────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case msgReceived:
		m.messages = append(m.messages, msg.content)
		// Keep only the last N messages that fit on screen
		maxMessages := m.height - 4
		if maxMessages > 0 && len(m.messages) > maxMessages {
			m.messages = m.messages[len(m.messages)-maxMessages:]
		}
		return m, waitForMessage(m.conn)

	case wsError:
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			if m.input == "" {
				return m, nil
			}
			err := m.conn.WriteMessage(websocket.TextMessage, []byte(m.input))
			if err != nil {
				log.Println("write error:", err)
				return m, tea.Quit
			}
			m.input = ""

		case tea.KeyBackspace:
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}

		default:
			m.input += msg.String()
		}
	}

	return m, nil
}

// View ──────────────────────────────────────────────────────────────────────

func (m model) View() string {
	// Messages area
	messagesArea := ""
	for _, msg := range m.messages {
		if strings.HasPrefix(msg, "User ") && strings.HasSuffix(msg, "disconnected") {
			messagesArea += disconnectStyle.Render(msg) + "\n"
		} else if strings.HasPrefix(msg, "User ") && strings.HasSuffix(msg, "connected") {
			messagesArea += connectStyle.Render(msg) + "\n"
		} else {
			messagesArea += messageStyle.Render(msg) + "\n"
		}
	}

	// Separator
	separator := borderStyle.Render("─────────────────────────────────────────")

	// Input area
	inputArea := inputStyle.Render("> " + m.input)

	return fmt.Sprintf("%s\n%s\n%s", messagesArea, separator, inputArea)
}

// Entry point ───────────────────────────────────────────────────────────────

func (cfg *config) handlerWebSocket() {
	header := http.Header{}
	header.Set("Authorization", "Bearer "+cfg.token)

	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", header)
	if err != nil {
		log.Fatal("dial error:", err)
	}
	defer conn.Close()

	p := tea.NewProgram(initialModel(conn), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal("error running TUI:", err)
	}
}
