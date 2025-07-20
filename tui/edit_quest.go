package tui

import (
	"fmt"
	"magus/storage"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
)

// initQuestEdit sets up the model for editing an existing quest.
func (m *Model) initQuestEdit() {
	m.editInputs = make([]textinput.Model, 4) // Title, XP, Tags, Deadline
	m.editFocusIndex = 0

	if m.cursor >= len(m.displayQuests) {
		// Should not happen, but as a safeguard
		m.state = stateQuests
		m.statusMessage = "Ошибка: не удалось найти квест для редактирования."
		return
	}
	m.editingQuest = m.displayQuests[m.cursor]

	var t textinput.Model

	t = textinput.New()
	t.CursorStyle = cursorStyle
	t.CharLimit = 100
	t.SetValue(m.editingQuest.Title)
	m.editInputs[0] = t

	t = textinput.New()
	t.CursorStyle = cursorStyle
	t.CharLimit = 5
	t.SetValue(fmt.Sprintf("%d", m.editingQuest.XP))
	m.editInputs[1] = t

	t = textinput.New()
	t.CursorStyle = cursorStyle
	t.CharLimit = 100
	t.SetValue(strings.Join(m.editingQuest.Tags, ", "))
	m.editInputs[2] = t

	t = textinput.New()
	t.CursorStyle = cursorStyle
	t.CharLimit = 10
	deadlineStr := ""
	if m.editingQuest.Deadline != nil && !m.editingQuest.Deadline.IsZero() {
		deadlineStr = m.editingQuest.Deadline.Format("2006-01-02")
	}
	t.SetValue(deadlineStr)
	t.Placeholder = "YYYY-MM-DD"
	m.editInputs[3] = t

	m.editInputs[m.editFocusIndex].Focus()
}

func (m *Model) updateQuestEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.state = stateQuests
			m.statusMessage = "Редактирование отменено."
			return m, nil
		case "tab", "shift+tab", "up", "down":
			s := msg.String()
			if s == "up" || s == "shift+tab" {
				m.editFocusIndex--
			} else {
				m.editFocusIndex++
			}

			if m.editFocusIndex > len(m.editInputs)-1 {
				m.editFocusIndex = 0
			} else if m.editFocusIndex < 0 {
				m.editFocusIndex = len(m.editInputs) - 1
			}

			cmds := make([]tea.Cmd, len(m.editInputs))
			for i := 0; i <= len(m.editInputs)-1; i++ {
				if i == m.editFocusIndex {
					cmds[i] = m.editInputs[i].Focus()
					continue
				}
				m.editInputs[i].Blur()
			}
			return m, tea.Batch(cmds...)
		case "enter":
			// Find the quest in the main list and update it
			for i, q := range m.quests {
				if q.ID == m.editingQuest.ID {
					m.quests[i].Title = m.editInputs[0].Value()
					xp, _ := strconv.Atoi(m.editInputs[1].Value())
					m.quests[i].XP = xp
					
					tagsStr := m.editInputs[2].Value()
					if tagsStr == "" {
						m.quests[i].Tags = []string{}
					} else {
						m.quests[i].Tags = strings.Split(tagsStr, ",")
						for j, tag := range m.quests[i].Tags {
							m.quests[i].Tags[j] = strings.TrimSpace(tag)
						}
					}

					deadlineStr := m.editInputs[3].Value()
					if deadlineStr != "" {
						dl, err := time.Parse("2006-01-02", deadlineStr)
						if err == nil {
							m.quests[i].Deadline = &dl
						}
					} else {
						m.quests[i].Deadline = nil
					}
					
					break
				}
			}

			storage.SaveAllQuests(m.quests)
			m.sortAndBuildDisplayQuests()
			m.state = stateQuests
			m.statusMessage = fmt.Sprintf("Квест '%s' обновлен.", m.editingQuest.Title)
			return m, nil
		}
	}

	cmd := m.updateEditInputs(msg)
	return m, cmd
}

func (m *Model) viewQuestEdit() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Редактирование квеста") + "\n\n")
	
	b.WriteString("Название\n")
	b.WriteString(m.editInputs[0].View())
	b.WriteString("\n\n")

	b.WriteString("XP\n")
	b.WriteString(m.editInputs[1].View())
	b.WriteString("\n\n")

	b.WriteString("Теги (через запятую)\n")
	b.WriteString(m.editInputs[2].View())
	b.WriteString("\n\n")

	b.WriteString("Дедлайн (YYYY-MM-DD)\n")
	b.WriteString(m.editInputs[3].View())
	b.WriteString("\n\n")

	b.WriteString(faintQuestCardStyle.Render("Enter - сохранить, Esc - отмена"))
	return docStyle.Render(b.String())
}

func (m *Model) updateEditInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.editInputs))
	for i := range m.editInputs {
		m.editInputs[i], cmds[i] = m.editInputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}
