package tui

import (
	"fmt"
	"sort"
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
		m.viewport.Width = msg.Width - 30
		m.viewport.Height = msg.Height - 10
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
					m.pushState(stateManageTags)
					return m, nil
				}
				if filter == "---" {
					return m, nil
				}
				m.activeQuestFilter = filter
				m.state = stateQuests // Не меняем стек, т.к. это внутреннее переключение
				m.cursor = 0
				m.sortAndBuildDisplayQuests()
				m.viewport.GotoTop()
			}
		case stateQuests:
			switch msg.String() {
			case "left", "h":
				m.state = stateQuestsFilter // Не меняем стек, т.к. это внутреннее переключение
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
						// Загружаем все дерево навыков
						skillTree, err := rpg.LoadSkillTree(&m.player)
						if err == nil {
							var availableSkills []player.SkillNode
							for _, node := range skillTree {
								if rpg.IsSkillAvailable(&m.player, node) {
									availableSkills = append(availableSkills, node)
								}
							}

							if len(availableSkills) > 0 {
								m.pushState(stateLevelUp)
								m.skillChoices = availableSkills
							} else {
								// Если доступных навыков нет, просто повышаем уровень
								player.LevelUpPlayer("")
								p, _ := player.LoadPlayer()
								m.player = *p
								m.statusMessage = "🔮 Новый уровень! Доступных для изучения навыков пока нет."
							}
						} else {
							// Если произошла ошибка загрузки дерева, просто повышаем уровень
							player.LevelUpPlayer("")
							p, _ := player.LoadPlayer()
							m.player = *p
							m.statusMessage = "🔮 Новый уровень! Ошибка загрузки дерева навыков."
						}
					}
				} else {
					p, _ := player.LoadPlayer()
					m.player = *p
					m.statusMessage = fmt.Sprintf("✅ '%s' выполнено!", quest.Title)
				}
			}
		}
	}

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
	var filterPanel strings.Builder
	filterPanel.WriteString(titleStyle.Render("Фильтры") + "\n\n")
	for i, f := range m.questFilters {
		style := filterStyle
		if i == m.questFilterCursor && m.state == stateQuestsFilter {
			style = selectedFilterStyle
		}
		filterPanel.WriteString(style.Render(f) + "\n")
	}

	mainView := lipgloss.JoinHorizontal(lipgloss.Top, filterPanel.String(), m.viewport.View())

	return docStyle.Render(mainView)
}

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
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *Model) viewCompletedQuests() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("✅ Завершенные квесты") + "\n\n")

	completedQuests := []player.Quest{}
	totalXP := 0
	for _, quest := range m.quests {
		if quest.Completed || (quest.Type == player.Daily && !quest.CompletedAt.IsZero()) {
			completedQuests = append(completedQuests, quest)
			if quest.Completed { // Daily completed today might not have the 'Completed' flag set
				totalXP += quest.XP
			}
		}
	}

	if len(completedQuests) == 0 {
		m.viewport.SetContent("Пока нет завершенных квестов.")
		return m.viewport.View()
	}

	// Sort by completion date, newest first
	sort.Slice(completedQuests, func(i, j int) bool {
		return completedQuests[i].CompletedAt.After(completedQuests[j].CompletedAt)
	})

	b.WriteString(fmt.Sprintf("Всего завершено: %d | Суммарный опыт: %d\n\n", len(completedQuests), totalXP))

	groupedByDate := make(map[string][]player.Quest)
	var order []string
	now := time.Now()

	for _, q := range completedQuests {
		dateStr := q.CompletedAt.Format("2 January 2006")
		if isToday(q.CompletedAt) {
			dateStr = "Сегодня"
		} else if isYesterday(q.CompletedAt, now) {
			dateStr = "Вчера"
		}
		
		if _, ok := groupedByDate[dateStr]; !ok {
			order = append(order, dateStr)
		}
		groupedByDate[dateStr] = append(groupedByDate[dateStr], q)
	}

	for _, date := range order {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#36A2EB")).Render(date) + "\n")
		for _, q := range groupedByDate[date] {
			b.WriteString(fmt.Sprintf("  - %s [%s] (XP: %d)\n", q.Title, q.Type, q.XP))
		}
		b.WriteString("\n")
	}
	
	m.viewport.SetContent(b.String())
	return m.viewport.View()
}


