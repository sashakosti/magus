package tui

import (
	"fmt"

	"magus/player"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func newCreatePlayerInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Имя твоего героя"
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 50
	return ti
}

func (m *Model) updateCreatePlayer(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
		if m.createPlayerInput.Value() == "" {
			return m, nil
		}
		if _, err := player.CreatePlayer(m.createPlayerInput.Value()); err != nil {
			m.statusMessage = fmt.Sprintf("Ошибка: %v", err)
			return m, nil
		}
		newModel := InitialModel()
		return &newModel, nil
	}
	m.createPlayerInput, cmd = m.createPlayerInput.Update(msg)
	return m, cmd
}

func (m *Model) viewCreatePlayer() string {
	content := fmt.Sprintf(
		"Добро пожаловать в Magus!\n\nДавай создадим твоего персонажа.\n\n%s\n\nНажми Enter, чтобы начать.",
		m.createPlayerInput.View(),
	)
	return lipgloss.Place(m.terminalWidth, m.terminalHeight, lipgloss.Center, lipgloss.Center, content)
}
