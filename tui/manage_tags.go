package tui

import (
	"fmt"
	"strings"

	"magus/player"
	"magus/storage"

	"github.com/charmbracelet/bubbletea"
)

func (m *Model) updateManageTags(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.renameTagInput.Focused() {
		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
			oldTag := m.allTags[m.tagCursor]
			newTag := m.renameTagInput.Value()
			if newTag != "" && newTag != oldTag {
				for i := range m.quests {
					for j, tag := range m.quests[i].Tags {
						if tag == oldTag {
							m.quests[i].Tags[j] = newTag
						}
					}
				}
				storage.SaveAllQuests(m.quests)
				m.buildQuestFilters() // Rebuild filters with new tag
			}
			m.renameTagInput.Blur()
			m.renameTagInput.Reset()
		} else {
			m.renameTagInput, cmd = m.renameTagInput.Update(msg)
		}
		return m, cmd
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "k":
			if m.tagCursor > 0 {
				m.tagCursor--
			}
		case "down", "j":
			if m.tagCursor < len(m.allTags)-1 {
				m.tagCursor++
			}
		case "d":
			if len(m.allTags) > 0 {
				tagToDelete := m.allTags[m.tagCursor]
				var updatedQuests []player.Quest
				for _, quest := range m.quests {
					var newTags []string
					for _, tag := range quest.Tags {
						if tag != tagToDelete {
							newTags = append(newTags, tag)
						}
					}
					quest.Tags = newTags
					updatedQuests = append(updatedQuests, quest)
				}
				m.quests = updatedQuests
				storage.SaveAllQuests(m.quests)
				m.buildQuestFilters() // Rebuild filters
				m.tagCursor = 0
			}
		case "r":
			if len(m.allTags) > 0 {
				m.renameTagInput.Focus()
				m.renameTagInput.SetValue(m.allTags[m.tagCursor])
				m.renameTagInput.CursorEnd()
			}
		case "q", "esc":
			m.state = stateQuestsFilter
			m.statusMessage = ""
		}
	}

	return m, nil
}

func (m *Model) viewManageTags() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Управление тегами") + "\n\n")

	if len(m.allTags) == 0 {
		b.WriteString("У вас пока нет тегов.\n")
	}

	for i, tag := range m.allTags {
		cursor := " "
		if m.tagCursor == i {
			cursor = ">"
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, tag))
	}

	if m.renameTagInput.Focused() {
		b.WriteString("\nПереименовать в: " + m.renameTagInput.View())
	}

	b.WriteString("\n\nНавигация: ↑/↓, 'd' - удалить, 'r' - переименовать, 'q' - назад.")
	return docStyle.Render(b.String())
}