func (m *Model) sortAndBuildDisplayQuests() {
	allQuests, _ := storage.LoadAllQuests()
	m.quests = allQuests
	questMap := make(map[string]player.Quest)
	for _, q := range allQuests {
		questMap[q.ID] = q
	}

	filteredQuestIDs := make(map[string]bool)
	now := time.Now()

	// Step 1: Initial filter pass
	for _, q := range m.quests {
		isCompleted := q.Completed || (q.Type == player.Daily && isToday(q.CompletedAt))
		passesFilter := false

		switch m.activeQuestFilter {
		case "Все":
			if !isCompleted {
				passesFilter = true
			}
		case "Daily":
			if q.Type == player.Daily && !isToday(q.CompletedAt) {
				passesFilter = true
			}
		case "Завершенные":
			if isCompleted {
				passesFilter = true
			}
		case "Просроченные":
			if q.Deadline != nil && !q.Deadline.IsZero() && now.After(*q.Deadline) && !isCompleted {
				passesFilter = true
			}
		default: // Tag filter
			if !isCompleted {
				for _, tag := range q.Tags {
					if tag == m.activeQuestFilter {
						passesFilter = true
						break
					}
				}
			}
		}

		if passesFilter {
			filteredQuestIDs[q.ID] = true
		}
	}

	// Step 2: Preserve hierarchy by including all parents
	finalQuestIDs := make(map[string]bool)
	for id := range filteredQuestIDs {
		currID := id
		for currID != "" {
			if _, exists := finalQuestIDs[currID]; exists {
				break // This chain is already included
			}
			finalQuestIDs[currID] = true
			quest, ok := questMap[currID]
			if !ok {
				break // Should not happen in a consistent dataset
			}
			currID = quest.ParentID
		}
	}

	// Step 3: Build the hierarchical list for display
	var parentQuests []player.Quest
	childQuests := make(map[string][]player.Quest)

	for id := range finalQuestIDs {
		quest := questMap[id]
		if quest.ParentID == "" {
			parentQuests = append(parentQuests, quest)
		} else {
			if _, parentExists := finalQuestIDs[quest.ParentID]; parentExists {
				childQuests[quest.ParentID] = append(childQuests[quest.ParentID], quest)
			}
		}
	}
	
	// Sort parents by creation date
	sort.Slice(parentQuests, func(i, j int) bool {
		return parentQuests[i].CreatedAt.Before(parentQuests[j].CreatedAt)
	})


	// Step 4: Assemble the final list, respecting collapsed state
	var displayQuests []player.Quest
	for _, p := range parentQuests {
		if m.activeQuestFilter == "Daily" && p.Type != player.Daily {
			// We need to check all descendants, not just immediate children
			var checkDescendants func(string) bool
			checkDescendants = func(parentID string) bool {
				children, ok := childQuests[parentID]
				if !ok {
					return false
				}
				for _, child := range children {
					if child.Type == player.Daily {
						if _, ok := finalQuestIDs[child.ID]; ok {
							return true
						}
					}
					if checkDescendants(child.ID) {
						return true
					}
				}
				return false
			}

			if !checkDescendants(p.ID) {
				continue
			}
		}

		displayQuests = append(displayQuests, p)
		if !m.collapsed[p.ID] {
			// Sort children by creation date before adding
			if children, ok := childQuests[p.ID]; ok {
				sort.Slice(children, func(i, j int) bool {
					return children[i].CreatedAt.Before(children[j].CreatedAt)
				})
				displayQuests = append(displayQuests, children...)
			}
		}
	}

	m.displayQuests = displayQuests
}
