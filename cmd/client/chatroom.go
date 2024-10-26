package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/oreoluwa-bs/sweetnuttings/cmd/client/constants"
)

type Chatroom struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Messages []ChatMessage `json:"messages"`
}

func (r Chatroom) FilterValue() string {
	return r.Name
}

func (r Chatroom) Title() string {
	return r.Name
}

func (r Chatroom) Description() string {
	return fmt.Sprintf("id: %s", r.ID)
}

type ChatMessage struct {
	MemberID string `json:"member_id"`
	RoomID   string `json:"room_id"`
	Message  string `json:"message"`
}

func getChatRooms() ([]Chatroom, error) {
	c := &http.Client{Timeout: 10 * time.Second}
	var rooms []Chatroom

	resp, err := c.Get("http://localhost:3001/rooms")
	if err != nil {
		return rooms, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return rooms, err
	}

	err = json.Unmarshal(body, &rooms)
	if err != nil {
		return rooms, err
	}

	return rooms, nil
}

type chatroomModel struct {
	rooms            list.Model
	activeChatWindow tea.Model
	choosingRoom     bool
}

func NewChatroomModel() tea.Model {

	rs, err := getChatRooms()
	if err != nil {
		log.Panicf("could not get rooms %v", err)
	}

	m := chatroomModel{
		rooms:        list.New(chatroomsToListItem(rs), list.NewDefaultDelegate(), 0, 0),
		choosingRoom: true,
	}
	m.rooms.Title = "Chat rooms"

	return m
}

func (m chatroomModel) Init() tea.Cmd {
	return nil
}

func (m chatroomModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.choosingRoom {
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.rooms.SetSize(msg.Width, msg.Height)
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, constants.Keymap.Quit):
				return m, tea.Quit
			case key.Matches(msg, constants.Keymap.Enter):
				// Enter chat window for selected room
				activeRoom := m.rooms.SelectedItem().(Chatroom)
				m.activeChatWindow = NewChatWindow(activeRoom.ID)
				m.choosingRoom = false
				return m, nil
			default:
				m.rooms, cmd = m.rooms.Update(msg)
				return m, cmd
			}
		}
	} else if m.activeChatWindow != nil {
		var newWindow tea.Model
		newWindow, cmd = m.activeChatWindow.Update(msg)

		// Debug log to track key handling
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			// fmt.Printf("Key pressed: %s\n", keyMsg.String())
			if key.Matches(keyMsg, constants.Keymap.Back) {
				// fmt.Println("Back key pressed")
				m.choosingRoom = true
				m.activeChatWindow = nil
				return m, nil
			}
		}

		m.activeChatWindow = newWindow
	}

	return m, cmd
}

func (m chatroomModel) View() string {

	if m.choosingRoom {
		return m.rooms.View()
	}
	if m.activeChatWindow != nil {
		return m.activeChatWindow.View()
	}
	return ""
}

func chatroomsToListItem(ch []Chatroom) []list.Item {
	var items = make([]list.Item, len(ch))

	for i, cr := range ch {
		items[i] = list.Item(cr)
	}

	return items
}

func (m chatroomModel) getActiveRoomID() string {
	items := m.rooms.Items()
	activeItem := items[m.rooms.Index()]
	return activeItem.(Chatroom).ID
}
