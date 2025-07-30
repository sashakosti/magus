package tui

import (
	"crypto/rand"
	"encoding/hex"
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

type AddQuestState struct {
	inputs         []textinput.Model
	focusIdx       int
	typeIdx        int
	subtypeIdx     int
	questTypes     []player.QuestType
	ritualSubtypes []player.RitualType
	parentId       string // To pre-fill if adding a sub-quest
}

const (
	fieldTitle = iota
	fieldType
	fieldRitualSubtype
	fieldHP
	fieldXP
	fieldTags
	fieldDeadline
	fieldButton
)

// numFields - теперь это не константа, а функция, зависящая от типа квеста
func (s *AddQuestState) numFields() int {
	switch s.questTypes[s.typeIdx] {
	case player.TypeGoal:
		return 3 // Title, Type, Button
	case player.TypeRitual:
		return 4 // Title, Type, Subtype, Button
	case player.TypeFocus:
		return 7 // Title, Type, HP, XP, Tags, Deadline, Button
	default:
		return 3
	}
}

func NewAddQuestState(m *Model, parentId ...string) State {
	inputs := make([]textinput.Model, 5) // Title, HP, XP, Tags, Deadline
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
		inputs[i].CharLimit = 120
	}

	inputs[0].Placeholder = "Название квеста"
	inputs[1].Placeholder = "100" // HP
	inputs[2].Placeholder = "50"  // XP
	inputs[3].Placeholder = "работа,дом"
	inputs[4].Placeholder = "ГГГГ-ММ-ДД"
	inputs[0].Focus()

	pid := ""
	if len(parentId) > 0 {
		pid = parentId[0]
	}

	return &AddQuestState{
		inputs:         inputs,
		questTypes:     []player.QuestType{player.TypeFocus, player.TypeRitual, player.TypeGoal},
		ritualSubtypes: []player.RitualType{player.RitualRestoration, player.RitualMaintenance},
		parentId:       pid,
	}
}

func (s *AddQuestState) Init() tea.Cmd {
	return textinput.Blink
}

func (s *AddQuestState) Update(m *Model, msg tea.Msg) (State, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "q", "esc":
			return PopState{}, nil
		case "tab", "down", "enter":
			if key.String() == "enter" && s.focusIdx == s.numFields()-1 {
				return s.saveQuest(m)
			}
			s.focusIdx = (s.focusIdx + 1) % s.numFields()
			return s.syncFocus()
		case "shift+tab", "up":
			s.focusIdx--
			if s.focusIdx < 0 {
				s.focusIdx = s.numFields() - 1
			}
			return s.syncFocus()
		case "left", "right":
			if s.focusIdx == fieldType {
				if key.String() == "left" {
					s.typeIdx--
					if s.typeIdx < 0 {
						s.typeIdx = len(s.questTypes) - 1
					}
				} else {
					s.typeIdx = (s.typeIdx + 1) % len(s.questTypes)
				}
				// Сбрасываем фокус, чтобы пересчитать количество полей
				s.focusIdx = fieldType
				return s.syncFocus()
			}
			if s.focusIdx == fieldRitualSubtype {
				if key.String() == "left" {
					s.subtypeIdx--
					if s.subtypeIdx < 0 {
						s.subtypeIdx = len(s.ritualSubtypes) - 1
					}
				} else {
					s.subtypeIdx = (s.subtypeIdx + 1) % len(s.ritualSubtypes)
				}
			}
		}
	}

	var cmds []tea.Cmd = make([]tea.Cmd, len(s.inputs))
	for i := range s.inputs {
		s.inputs[i], cmds[i] = s.inputs[i].Update(msg)
	}

	return s, tea.Batch(cmds...)
}

