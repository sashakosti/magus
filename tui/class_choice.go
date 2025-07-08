package tui

import (
	"fmt"

	"magus/player"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) updateClassChoice(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.classChoices)-1 {
				m.cursor++
			}
		case "enter":
			chosenClass := m.classChoices[m.cursor]
			m.player.Class = chosenClass.Name
			player.SavePlayer(&m.player)

			m.state = stateHomepage
			m.statusMessage = fmt.Sprintf("Вы выбрали класс: %s!", chosenClass.Name)
		}
	}
	return m, nil
}

func (m *Model) viewClassChoice() string {
	s := "⚔️ Пришло время выбрать свой путь!\n\n"
	s += "Выберите класс:\n\n"
	for i, class := range m.classChoices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s: %s\n", cursor, class.Name, class.Description)
	}
	s += "\nНажмите 'enter' для выбора. Этот выбор нельзя будет изменить.\n"
	return lipgloss.NewStyle().Border(lipgloss.DoubleBorder(), true).Padding(2).Render(s)
}
