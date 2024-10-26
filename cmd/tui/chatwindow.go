package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/oreoluwa-bs/sweetnuttings/cmd/tui/constants"
)

type ChatWindow struct {
	activeRoomID string
	viewport     viewport.Model
	textarea     textarea.Model
	messages     []ChatMessage
	senderStyle  lipgloss.Style
}

func NewChatWindow(activeRoomID string) tea.Model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)

	vp.SetContent(fmt.Sprintf(`Welcome to the chat room!
Type a message and press Enter to send.`))

	ta.KeyMap.InsertNewline.SetEnabled(false)

	cw := ChatWindow{
		activeRoomID: activeRoomID,
		viewport:     vp,
		textarea:     ta,
	}
	return cw
}

func (cw ChatWindow) Init() tea.Cmd {
	return nil
}

func (cw ChatWindow) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cw.viewport.Width = msg.Width
		cw.viewport.Height = msg.Height
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, constants.Keymap.Back):
			m := NewModel()
			return m.chatroom.Update(msg)
		case key.Matches(msg, constants.Keymap.Enter):
			v := cw.textarea.Value()

			if v == "" {
				// Don't send empty messages.
				return cw, nil
			}

			// Simulate sending a message. In your application you'll want to
			// also return a custom command to send the message off to
			// a server.
			cw.messages = append(cw.messages, ChatMessage{
				MemberID: "",
				RoomID:   cw.activeRoomID,
				Message:  cw.senderStyle.Render("You: ") + v,
			})

			var s []string

			for _, cm := range cw.messages {
				s = append(s, cm.Message)
			}

			cw.viewport.SetContent(strings.Join(s, "\n"))
			cw.textarea.Reset()
			cw.viewport.GotoBottom()

			return cw, nil
		}
	}

	cw.textarea, cmd = cw.textarea.Update(msg)
	return cw, cmd
}

func (cw ChatWindow) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		cw.viewport.View(),
		cw.textarea.View(),
	) + "\n\n"
}
