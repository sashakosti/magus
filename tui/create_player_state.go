package tui

import (
	"fmt"
	"magus/player"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CreatePlayerState struct {
	input textinput.Model
}

func NewCreatePlayerState() *CreatePlayerState {
	ti := textinput.New()
	ti.Placeholder = "Имя твоего героя"
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 50
	return &CreatePlayerState{input: ti}
}

func (s *CreatePlayerState) Init() tea.Cmd {
	return textinput.Blink
}

func (s *CreatePlayerState) Update(m *Model, msg tea.Msg) (State, tea.Cmd) {
	var cmd tea.Cmd
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
		name := s.input.Value()
		if name == "" {
			return s, nil
		}
		p, err := player.CreatePlayer(name)
		if err != nil {
			// TODO: Handle error display
			return s, nil
		}
		m.Player = p // Обновляем глобального игрока
		return NewHomepageState(m), nil
	}
	s.input, cmd = s.input.Update(msg)
	return s, cmd
}

func (s *CreatePlayerState) View(m *Model) string {
	content := fmt.Sprintf(
		"Добро пожаловать в Magus!\n\nДавай создадим твоего персонажа.\n\n%s\n\nНажми Enter, чтобы начать.",
		s.input.View(),
	)
	// The main TUI now handles placing content in the center, so we just return the content.
	return lipgloss.NewStyle().
		Width(m.TerminalWidth).
		Height(m.TerminalHeight).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}