func (s *AddQuestState) View(m *Model) string {
	var b strings.Builder
	b.WriteString("📝 Новый квест\n\n")

	currentType := s.questTypes[s.typeIdx]

	// Title (always shown)
	b.WriteString(s.fieldView("Название", s.inputs[0], s.focusIdx == fieldTitle))
	// Type (always shown)
	b.WriteString(s.typeSelectorView("Тип", s.questTypes[s.typeIdx].String(), s.focusIdx == fieldType))

	switch currentType {
	case player.TypeFocus:
		b.WriteString(s.fieldView("HP", s.inputs[1], s.focusIdx == fieldHP))
		b.WriteString(s.fieldView("XP", s.inputs[2], s.focusIdx == fieldXP))
		b.WriteString(s.fieldView("Теги (через запятую)", s.inputs[3], s.focusIdx == fieldTags))
		b.WriteString(s.fieldView("Дедлайн (ГГГГ-ММ-ДД)", s.inputs[4], s.focusIdx == fieldDeadline))
	case player.TypeRitual:
		b.WriteString(s.typeSelectorView("Подтип", string(s.ritualSubtypes[s.subtypeIdx]), s.focusIdx == fieldRitualSubtype))
	case player.TypeGoal:
		// No extra fields needed
	}

	// Button
	saveButtonStyle := lipgloss.NewStyle().Padding(0, 1)
	if s.focusIdx == s.numFields()-1 {
		saveButtonStyle = saveButtonStyle.Background(lipgloss.Color("205")).Foreground(lipgloss.Color("0"))
	}
	b.WriteString(fmt.Sprintf("\n%s\n", saveButtonStyle.Render("[ Сохранить ]")))
	b.WriteString("\n" + m.styles.FaintQuestCardStyle.Render("esc - отмена"))

	return lipgloss.NewStyle().Margin(1, 2).Render(b.String())
}

func (s *AddQuestState) fieldView(label string, input textinput.Model, focused bool) string {
	cursor := "  "
	if focused {
		cursor = "> "
	}
	return fmt.Sprintf("%s%s\n  %s\n\n", cursor, label, input.View())
}

func (s *AddQuestState) typeSelectorView(label, value string, focused bool) string {
	cursor := "  "
	if focused {
		cursor = "> "
	}
	style := lipgloss.NewStyle()
	if focused {
		style = style.Foreground(lipgloss.Color("205"))
	}
	return fmt.Sprintf("%s%s\n  %s\n\n", cursor, label, style.Render(fmt.Sprintf("< %s >", value)))
}

func (s *AddQuestState) syncFocus() (*AddQuestState, tea.Cmd) {
	for i := range s.inputs {
		s.inputs[i].Blur()
	}

	var cmd tea.Cmd
	currentType := s.questTypes[s.typeIdx]

	switch s.focusIdx {
	case fieldTitle:
		cmd = s.inputs[0].Focus()
	case fieldHP:
		if currentType == player.TypeFocus {
			cmd = s.inputs[1].Focus()
		}
	case fieldXP:
		if currentType == player.TypeFocus {
			cmd = s.inputs[2].Focus()
		}
	case fieldTags:
		if currentType == player.TypeFocus {
			cmd = s.inputs[3].Focus()
		}
	case fieldDeadline:
		if currentType == player.TypeFocus {
			cmd = s.inputs[4].Focus()
		}
	}
	return s, cmd
}

func (s *AddQuestState) saveQuest(m *Model) (State, tea.Cmd) {
	title := s.inputs[0].Value()
	if title == "" {
		return s, nil // TODO: Show status message
	}

	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	id := hex.EncodeToString(bytes)

	newQuest := player.Quest{
		ID:        id,
		Title:     title,
		ParentID:  s.parentId,
		Type:      s.questTypes[s.typeIdx],
		CreatedAt: time.Now(),
	}

	switch newQuest.Type {
	case player.TypeFocus:
		hp, _ := strconv.Atoi(s.inputs[1].Value())
		if hp == 0 {
			hp = 100
		}
		newQuest.HP = hp

		xp, _ := strconv.Atoi(s.inputs[2].Value())
		if xp == 0 {
			xp = 10
		}
		newQuest.XP = xp

		tagsStr := s.inputs[3].Value()
		if tagsStr != "" {
			newQuest.Tags = strings.Split(tagsStr, ",")
		}

		deadlineStr := s.inputs[4].Value()
		if deadlineStr != "" {
			t, err := time.Parse("2006-01-02", deadlineStr)
			if err == nil {
				newQuest.Deadline = &t
			}
		}
	case player.TypeRitual:
		newQuest.RitualSubtype = s.ritualSubtypes[s.subtypeIdx]
		newQuest.XP = 0 // Ритуалы не дают опыта
	case player.TypeGoal:
		newQuest.XP = 100 // Цели дают много опыта при завершении
	}

	m.Quests = append(m.Quests, newQuest)
	storage.SaveAllQuests(m.Quests)

	// Возвращаемся и обновляем список квестов
	return PopState{refreshQuests: true}, nil
}
