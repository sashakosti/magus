package tui

import (
	"fmt"
	"sort"
	"strings"

	"magus/player"
	"magus/storage"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ManageTagsState struct {
	cursor         int
	allTags        []string
	renameTagInput textinput.Model
}

func NewManageTagsState(m *Model) *ManageTagsState {
	s := &ManageTagsState{}
	s.buildTagList(m)

	ti := textinput.New()
	ti.Placeholder = "Новое имя тега"
	ti.CharLimit = 30
	s.renameTagInput = ti

	return s
}

func (s *ManageTagsState) buildTagList(m *Model) {
	tagSet := make(map[string]bool)
	for _, q := range m.Quests {
		for _, tag := range q.Tags {
			tagSet[tag] = true
		}
	}
	s.allTags = make([]string, 0, len(tagSet))
	for tag := range tagSet {
		s.allTags = append(s.allTags, tag)
	}
	sort.Strings(s.allTags)
}

func (s *ManageTagsState) Init() tea.Cmd {
	return nil
}

func (s *ManageTagsState) Update(m *Model, msg tea.Msg) (State, tea.Cmd) {
	var cmd tea.Cmd

	if s.renameTagInput.Focused() {
		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
			oldTag := s.allTags[s.cursor]
			newTag := s.renameTagInput.Value()
			if newTag != "" && newTag != oldTag {
				for i := range m.Quests {
					for j, tag := range m.Quests[i].Tags {
						if tag == oldTag {
							m.Quests[i].Tags[j] = newTag
						}
					}
				}
				storage.SaveAllQuests(m.Quests)
				s.buildTagList(m) // Rebuild our own list
			}
			s.renameTagInput.Blur()
			s.renameTagInput.Reset()
		} else {
			s.renameTagInput, cmd = s.renameTagInput.Update(msg)
		}
		return s, cmd
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.cursor < len(s.allTags)-1 {
				s.cursor++
			}
		case "d":
			if len(s.allTags) > 0 {
				tagToDelete := s.allTags[s.cursor]
				var updatedQuests []player.Quest
				for _, quest := range m.Quests {
					var newTags []string
					for _, tag := range quest.Tags {
						if tag != tagToDelete {
							newTags = append(newTags, tag)
						}
					}
					quest.Tags = newTags
					updatedQuests = append(updatedQuests, quest)
				}
				m.Quests = updatedQuests
				storage.SaveAllQuests(m.Quests)
				s.buildTagList(m) // Rebuild
				if s.cursor >= len(s.allTags) && len(s.allTags) > 0 {
					s.cursor = len(s.allTags) - 1
				}
			}
		case "r":
			if len(s.allTags) > 0 {
				s.renameTagInput.Focus()
				s.renameTagInput.SetValue(s.allTags[s.cursor])
				s.renameTagInput.CursorEnd()
			}
		case "q", "esc":
			return PopState{}, nil // Signal to pop the state
		}
	}

	return s, nil
}

func (s *ManageTagsState) View(m *Model) string {
	var b strings.Builder
	b.WriteString(m.styles.TitleStyle.Render("Управление тегами") + "\n\n")

	if len(s.allTags) == 0 {
		b.WriteString("У вас пока нет тегов.\n")
	}

	for i, tag := range s.allTags {
		cursor := " "
		style := lipgloss.NewStyle()
		if s.cursor == i {
			cursor = ">"
			style = style.Foreground(lipgloss.Color("205"))
		}
		b.WriteString(style.Render(fmt.Sprintf("%s %s\n", cursor, tag)))
	}

	if s.renameTagInput.Focused() {
		b.WriteString("\nПереименовать в: " + s.renameTagInput.View())
	}

	b.WriteString("\n\nНавигация: ↑/↓, 'd' - удалить, 'r' - переименовать, 'q' - назад.")
	return lipgloss.NewStyle().Margin(1, 2).Render(b.String())
}
