package tui

import (
	"fmt"
	"magus/player"
	"math/rand"
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	distractionTickDuration = 10 * time.Second // Атаки на концентрацию происходят реже
)

// distractionTickMsg - это сообщение, которое инициирует проверку на отвлечение.
type distractionTickMsg struct{}

// distractionTick - это команда, которая отправляет distractionTickMsg через заданный интервал.
func distractionTick() tea.Cmd {
	return tea.Tick(distractionTickDuration, func(t time.Time) tea.Msg {
		return distractionTickMsg{}
	})
}

// --- MAIN DUNGEON MODEL ---

type dungeonModel struct {
	player             *player.Player
	timer              timer.Model
	duration           time.Duration
	distractionAttacks int // Количество симулированных атак на концентрацию
	isConfirmingExit   bool
}

func NewDungeonState(m *Model, duration time.Duration) State {
	p := *m.Player // Создаем копию, чтобы не менять глобальное состояние до конца сессии

	return &dungeonModel{
		player:             &p,
		timer:              timer.New(duration),
		duration:           duration,
		distractionAttacks: 0,
		isConfirmingExit:   false,
	}
}

func (s *dungeonModel) Init() tea.Cmd {
	return tea.Batch(s.timer.Init(), distractionTick())
}

func (s *dungeonModel) Update(m *Model, msg tea.Msg) (State, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if s.isConfirmingExit {
		if msg, ok := msg.(tea.KeyMsg); ok {
			switch msg.String() {
			case "y", "Y":
				// При досрочном выходе переходим на экран отчета, но без наград.
				result := DungeonResult{
					Duration:           s.duration - s.timer.Timeout, // Фиксируем, сколько времени прошло
					DistractionAttacks: s.distractionAttacks,
					Success:            false, // Сессия не была успешной
				}
				return NewDungeonSummaryState(m, result), nil
			case "n", "N", "esc":
				s.isConfirmingExit = false
				return s, s.timer.Toggle() // Возобновляем таймер
			}
		}
		return s, nil
	}

	switch msg := msg.(type) {
	case timer.TimeoutMsg:
		result := DungeonResult{
			Duration:           s.duration,
			DistractionAttacks: s.distractionAttacks,
			Success:            true,
		}
		return NewDungeonSummaryState(m, result), nil

	case distractionTickMsg:
		// Логика "атаки-отвлечения"
		// Шанс на атаку зависит от навыка "Концентрация"
		focusSkill := s.player.Skills["Концентрация"]
		// Чем выше навык, тем меньше шанс. Примерная формула:
		chance := 50 - (focusSkill * 2)
		if chance < 5 {
			chance = 5 // Минимальный шанс 5%
		}

		if rand.Intn(100) < chance {
			s.distractionAttacks++
		}

		cmds = append(cmds, distractionTick()) // Запускаем следующий тик

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			s.isConfirmingExit = true
			return s, s.timer.Toggle() // Паузим таймер
		}
	}

	s.timer, cmd = s.timer.Update(msg)
	cmds = append(cmds, cmd)

	return s, tea.Batch(cmds...)
}

func (s *dungeonModel) View(m *Model) string {
	timerView := lipgloss.NewStyle().Bold(true).Render("Осталось времени: ", s.timer.View())
	focusMessage := "Вы в подземелье. Сконцентрируйтесь на задаче."
	attacksView := fmt.Sprintf("Атаки на концентрацию: %d", s.distractionAttacks)

	mainView := lipgloss.JoinVertical(
		lipgloss.Center,
		timerView,
		"\n",
		focusMessage,
		attacksView,
		"\n\n",
		m.styles.StatusMessageStyle.Render("Нажмите 'q' или 'esc' для досрочного завершения."),
	)

	if s.isConfirmingExit {
		dialogBox := m.styles.QuestCardStyle.Copy().BorderForeground(lipgloss.Color("202")).Render(
			lipgloss.JoinVertical(
				lipgloss.Center,
				"Вы уверены, что хотите прервать фокус-сессию?",
				"Прогресс не будет засчитан.",
				"\n(y/n)",
			),
		)
		return lipgloss.Place(m.TerminalWidth, m.TerminalHeight, lipgloss.Center, lipgloss.Center, dialogBox)
	}

	return lipgloss.Place(
		m.TerminalWidth, m.TerminalHeight,
		lipgloss.Center, lipgloss.Center,
		mainView,
	)
}


