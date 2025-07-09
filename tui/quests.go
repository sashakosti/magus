package tui

import (
	"fmt"
	"strings"
	"time"

	"magus/player"
	"magus/rpg"
	"magus/storage"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func (m *Model) updateQuests(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.displayQuests)-1 {
				m.cursor++
			}
		case " ":
			if m.cursor < len(m.displayQuests) {
				quest := m.displayQuests[m.cursor]
				if quest.ParentID == "" {
					m.collapsed[quest.ID] = !m.collapsed[quest.ID]
					m.sortAndBuildDisplayQuests()
				}
			}
		case "enter":
			if m.cursor >= len(m.displayQuests) {
				return m, nil
			}
			quest := m.displayQuests[m.cursor]
			if (quest.Type == player.Daily && isToday(quest.CompletedAt)) || quest.Completed {
				return m, nil
			}

			var xpGained int
			for i, q := range m.quests {
				if q.ID == quest.ID {
					if m.quests[i].Type == player.Daily {
						m.quests[i].CompletedAt = time.Now()
					} else {
						m.quests[i].Completed = true
					}
					xpGained = m.quests[i].XP
					break
				}
			}

			storage.SaveAllQuests(m.quests)
			m.sortAndBuildDisplayQuests()

			if xpGained > 0 {
				hpHealed := 0
				if quest.Type == player.Chore {
					hpHealed = xpGained / 2
					m.player.HP += hpHealed
					if m.player.HP > m.player.MaxHP {
						m.player.HP = m.player.MaxHP
					}
				}

				canLevelUp, _ := player.AddXP(xpGained)
				p, _ := player.LoadPlayer()
				m.player = *p

				if hpHealed > 0 {
					m.statusMessage = fmt.Sprintf("Квест '%s' выполнен! +%d XP, +%d HP", quest.Title, xpGained, hpHealed)
				} else {
					m.statusMessage = fmt.Sprintf("Квест '%s' выполнен! +%d XP", quest.Title, xpGained)
				}

				if canLevelUp {
					perkChoices, _ := rpg.GetPerkChoices(&m.player)
					if len(perkChoices) > 0 {
						m.state = stateLevelUp
						m.perkChoices = perkChoices
					} else {
						player.LevelUpPlayer("")
						p, _ := player.LoadPlayer()
						m.player = *p
						m.statusMessage = "Новый уровень! Доступных перков пока нет."
					}
					m.cursor = 0
				}
			} else {
				p, _ := player.LoadPlayer()
				m.player = *p
				m.statusMessage = fmt.Sprintf("Квест '%s' выполнен!", quest.Title)
			}
		}
	}
	return m, nil
}

func (m *Model) viewQuests() string {
	s := "📜 Активные квесты\n\n"
	for i, quest := range m.displayQuests {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		icon := "⏳"
		style := lipgloss.NewStyle()
		if quest.Completed || (quest.Type == player.Daily && isToday(quest.CompletedAt)) {
			icon = "✅"
			style = style.Strikethrough(true).Faint(true)
		}

		indent := ""
		collapseIcon := " "
		if quest.ParentID == "" {
			isParent := false
			for _, q := range m.quests {
				if q.ParentID == quest.ID {
					isParent = true
					break
				}
			}
			if isParent {
				collapseIcon = "⊖"
				if m.collapsed[quest.ID] {
					collapseIcon = "⊕"
				}
			}
		} else {
			indent = "  └─ "
		}

		tags := ""
		for _, tag := range quest.Tags {
			tags += fmt.Sprintf(" [#%s]", tag)
		}

		iconWithPadding := icon + strings.Repeat(" ", 2-runewidth.StringWidth(icon))
		collapseIconWithPadding := collapseIcon + strings.Repeat(" ", 2-runewidth.StringWidth(collapseIcon))

		s += style.Render(fmt.Sprintf("%s %s%s%s[%s] %s%s %s", cursor, indent, iconWithPadding, collapseIconWithPadding, quest.Type, quest.Title, tags, deadlineStatus(quest.Deadline))) + "\n"
	}

	s += fmt.Sprintf("\n%s\n", m.statusMessage)
	return s
}

func (m *Model) updateCompletedQuests(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *Model) viewCompletedQuests() string {
	s := "✅ Завершенные квесты\n\n"
	found := false
	for _, quest := range m.quests {
		if quest.Completed && quest.Type != player.Daily {
			s += fmt.Sprintf("  - %s [%s] (XP: %d)\n", quest.Title, quest.Type, quest.XP)
			found = true
		}
	}
	if !found {
		s += "Пока нет завершенных квестов."
	}
	s += "\nНажмите 'q' для возврата."
	return s
}
