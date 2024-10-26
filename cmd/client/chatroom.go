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

const (
	chatroomListView = iota
	chatwindowView
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
	currentView currentView

	rooms list.Model

	window tea.Model
}

func NewChatroomModel() tea.Model {

	rs, err := getChatRooms()
	if err != nil {
		log.Panicf("could not get rooms %v", err)
	}

	m := chatroomModel{
		rooms: list.New(chatroomsToListItem(rs), list.NewDefaultDelegate(), 0, 0),
	}
	m.rooms.Title = "Chat rooms"

	return m
}

func (m chatroomModel) Init() tea.Cmd {
	return nil
}

func (m chatroomModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.rooms.SetSize(msg.Width, msg.Height)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, constants.Keymap.Quit):
			return m, tea.Quit
		case key.Matches(msg, constants.Keymap.Enter):
			activeRoom := m.rooms.SelectedItem().(Chatroom)
			cw := NewChatWindow(activeRoom.ID)
			m.currentView = chatwindowView
			m.window = cw
		default:
			m.rooms, cmd = m.rooms.Update(msg)
		}
	}

	switch m.currentView {
	case chatwindowView:
		newR, newCmd := m.window.Update(msg)
		rm, ok := newR.(ChatWindow)
		if !ok {
			panic("could not assert that the model is a chat window ui model")
		}

		m.window = rm
		cmd = newCmd
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m chatroomModel) View() string {
	switch m.currentView {
	case chatwindowView:
		return m.window.View()
	default:
		return m.rooms.View()
	}
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
