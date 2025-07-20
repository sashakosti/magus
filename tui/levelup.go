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
			if m.cursor < len(m.skillChoices)-1 {
				m.cursor++
			}
		case "enter":
			chosenSkill := m.skillChoices[m.cursor]
			// Мы не вызываем LevelUpPlayer здесь, так как очки уже добавлены.
			// Вместо этого, мы разблокируем выбранный навык.
			err := m.unlockSkill(chosenSkill.ID)
			if err != nil {
				m.statusMessage = fmt.Sprintf("❗ Ошибка изучения навыка: %v", err)
			} else {
				m.statusMessage = fmt.Sprintf("✨ Вы изучили навык: %s!", chosenSkill.Name)
			}

			// Обновляем игрока после возможного изучения навыка
			p, _ := player.LoadPlayer()
			m.player = *p

			// Проверяем, нужно ли выбирать класс
			if m.player.Level >= 3 && m.player.Class == player.ClassNone {
				m.state = stateClassChoice
				m.classChoices = rpg.GetAvailableClasses()
				m.cursor = 0
			} else {
				m.state = stateHomepage
			}
			return m, nil
		}
	}
	return m, nil
}

func (m *Model) viewLevelUp() string {
	s := "🔥 Поздравляем! Новый уровень!\n\n"
	s += "Выберите навык для изучения:\n\n"
	for i, skill := range m.skillChoices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s: %s\n", cursor, skill.Name, skill.Description)
	}
	s += "\nНажмите 'enter' для выбора.\n"
	return lipgloss.NewStyle().Border(lipgloss.DoubleBorder(), true).Padding(2).Render(s)
}
