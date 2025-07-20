package tui

import (
	"fmt"
	"strings"

	"magus/player"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func (m *Model) updateHomepage(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "up", "k":
			if m.homepageCursor > 0 {
				m.homepageCursor--
			}
		case "down", "j":
			if m.homepageCursor < 3 { // Now 4 items: 0, 1, 2, 3
				m.homepageCursor++
			}
		case "enter":
			switch m.homepageCursor {
			case 0:
				m.pushState(stateQuests)
			case 1:
				m.pushState(statePerks)
			case 2:
				m.pushState(stateDungeonPrep)
			case 3:
				return m, tea.Quit
			}
			m.cursor = 0
			m.statusMessage = ""
		}
	}
	return m, nil
}

func (m *Model) viewHomepage() string {
	lines := []struct {
		icon string
		text string
	}{
		{"🧙", fmt.Sprintf("Игрок: %s (Уровень: %d)", m.player.Name, m.player.Level)},
		{"🛡", fmt.Sprintf("Класс: %s", m.player.Class)},
		{"❤", fmt.Sprintf("HP: %d / %d", m.player.HP, m.player.MaxHP)},
		{"💧", fmt.Sprintf("Мана: %d / %d", m.player.Mana, m.player.MaxMana)},
		{"💰", fmt.Sprintf("Золото: %d", m.player.Gold)},
		{"🎁", fmt.Sprintf("Навыки: %d", len(m.player.UnlockedSkills))},
		{"✨", fmt.Sprintf("Очки навыков: %d", m.player.SkillPoints)},
	}

	var iconLines []string
	var textLines []string

	maxIconWidth := 0
	for _, line := range lines {
		if w := runewidth.StringWidth(line.icon); w > maxIconWidth {
			maxIconWidth = w
		}
	}

	for _, line := range lines {
		if (line.icon == "🛡" && m.player.Class == player.ClassNone) ||
			(line.icon == "🎁" && len(m.player.UnlockedSkills) == 0) ||
			(line.icon == "✨" && m.player.SkillPoints == 0) {
			continue
		}
		padding := strings.Repeat(" ", maxIconWidth-runewidth.StringWidth(line.icon))
		iconLines = append(iconLines, line.icon+padding)
		textLines = append(textLines, line.text)
	}

	iconsBlock := lipgloss.JoinVertical(lipgloss.Left, iconLines...)
	textsBlock := lipgloss.JoinVertical(lipgloss.Left, textLines...)

	playerStats := lipgloss.JoinHorizontal(lipgloss.Top, iconsBlock, " ", textsBlock)

	xpText := fmt.Sprintf("📈 XP: %d / %d", m.player.XP, m.player.NextLevelXP)
	xpBlock := lipgloss.JoinHorizontal(lipgloss.Bottom,
		lipgloss.NewStyle().PaddingRight(1).Render(xpText),
		m.progressBar.ViewAs(float64(m.player.XP)/float64(m.player.NextLevelXP)),
	)

	playerInfoContent := lipgloss.JoinVertical(
		lipgloss.Left,
		playerStats,
		"\n"+xpBlock,
	)

	playerInfoBox := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(1, 2).Render(playerInfoContent)

	var menuLines []string
	menuItems := []string{"Активные квесты", "Дерево перков", "Отправиться в данж", "Выход"}
	for i, item := range menuItems {
		cursor := " "
		if m.homepageCursor == i {
			cursor = ">"
		}
		menuLines = append(menuLines, fmt.Sprintf("%s %s", cursor, item))
	}
	menuContent := strings.Join(menuLines, "\n")
	menuBox := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("205")).Padding(1, 2).Render(menuContent)

	art := `

   ▄▄▄▄███▄▄▄▄      ▄████████    ▄██████▄  ███    █▄     ▄████████ 
 ▄██▀▀▀███▀▀▀██▄   ███    ███   ███    ███ ███    ███   ███    ███ 
 ███   ███   ███   ███    ███   ███    █▀  ███    ███   ███    █▀  
 ███   ███   ███   ███    ███  ▄███        ███    ███   ███        
 ███   ███   ███ ▀███████████ ▀▀███ ████▄  ███    ███ ▀███████████ 
 ███   ███   ███   ███    ███   ███    ███ ███    ███          ███ 
 ███   ███   ███   ███    ███   ███    ███ ███    ███    ▄█    ███ 
  ▀█   ███   █▀    ███    █▀    ████████▀  ████████▀   ▄████████▀  
                                                                   
`
	ui := lipgloss.JoinHorizontal(lipgloss.Top, playerInfoBox, menuBox)
	artBox := lipgloss.NewStyle().Align(lipgloss.Center).Width(m.terminalWidth).PaddingTop(1).Render(ansiGradient(art, [3]uint8{255, 0, 255}, [3]uint8{0, 0, 255}))
	return lipgloss.JoinVertical(lipgloss.Left, artBox, lipgloss.PlaceHorizontal(m.terminalWidth, lipgloss.Center, ui), lipgloss.NewStyle().Padding(1, 2).Render(m.statusMessage))
}