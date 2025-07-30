package tui

import (
	"fmt"
	"magus/player"
	"magus/rpg"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LevelUpState struct {
	skillChoices []player.SkillNode
	cursor       int
}

func NewLevelUpState(m *Model) (*LevelUpState, error) {
	s := &LevelUpState{}
	skillTrees, err := rpg.LoadSkillTrees(m.Player)
	if err != nil {
		return nil, fmt.Errorf("ошибка загрузки дерева навыков: %w", err)
	}

	allSkills := make(map[string]player.SkillNode)
	for id, node := range skillTrees.Common {
		allSkills[id] = node
	}
	for id, node := range skillTrees.Class {
		allSkills[id] = node
	}

	for _, node := range allSkills {
		if rpg.IsSkillAvailable(m.Player, node, allSkills) {
			s.skillChoices = append(s.skillChoices, node)
		}
	}

	if len(s.skillChoices) == 0 {
		// Если нет доступных навыков, просто повышаем уровень и выходим
		player.LevelUpPlayer("")
		player.SavePlayer(m.Player)
		return nil, fmt.Errorf("нет доступных навыков для изучения")
	}

	return s, nil
}

func (s *LevelUpState) Init() tea.Cmd {
	return nil
}

func (s *LevelUpState) Update(m *Model, msg tea.Msg) (State, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.cursor < len(s.skillChoices)-1 {
				s.cursor++
			}
		case "enter":
			chosenSkill := s.skillChoices[s.cursor]
			player.LevelUpPlayer(chosenSkill.ID)
			// No need to save here, LevelUpPlayer does it.
			// We need to reload the player to get the updated state.
			reloadedPlayer, err := player.LoadPlayer()
			if err == nil {
				m.Player = reloadedPlayer
			}
			return NewHomepageState(m), nil
		}
	}
	return s, nil
}

func (s *LevelUpState) View(m *Model) string {
	var b strings.Builder
	b.WriteString("🔥 Поздравляем! Новый уровень!\n\n")
	b.WriteString("Выберите навык для изучения:\n\n")
	for i, skill := range s.skillChoices {
		cursor := " "
		if s.cursor == i {
			cursor = ">"
		}
		b.WriteString(fmt.Sprintf("%s %s: %s\n", cursor, skill.Name, skill.Description))
	}
	b.WriteString("\nНажмите 'enter' для выбора.\n")
	return lipgloss.NewStyle().Border(lipgloss.DoubleBorder(), true).Padding(2).Render(b.String())
}
