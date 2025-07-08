package tui

import (
	"fmt"

	"magus/player"
	"magus/rpg"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) updateLevelUp(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.perkChoices)-1 {
				m.cursor++
			}
		case "enter":
			chosenPerk := m.perkChoices[m.cursor]
			player.LevelUpPlayer(chosenPerk.Name)
			p, _ := player.LoadPlayer()
			m.player = *p

			if m.player.Level >= 3 && m.player.Class == player.ClassNone {
				m.state = stateClassChoice
				m.classChoices = rpg.GetAvailableClasses()
				m.cursor = 0
			} else {
				m.state = stateHomepage
				m.statusMessage = fmt.Sprintf("Вы выучили перк: %s! И получили 10 очков навыков.", chosenPerk.Name)
			}
			return m, nil
		}
	}
	return m, nil
}

func (m *Model) viewLevelUp() string {
	s := "🔥 Поздравляем! Новый уровень!\n\n"
	s += "Выберите новый перк:\n\n"
	for i, perk := range m.perkChoices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s: %s\n", cursor, perk.Name, perk.Description)
	}
	s += "\nНажмите 'enter' для выбора.\n"
	return lipgloss.NewStyle().Border(lipgloss.DoubleBorder(), true).Padding(2).Render(s)
}
