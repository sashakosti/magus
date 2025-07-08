package tui

import (
	"fmt"

	"magus/player"
	"magus/rpg"

	"github.com/charmbracelet/bubbletea"
)

func (m *Model) updateSkills(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.skills)-1 {
				m.cursor++
			}
		case "enter":
			if m.player.SkillPoints > 0 {
				skillToIncrease := m.skills[m.cursor]
				err := rpg.IncreaseSkill(&m.player, skillToIncrease.Name)
				if err != nil {
					m.statusMessage = fmt.Sprintf("Ошибка: %v", err)
				} else {
					p, _ := player.LoadPlayer()
					m.player = *p
					m.statusMessage = fmt.Sprintf("Навык '%s' увеличен!", skillToIncrease.Name)
				}
			} else {
				m.statusMessage = "Недостаточно очков навыков."
			}
		}
	}
	return m, nil
}

func (m *Model) viewSkills() string {
	s := fmt.Sprintf("🧠 Навыки (Очки: %d)\n\n", m.player.SkillPoints)
	for i, skill := range m.skills {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		level := m.player.Skills[skill.Name]
		s += fmt.Sprintf("%s %s: %d\n  %s\n\n", cursor, skill.Name, level, skill.Description)
	}
	s += fmt.Sprintf("\n%s\n", m.statusMessage)
	return s
}
