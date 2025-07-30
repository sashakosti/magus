package tui

import (
	"fmt"
	"magus/player"
	"magus/storage"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type summaryFocusable int

const (
	focusDistInput summaryFocusable = iota
	focusReflectionArea
	focusSummaryButton
)

type dungeonSummaryModel struct {
	result         DungeonResult
	distInput      textinput.Model
	reflectionArea textarea.Model
	focused        summaryFocusable
}

func NewDungeonSummaryState(m *Model, result DungeonResult) State {
	distInput := textinput.New()
	distInput.Placeholder = "0"
	distInput.Focus()
	distInput.CharLimit = 3
	distInput.Width = 5
	distInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	
	reflectionArea := textarea.New()
	reflectionArea.Placeholder = "Что было сделано? Какие возникли трудности?"
	reflectionArea.SetHeight(5) // Высота фиксирована

	return &dungeonSummaryModel{
		result:         result,
		distInput:      distInput,
		reflectionArea: reflectionArea,
		focused:        focusDistInput,
	}
}

func (s *dungeonSummaryModel) Init() tea.Cmd {
	return textinput.Blink
}

func (s *dungeonSummaryModel) Update(m *Model, msg tea.Msg) (State, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))):
			return NewHomepageState(m), nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			s.focused = (s.focused + 1) % 3
			s.updateFocus()
			return s, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab"))):
			s.focused = (s.focused - 1 + 3) % 3
			s.updateFocus()
			return s, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if s.focused == focusSummaryButton {
				return s.finalizeSession(m)
			}
		}
	}

	if s.focused == focusDistInput {
		s.distInput, cmd = s.distInput.Update(msg)
		cmds = append(cmds, cmd)
	} else if s.focused == focusReflectionArea {
		s.reflectionArea, cmd = s.reflectionArea.Update(msg)
		cmds = append(cmds, cmd)
	}

	return s, tea.Batch(cmds...)
}

func (s *dungeonSummaryModel) updateFocus() {
	if s.focused == focusDistInput {
		s.distInput.Focus()
		s.reflectionArea.Blur()
	} else if s.focused == focusReflectionArea {
		s.distInput.Blur()
		s.reflectionArea.Focus()
	} else {
		s.distInput.Blur()
		s.reflectionArea.Blur()
	}
}

func (s *dungeonSummaryModel) finalizeSession(m *Model) (State, tea.Cmd) {
	// 1. Рассчитать урон и XP
	realDistractions, _ := strconv.Atoi(s.distInput.Value())
	hpLoss := (realDistractions - s.result.DistractionAttacks) * 5
	if hpLoss < 0 {
		hpLoss = 0
	}

	xpGained := int(s.result.Duration.Minutes()) * 2
	if s.result.Success && realDistractions <= s.result.DistractionAttacks {
		xpGained += 25 // Бонус за хорошую концентрацию
	}

	// 2. Обновить данные игрока
	p, _ := player.LoadPlayer()
	p.HP -= hpLoss
	player.SavePlayer(p)
	m.Player = p

	canLevelUp, _ := player.AddXP(xpGained)
	if canLevelUp {
		// TODO: Handle level up
	}

	// 3. Сохранить рефлексию
	reflection := s.reflectionArea.Value()
	if reflection != "" {
		note := storage.ReflectionNote{
			Date:     time.Now(),
			Duration: s.result.Duration,
			Content:  reflection,
			XPEarned: xpGained,
			HPLoss:   hpLoss,
		}
		storage.SaveReflection(note)
	}

	// 4. Вернуться на главный экран
	return NewHomepageState(m), nil
}

func (s *dungeonSummaryModel) View(m *Model) string {
	// Set width before rendering
	s.reflectionArea.SetWidth(m.TerminalWidth - m.styles.QuestCardStyle.GetHorizontalFrameSize()*2 - 4)

	title := m.styles.TitleStyle.Render("Отчет о сессии")

	stats := fmt.Sprintf("Длительность: %s\nАтаки на концентрацию: %d",
		formatDuration(s.result.Duration), s.result.DistractionAttacks)

	distractionPrompt := "Сколько раз вы отвлеклись на самом деле?"

	button := "[ Завершить ]"
	if s.focused == focusSummaryButton {
		button = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("> Завершить <")
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		stats,
		"\n",
		distractionPrompt,
		s.distInput.View(),
		"\nЗаметки о сессии:",
		s.reflectionArea.View(),
		"\n",
		button,
	)

	return lipgloss.Place(
		m.TerminalWidth, m.TerminalHeight,
		lipgloss.Center, lipgloss.Center,
		m.styles.QuestCardStyle.Render(content),
	)
}
