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
)

func (m *Model) updateQuests(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// This is important to ensure the viewport is aware of the terminal size.
		m.viewport.Width = msg.Width - 30 // Leave space for filter panel
		m.viewport.Height = msg.Height - 10 // Leave space for header/footer
	case tea.KeyMsg:
		switch m.state {
		case stateQuestsFilter:
			switch msg.String() {
			case "up", "k":
				if m.questFilterCursor > 0 {
					m.questFilterCursor--
				}
			case "down", "j":
				if m.questFilterCursor < len(m.questFilters)-1 {
					m.questFilterCursor++
				}
			case "enter", "right", "l":
                filter := m.questFilters[m.questFilterCursor]
                if filter == "[Управление тегами]" {
                    m.state = stateManageTags
                    m.tagCursor = 0
                    return m, nil
                }
                if filter == "---" {
                    return m, nil // Do nothing for separator
                }
                m.activeQuestFilter = filter
                m.state = stateQuests
                m.cursor = 0 // Reset quest cursor
                m.sortAndBuildDisplayQuests()
                m.viewport.GotoTop()
            }
        case stateQuests:
            switch msg.String() {
            case "left", "h":
                m.state = stateQuestsFilter
                m.statusMessage = ""
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
                if (quest.Type == player.Daily && isToday(quest.CompletedAt)) || (quest.Completed && quest.Type != player.Chore) {
                    return m, nil
                }

                if quest.Type == player.Epic {
                    childrenIncomplete := false
                    for _, q := range m.quests {
                        if q.ParentID == quest.ID && !q.Completed {
                            childrenIncomplete = true
                            break
                        }
                    }
                    if childrenIncomplete {
                        m.statusMessage = "❗ Нельзя завершить эпический квест, пока не выполнены все его части."
                        return m, nil
                    }
                }

                var xpGained int
                for i, q := range m.quests {
                    if q.ID == quest.ID {
                        if m.quests[i].Type == player.Chore {
                            // Chores are repeatable, no state change, just reward
                        } else if m.quests[i].Type == player.Daily {
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
                        m.statusMessage = fmt.Sprintf("✨ '%s' выполнено! Получено %d XP и восстановлено %d HP.", quest.Title, xpGained, hpHealed)
                    } else {
                        m.statusMessage = fmt.Sprintf("✨ '%s' выполнено! Получено %d XP.", quest.Title, xpGained)
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
                            m.statusMessage = "🔮 Новый уровень! Доступных перков пока нет."
                        }
                        m.cursor = 0
                    }
                } else {
                    p, _ := player.LoadPlayer()
                    m.player = *p
                    m.statusMessage = fmt.Sprintf("✅ '%s' выполнено!", quest.Title)
                }
            }
        }
    }

    // Update viewport content
    m.viewport.SetContent(m.renderQuestList())
    m.viewport, cmd = m.viewport.Update(msg)
    cmds = append(cmds, cmd)

    return m, tea.Batch(cmds...)
}

var (
    docStyle               = lipgloss.NewStyle().Margin(1, 2)
    titleStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFDF5")).Bold(true).Padding(0, 1)
    questCardStyle         = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(lipgloss.Color("240")).PaddingLeft(2).MarginBottom(1).Width(60)
    selectedQuestCardStyle = questCardStyle.Copy().Border(lipgloss.ThickBorder(), false, false, false, true).BorderForeground(lipgloss.Color("#AD58B4"))
    faintQuestCardStyle    = questCardStyle.Copy().Foreground(lipgloss.Color("240")).BorderForeground(lipgloss.Color("238"))
    dailyStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color("#36A2EB"))
    choreStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFC300"))
    questStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color("#9B59B6"))
    tagStyle               = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Background(lipgloss.Color("236")).Padding(0, 1)
    completedIcon          = lipgloss.NewStyle().Foreground(lipgloss.Color("#2ECC71")).Render("✔")
    pendingIcon            = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("◈")
    collapseIconOpened     = "▼"
    collapseIconClosed     = "▶"
    subQuestIndent         = "   "
    statusMessageStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#A89F94")).Italic(true)
    deadlineStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("#E74C3C"))
    filterStyle            = lipgloss.NewStyle().Padding(0, 1).MarginRight(2)
    selectedFilterStyle    = filterStyle.Copy().Background(lipgloss.Color("#513B56")).Foreground(lipgloss.Color("#FFFDF5"))
)

