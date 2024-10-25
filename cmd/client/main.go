package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gorilla/websocket"
)

// WSMessage represents a message from the WebSocket
type WSMessage struct {
	Error   error
	Message []byte
}

// Custom tea.Msg types for WebSocket events
type wsMessageMsg WSMessage
type wsErrorMsg struct{ error }
type wsCloseMsg struct{}

type Message struct {
	MemberID string `json:"member_id"`
	RoomID   string `json:"room_id"`
	Message  string `json:"message"`
}

type ChatRoom struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (cr ChatRoom) FilterValue() string { return cr.Name }
func (cr ChatRoom) Title() string       { return cr.Name }
func (cr ChatRoom) Description() string {
	return fmt.Sprintf("Chat room: %s", cr.ID)
}

const (
	roomListView = iota
	chatRoomView
)

// Model represents the application state
type Model struct {
	// Common
	width  int
	height int
	err    error

	// View state
	currentView int

	// Room list
	rooms list.Model

	// Chat room
	currentRoom string
	messages    []string
	textarea    textarea.Model
	viewport    viewport.Model

	// ws
	wsConn *websocket.Conn
}

// Initialize a new model
func New() *Model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()
	ta.Prompt = "┃ "
	ta.CharLimit = 280
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(30, 5)

	return &Model{
		currentView: roomListView,
		textarea:    ta,
		viewport:    vp,
		messages:    []string{},
	}
}

func getRooms() ([]ChatRoom, error) {
	var rooms []ChatRoom
	req, err := http.NewRequest(http.MethodGet, "http://localhost:3001/rooms", nil)
	if err != nil {
		return rooms, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return rooms, err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return rooms, err
	}

	if err := json.Unmarshal(resBody, &rooms); err != nil {
		return rooms, err
	}

	return rooms, nil
}

// WebSocket listener command
func listenForMessages(conn *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				return wsCloseMsg{}
			}
			return wsErrorMsg{err}
		}
		return wsMessageMsg{Message: message}
	}
}

// Initialize WebSocket connection
func (m *Model) initWS() tea.Cmd {
	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:3001/ws", nil)
	if err != nil {
		return func() tea.Msg {
			return wsErrorMsg{err}
		}
	}
	m.wsConn = c
	return listenForMessages(c)
}

// Send message through WebSocket
func (m *Model) sendWSMessage(message string) tea.Cmd {
	return func() tea.Msg {
		if m.wsConn == nil {
			return wsErrorMsg{fmt.Errorf("websocket connection not established")}
		}

		msg := Message{
			RoomID:  m.currentRoom,
			Message: message,
		}

		jsonMsg, err := json.Marshal(msg)
		if err != nil {
			return wsErrorMsg{err}
		}

		if err := m.wsConn.WriteMessage(websocket.TextMessage, jsonMsg); err != nil {
			return wsErrorMsg{err}
		}
		return nil
	}
}

func (m *Model) initRooms() error {
	delegate := list.NewDefaultDelegate()
	m.rooms = list.New([]list.Item{}, delegate, m.width, m.height)
	m.rooms.Title = "Chat Rooms"

	rooms, err := getRooms()
	if err != nil {
		return err
	}
	items := make([]list.Item, len(rooms))
	for i, room := range rooms {
		items[i] = room
	}
	m.rooms.SetItems(items)
	return nil
}

func (m Model) Init() tea.Cmd {
	m.initWS()
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.initRooms()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			if m.currentView == chatRoomView {
				m.currentView = roomListView
				m.currentRoom = ""
				if err := m.initRooms(); err != nil {
					m.err = err
				}
				return m, nil
			}
		case " ":
			if m.currentView == roomListView {
				selected := m.rooms.SelectedItem().(ChatRoom)
				m.currentRoom = selected.ID
				m.currentView = chatRoomView
				m.viewport.SetContent(fmt.Sprintf("Welcome to %s!\n", selected.Name))
			}
		case "enter":
			if m.currentView == roomListView {
				selected := m.rooms.SelectedItem().(ChatRoom)
				m.currentRoom = selected.ID
				m.currentView = chatRoomView
				m.viewport.SetContent(fmt.Sprintf("Welcome to %s!\n", selected.Name))
			} else if m.currentView == chatRoomView {
				// Send message
				msg := m.textarea.Value()
				if msg != "" {
					m.messages = append(m.messages, fmt.Sprintf("You: %s", msg))
					m.viewport.SetContent(strings.Join(m.messages, "\n"))
					m.textarea.Reset()
					m.viewport.GotoBottom()
				}
				return m, m.sendWSMessage(msg)
			}
		}
	}

	var cmd tea.Cmd
	switch m.currentView {
	case roomListView:
		m.rooms, cmd = m.rooms.Update(msg)
	case chatRoomView:
		m.textarea, cmd = m.textarea.Update(msg)
	}

	return m, cmd
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v", m.err)
	}

	switch m.currentView {
	case roomListView:
		return m.rooms.View()
	case chatRoomView:
		return fmt.Sprintf(
			"%s\n\n%s",
			m.viewport.View(),
			m.textarea.View(),
		)
	default:
		return "Loading..."
	}
}

func main() {
	p := tea.NewProgram(
		New(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

}
