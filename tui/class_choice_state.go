package tui

import (
	"fmt"
	"magus/player"
	"magus/rpg"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ClassChoiceState struct {
	choices []rpg.Class
	cursor  int
}

func NewClassChoiceState(m *Model) *ClassChoiceState {
	return &ClassChoiceState{
		choices: rpg.GetAvailableClasses(),
		cursor:  0,
	}
}

func (s *ClassChoiceState) Init() tea.Cmd {
	return nil
}

func (s *ClassChoiceState) Update(m *Model, msg tea.Msg) (State, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.cursor < len(s.choices)-1 {
				s.cursor++
			}
		case "enter":
			chosenClass := s.choices[s.cursor]
			m.Player.Class = chosenClass.Name
			player.SavePlayer(m.Player)
			// После выбора класса возвращаемся на главный экран
			return NewHomepageState(m), nil
		}
	}
	return s, nil
}

func (s *ClassChoiceState) View(m *Model) string {
	var b strings.Builder
	b.WriteString("⚔️ Пришло время выбрать свой путь!\n\n")
	b.WriteString("Выберите класс:\n\n")
	for i, class := range s.choices {
		cursor := " "
		if s.cursor == i {
			cursor = ">"
		}
		b.WriteString(fmt.Sprintf("%s %s: %s\n", cursor, class.Name, class.Description))
	}
	b.WriteString("\nНажмите 'enter' для выбора. Этот выбор нельзя будет изменить.\n")
	return lipgloss.NewStyle().Border(lipgloss.DoubleBorder(), true).Padding(2).Render(b.String())
}
