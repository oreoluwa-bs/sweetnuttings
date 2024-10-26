package main

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/oreoluwa-bs/sweetnuttings/cmd/client/constants"
)

type currentView int

const (
	roomListView = iota
)

type model struct {
	currentView currentView

	chatroom tea.Model
}

func NewModel() *model {
	return &model{
		currentView: roomListView,
		chatroom:    NewChatroomModel(),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, constants.Keymap.Quit):
			return m, tea.Quit
		}
	}

	switch m.currentView {
	case roomListView:
		newR, newCmd := m.chatroom.Update(msg)
		rm, ok := newR.(chatroomModel)
		if !ok {
			panic("could not assert that the model is a chatroom ui model")
		}

		m.chatroom = rm
		cmd = newCmd
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {

	switch m.currentView {
	case roomListView:
		return m.chatroom.View()
	default:
		return "\nPress q to quit.\n"

	}

}
