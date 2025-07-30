package tui

import (
	"fmt"
	"magus/player"
	"magus/storage"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type EditQuestState struct {
	questToEdit player.Quest
	inputs      []textinput.Model
	focusIndex  int
}

func NewEditQuestState(m *Model, quest player.Quest) *EditQuestState {
	s := &EditQuestState{
		questToEdit: quest,
		inputs:      make([]textinput.Model, 4), // Title, XP, Tags, Deadline
	}

	var t textinput.Model

	t = textinput.New()
	t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	t.CharLimit = 100
	t.SetValue(s.questToEdit.Title)
	s.inputs[0] = t

	t = textinput.New()
	t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	t.CharLimit = 5
	t.SetValue(fmt.Sprintf("%d", s.questToEdit.XP))
	s.inputs[1] = t

	t = textinput.New()
	t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	t.CharLimit = 100
	t.SetValue(strings.Join(s.questToEdit.Tags, ", "))
	s.inputs[2] = t

	t = textinput.New()
	t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	t.CharLimit = 10
	deadlineStr := ""
	if s.questToEdit.Deadline != nil && !s.questToEdit.Deadline.IsZero() {
		deadlineStr = s.questToEdit.Deadline.Format("2006-01-02")
	}
	t.SetValue(deadlineStr)
	t.Placeholder = "YYYY-MM-DD"
	s.inputs[3] = t

	s.inputs[s.focusIndex].Focus()
	return s
}

func (s *EditQuestState) Init() tea.Cmd {
	return textinput.Blink
}

func (s *EditQuestState) Update(m *Model, msg tea.Msg) (State, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "ctrl+c", "esc":
			return PopState{}, nil
		case "tab", "shift+tab", "up", "down":
			if key.String() == "up" || key.String() == "shift+tab" {
				s.focusIndex--
			} else {
				s.focusIndex++
			}

			if s.focusIndex > len(s.inputs)-1 {
				s.focusIndex = 0
			} else if s.focusIndex < 0 {
				s.focusIndex = len(s.inputs) - 1
			}

			cmds := make([]tea.Cmd, len(s.inputs))
			for i := 0; i <= len(s.inputs)-1; i++ {
				if i == s.focusIndex {
					cmds[i] = s.inputs[i].Focus()
					continue
				}
				s.inputs[i].Blur()
			}
			return s, tea.Batch(cmds...)
		case "enter":
			return s.saveChanges(m)
		}
	}

	cmd := s.updateInputs(msg)
	return s, cmd
}

func (s *EditQuestState) View(m *Model) string {
	var b strings.Builder
	b.WriteString(m.styles.TitleStyle.Render("Редактирование квеста") + "\n\n")

	b.WriteString("Название\n")
	b.WriteString(s.inputs[0].View())
	b.WriteString("\n\n")

	b.WriteString("XP\n")
	b.WriteString(s.inputs[1].View())
	b.WriteString("\n\n")

	b.WriteString("Теги (через запятую)\n")
	b.WriteString(s.inputs[2].View())
	b.WriteString("\n\n")

	b.WriteString("Дедлайн (YYYY-MM-DD)\n")
	b.WriteString(s.inputs[3].View())
	b.WriteString("\n\n")

	b.WriteString(m.styles.FaintQuestCardStyle.Render("Enter - сохранить, Esc - отмена"))
	return lipgloss.NewStyle().Margin(1, 2).Render(b.String())
}

func (s *EditQuestState) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(s.inputs))
	for i := range s.inputs {
		s.inputs[i], cmds[i] = s.inputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (s *EditQuestState) saveChanges(m *Model) (State, tea.Cmd) {
	for i, q := range m.Quests {
		if q.ID == s.questToEdit.ID {
			m.Quests[i].Title = s.inputs[0].Value()
			xp, _ := strconv.Atoi(s.inputs[1].Value())
			m.Quests[i].XP = xp

			tagsStr := s.inputs[2].Value()
			if tagsStr == "" {
				m.Quests[i].Tags = []string{}
			} else {
				m.Quests[i].Tags = strings.Split(tagsStr, ",")
				for j, tag := range m.Quests[i].Tags {
					m.Quests[i].Tags[j] = strings.TrimSpace(tag)
				}
			}

			deadlineStr := s.inputs[3].Value()
			if deadlineStr != "" {
				dl, err := time.Parse("2006-01-02", deadlineStr)
				if err == nil {
					m.Quests[i].Deadline = &dl
				}
			} else {
				m.Quests[i].Deadline = nil
			}
			break
		}
	}

	storage.SaveAllQuests(m.Quests)
	return PopState{}, nil
}
