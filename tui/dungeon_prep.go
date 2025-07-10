package tui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Regex to validate HHMMSS format
	timeInputRegex = regexp.MustCompile(`^\d{0,6}$`)

	// Styles for the timer input
	timerStyle      = lipgloss.NewStyle().Width(10).Align(lipgloss.Center)
	focusedTimerStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(lipgloss.Color("205")).
				Padding(0, 1)
	unfocusedTimerStyle = lipgloss.NewStyle().
				Border(lipgloss.HiddenBorder(), true).
				Padding(0, 1)
	placeholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func (m *Model) updateDungeonPrep(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter":
			duration, err := parseDungeonDuration(m.dungeonCustomDurationInput.Value())
			if err == nil && duration > 0 {
				m.dungeonSelectedDuration = duration
				return m.startDungeonRun()
			}
			m.statusMessage = "Неверный формат времени. Введите до 6 цифр."

		case "backspace":
			if len(m.dungeonCustomDurationInput.Value()) > 0 {
				m.dungeonCustomDurationInput.SetValue(m.dungeonCustomDurationInput.Value()[:len(m.dungeonCustomDurationInput.Value())-1])
			}
		default:
			newVal := m.dungeonCustomDurationInput.Value() + key.String()
			if timeInputRegex.MatchString(newVal) {
				m.dungeonCustomDurationInput.SetValue(newVal)
			}
		}
	}

	m.dungeonCustomDurationInput, cmd = m.dungeonCustomDurationInput.Update(msg)
	return m, cmd
}

func (m *Model) viewDungeonPrep() string {
	var b strings.Builder

	title := "⏳ Сколько времени вы хотите сфокусироваться?"
	b.WriteString(lipgloss.PlaceHorizontal(m.terminalWidth, lipgloss.Center, title))
	b.WriteString("\n\n")

	inputVal := m.dungeonCustomDurationInput.Value()
	placeholder := "000000"
	displayVal := ""

	paddedVal := fmt.Sprintf("%-6s", inputVal)
	for i := 0; i < 6; i++ {
		if i > 0 && i%2 == 0 {
			displayVal += ":"
		}
		if i < len(inputVal) {
			displayVal += string(paddedVal[i])
		} else {
			displayVal += placeholderStyle.Render(string(placeholder[i]))
		}
	}

	timerView := timerStyle.Render(displayVal)
	containerStyle := focusedTimerStyle
	if !m.dungeonCustomDurationInput.Focused() {
		containerStyle = unfocusedTimerStyle
	}

	b.WriteString(lipgloss.PlaceHorizontal(m.terminalWidth, lipgloss.Center, containerStyle.Render(timerView)))
	b.WriteString("\n\n")

	help := "(Введите 6 цифр для ЧЧ:ММ:СС и нажмите Enter)"
	b.WriteString(lipgloss.PlaceHorizontal(m.terminalWidth, lipgloss.Center, placeholderStyle.Render(help)))

	return docStyle.Render(b.String())
}

func (m *Model) startDungeonRun() (tea.Model, tea.Cmd) {
	m.state = stateDungeon
	m.dungeonFloor = 1
	m.dungeonRunXP = 0
	m.dungeonRunGold = 0
	m.dungeonLog = []string{fmt.Sprintf("Забег начался! Длительность: %s.", formatDuration(m.dungeonSelectedDuration))}
	m.dungeonState = DungeonStateExploring
	m.currentMonster = nil
	m.dungeonStartTime = time.Now()
	m.dungeonTicker = time.NewTicker(2 * time.Second)

	return m, func() tea.Msg {
		return dungeonTickMsg(<-m.dungeonTicker.C)
	}
}

func parseDungeonDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("пустая строка")
	}

	// Pad with leading zeros to 6 digits
	padded := fmt.Sprintf("%06s", s)

	h, _ := strconv.Atoi(padded[0:2])
	m, _ := strconv.Atoi(padded[2:4])
	sec, _ := strconv.Atoi(padded[4:6])

	return time.Hour*time.Duration(h) +
		time.Minute*time.Duration(m) +
		time.Second*time.Duration(sec), nil
}

