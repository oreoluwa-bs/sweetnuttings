package main

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/oreoluwa-bs/sweetnuttings/cmd/tui/constants"
)

type model struct {
	chatroom tea.Model
}

func NewModel() *model {
	return &model{
		chatroom: NewChatroomModel(),
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

	newR, newCmd := m.chatroom.Update(msg)
	m.chatroom = newR
	cmd = newCmd
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {

	return m.chatroom.View()
}