func (m *Model) viewQuests() string {
    // --- Filter Panel ---
    var filterPanel strings.Builder
    filterPanel.WriteString(titleStyle.Render("Фильтры") + "\n\n")
    for i, f := range m.questFilters {
        style := filterStyle
        if i == m.questFilterCursor && m.state == stateQuestsFilter {
            style = selectedFilterStyle
        }
        filterPanel.WriteString(style.Render(f) + "\n")
    }

    // --- Main View ---
    // The viewport now handles the quest list rendering
    mainView := lipgloss.JoinHorizontal(lipgloss.Top, filterPanel.String(), m.viewport.View())

    return docStyle.Render(mainView)
}

// renderQuestList generates the string content for the viewport
func (m *Model) renderQuestList() string {
    var b strings.Builder
    header := titleStyle.Render(fmt.Sprintf("~ %s ~", m.activeQuestFilter))
    b.WriteString(header + "\n\n")

    if len(m.displayQuests) == 0 {
        b.WriteString("Нет квестов, соответствующих фильтру.")
    }

    for i, quest := range m.displayQuests {
        isCompleted := quest.Completed || (quest.Type == player.Daily && isToday(quest.CompletedAt))
        isSelected := m.cursor == i && m.state == stateQuests

        cardStyle := questCardStyle
        if isSelected {
            cardStyle = selectedQuestCardStyle
        } else if isCompleted {
            cardStyle = faintQuestCardStyle
        }

        var questTypeStyle lipgloss.Style
        switch quest.Type {
        case player.Daily:
            questTypeStyle = dailyStyle
        case player.Chore:
            questTypeStyle = choreStyle
        case player.Arc, player.Meta, player.Epic:
            questTypeStyle = questStyle
        }

        icon := pendingIcon
        if isCompleted {
            icon = completedIcon
        }

        var content strings.Builder
        titleLine := fmt.Sprintf("%s %s %s", icon, quest.Title, questTypeStyle.Render(fmt.Sprintf("[%s]", quest.Type)))
        content.WriteString(titleLine)
        content.WriteString("\n")

        var infoLine []string
        if len(quest.Tags) > 0 {
            var renderedTags []string
            for _, tag := range quest.Tags {
                renderedTags = append(renderedTags, tagStyle.Render("#"+tag))
            }
            infoLine = append(infoLine, strings.Join(renderedTags, " "))
        }
        if dl := deadlineStatus(quest.Deadline); dl != "" {
            infoLine = append(infoLine, deadlineStyle.Render(dl))
        }
        if len(infoLine) > 0 {
            content.WriteString("  " + strings.Join(infoLine, "  "))
            content.WriteString("\n")
        }

        indentStr := " "
        if quest.ParentID != "" {
            indentStr = subQuestIndent + "└─ "
        } else {
            isParent := false
            for _, q := range m.quests {
                if q.ParentID == quest.ID {
                    isParent = true
                    break
                }
            }
            if isParent {
                if m.collapsed[quest.ID] {
                    indentStr += collapseIconClosed
                } else {
                    indentStr += collapseIconOpened
                }
            }
        }

        cardRender := cardStyle.Render(content.String())
        finalRender := lipgloss.JoinHorizontal(lipgloss.Top, indentStr+" ", cardRender)
        b.WriteString(finalRender + "\n")
    }

    b.WriteString("\n" + statusMessageStyle.Render(m.statusMessage) + "\n")
    return b.String()
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
