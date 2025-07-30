package tui

import (
	"fmt"
	"magus/player"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// HomepageState представляет собой состояние главного экрана.
type HomepageState struct {
	cursor      int
	progressBar progress.Model
}

func NewHomepageState(m *Model) *HomepageState {
	p := progress.New(progress.WithDefaultGradient(), progress.WithoutPercentage())
	return &HomepageState{
		cursor:      0,
		progressBar: p,
	}
}

func (s *HomepageState) Init() tea.Cmd {
	return nil
}

func (s *HomepageState) Update(m *Model, msg tea.Msg) (State, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// This is crucial: we need to update the progress bar's width
		// based on the available space.
		playerInfoBoxWidth := lipgloss.Width(s.playerInfoView(m.Player))
		s.progressBar.Width = playerInfoBoxWidth - 20 // Subtract padding and text length
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.cursor < 3 { // 4 пункта меню: 0, 1, 2, 3
				s.cursor++
			}
		case "enter":
			switch s.cursor {
			case 0:
				return NewQuestsState(m), nil
			case 1:
				return NewSkillsState(m), nil
			case 2:
				return NewDungeonPrepState(m), nil
			case 3:
				return s, tea.Quit
			}
		case "q", "esc":
			return s, tea.Quit
		}
	}

	var cmd tea.Cmd
	var progCmd tea.Cmd
	newProgressBar, progCmd := s.progressBar.Update(msg)
	s.progressBar = newProgressBar.(progress.Model)
	cmd = progCmd
	return s, cmd
}

func (s *HomepageState) playerInfoView(p *player.Player) string {
	lines := []struct {
		icon string
		text string
	}{
		{"🧙", fmt.Sprintf("Игрок: %s (Уровень: %d)", p.Name, p.Level)},
		{"🛡", fmt.Sprintf("Класс: %s", p.Class)},
		{"❤", fmt.Sprintf("HP: %d / %d", p.HP, p.MaxHP)},
		{"💧", fmt.Sprintf("Мана: %d / %d", p.Mana, p.MaxMana)},
		{"💰", fmt.Sprintf("Золото: %d", p.Gold)},
		{"🎁", fmt.Sprintf("Навыки: %d", len(p.UnlockedSkills))},
		{"✨", fmt.Sprintf("Очки навыков: %d", p.SkillPoints)},
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
		if (line.icon == "🛡" && p.Class == player.ClassNone) ||
			(line.icon == "🎁" && len(p.UnlockedSkills) == 0) ||
			(line.icon == "✨" && p.SkillPoints == 0) {
			continue
		}
		padding := strings.Repeat(" ", maxIconWidth-runewidth.StringWidth(line.icon))
		iconLines = append(iconLines, line.icon+padding)
		textLines = append(textLines, line.text)
	}

	iconsBlock := lipgloss.JoinVertical(lipgloss.Left, iconLines...)
	textsBlock := lipgloss.JoinVertical(lipgloss.Left, textLines...)
	playerStats := lipgloss.JoinHorizontal(lipgloss.Top, iconsBlock, " ", textsBlock)

	xpText := fmt.Sprintf("📈 XP: %d / %d", p.XP, p.NextLevelXP)
	xpBlock := lipgloss.JoinHorizontal(lipgloss.Bottom,
		lipgloss.NewStyle().PaddingRight(1).Render(xpText),
		s.progressBar.ViewAs(float64(p.XP)/float64(p.NextLevelXP)),
	)

	playerInfoContent := lipgloss.JoinVertical(lipgloss.Left, playerStats, "\n"+xpBlock)
	return lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(1, 2).Render(playerInfoContent)
}

func (s *HomepageState) View(m *Model) string {
	playerInfoBox := s.playerInfoView(m.Player)

	var menuLines []string
	menuItems := []string{"Активные квесты", "Дерево навыков", "Отправиться в данж", "Выход"}
	for i, item := range menuItems {
		cursor := " "
		if s.cursor == i {
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
	artBox := lipgloss.NewStyle().Align(lipgloss.Center).Width(m.TerminalWidth).PaddingTop(1).Render(ansiGradient(art, [3]uint8{255, 0, 255}, [3]uint8{0, 0, 255}))
	
	// TODO: Добавить статусное сообщение из глобальной модели
	return lipgloss.JoinVertical(lipgloss.Left, artBox, lipgloss.PlaceHorizontal(m.TerminalWidth, lipgloss.Center, ui))
}
