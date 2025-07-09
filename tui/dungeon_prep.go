package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) updateDungeonPrep(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if key, ok := msg.(tea.KeyMsg); ok {
		isCustomSelected := m.dungeonDurationChoices[m.dungeonDurationCursor] == "Custom"

		if isCustomSelected && m.dungeonCustomDurationInput.Focused() {
			if key.String() == "enter" {
				minutes, err := strconv.Atoi(m.dungeonCustomDurationInput.Value())
				if err == nil && minutes > 0 {
					m.dungeonSelectedDuration = time.Duration(minutes) * time.Minute
					return m.startDungeonRun()
				}
			}
			m.dungeonCustomDurationInput, cmd = m.dungeonCustomDurationInput.Update(msg)
			return m, cmd
		}

		switch key.String() {
		case "up", "k":
			if m.dungeonDurationCursor > 0 {
				m.dungeonDurationCursor--
			}
		case "down", "j":
			if m.dungeonDurationCursor < len(m.dungeonDurationChoices)-1 {
				m.dungeonDurationCursor++
			}
		case "enter":
			if isCustomSelected {
				m.dungeonCustomDurationInput.Focus()
				m.dungeonCustomDurationInput.Placeholder = "Введите минуты..."
				return m, nil
			}
			minutes, _ := strconv.Atoi(m.dungeonDurationChoices[m.dungeonDurationCursor])
			m.dungeonSelectedDuration = time.Duration(minutes) * time.Minute
			return m.startDungeonRun()
		}
	}

	return m, nil
}

func (m *Model) viewDungeonPrep() string {
	var b strings.Builder
	b.WriteString("⏳ Сколько времени вы хотите сфокусироваться?\n\n")

	for i, choice := range m.dungeonDurationChoices {
		cursor := " "
		if m.dungeonDurationCursor == i {
			cursor = ">"
		}
		line := fmt.Sprintf("%s %s минут", cursor, choice)
		if choice == "Custom" {
			line = fmt.Sprintf("%s %s", cursor, choice)
		}
		b.WriteString(line + "\n")
	}

	if m.dungeonDurationChoices[m.dungeonDurationCursor] == "Custom" {
		b.WriteString("\n" + m.dungeonCustomDurationInput.View())
	}

	b.WriteString("\n\nНавигация: ↑/↓, Enter для выбора, 'q' - назад.")
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).Padding(2).Render(b.String())
}

// startDungeonRun initializes the dungeon state and starts the ticker.
func (m *Model) startDungeonRun() (tea.Model, tea.Cmd) {
	m.state = stateDungeon
	m.dungeonFloor = 1
	m.dungeonRunXP = 0
	m.dungeonRunGold = 0
	m.dungeonLog = []string{fmt.Sprintf("Забег начался! Длительность: %v.", m.dungeonSelectedDuration)}
	m.dungeonState = DungeonStateExploring
	m.currentMonster = nil
	m.dungeonStartTime = time.Now()
	m.dungeonTicker = time.NewTicker(2 * time.Second) // Game tick every 2 seconds

	// Immediately start the first exploration
	m.handleExplore()

	// Return a command to listen for ticks
	return m, func() tea.Msg {
		return <-m.dungeonTicker.C
	}
}