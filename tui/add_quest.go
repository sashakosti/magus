package tui

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"magus/player"
	"magus/storage"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) initAddQuest() {
	m.addQuestCursor = 0
	m.addQuestTypeIdx = 0
	m.addQuestInputs = make([]textinput.Model, 5)
	for i := range m.addQuestInputs {
		m.addQuestInputs[i] = textinput.New()
		m.addQuestInputs[i].CursorStyle = cursorStyle
		m.addQuestInputs[i].CharLimit = 120
	}
	m.addQuestInputs[0].Placeholder = "Название квеста"
	m.addQuestInputs[1].Placeholder = "10"
	m.addQuestInputs[2].Placeholder = "работа,дом"
	m.addQuestInputs[3].Placeholder = "ГГГГ-ММ-ДД"
	m.addQuestInputs[4].Placeholder = "ID родителя (опционально)"
	m.addQuestInputs[0].Focus()
}

func (m *Model) updateAddQuest(msg tea.Msg) (tea.Model, tea.Cmd) {
	const numItems = 7

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "tab", "down":
			m.addQuestCursor = (m.addQuestCursor + 1) % numItems
			return m.focusAddQuestInputs()
		case "shift+tab", "up":
			m.addQuestCursor--
			if m.addQuestCursor < 0 {
				m.addQuestCursor = numItems - 1
			}
			return m.focusAddQuestInputs()
		case "left", "right":
			if m.addQuestCursor == 1 {
				if key.String() == "left" {
					m.addQuestTypeIdx--
					if m.addQuestTypeIdx < 0 {
						m.addQuestTypeIdx = len(m.addQuestTypes) - 1
					}
				} else {
					m.addQuestTypeIdx = (m.addQuestTypeIdx + 1) % len(m.addQuestTypes)
				}
			}
		case "enter":
			if m.addQuestCursor == numItems-1 {
				return m.saveQuest()
			}
			m.addQuestCursor = (m.addQuestCursor + 1) % numItems
			return m.focusAddQuestInputs()
		}
	}

	var cmds []tea.Cmd
	for i := range m.addQuestInputs {
		var singleCmd tea.Cmd
		m.addQuestInputs[i], singleCmd = m.addQuestInputs[i].Update(msg)
		cmds = append(cmds, singleCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) focusAddQuestInputs() (*Model, tea.Cmd) {
	for i := range m.addQuestInputs {
		m.addQuestInputs[i].Blur()
	}

	if m.addQuestCursor == 0 {
		return m, m.addQuestInputs[0].Focus()
	} else if m.addQuestCursor > 1 && m.addQuestCursor < 6 {
		return m, m.addQuestInputs[m.addQuestCursor-1].Focus()
	}
	return m, nil
}

func (m *Model) saveQuest() (*Model, tea.Cmd) {
	title := m.addQuestInputs[0].Value()
	if title == "" {
		m.statusMessage = "Название квеста не может быть пустым."
		return m, nil
	}

	xp, _ := strconv.Atoi(m.addQuestInputs[1].Value())
	if xp == 0 {
		xp = 10
	}

	tagsStr := m.addQuestInputs[2].Value()
	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
	}

	deadlineStr := m.addQuestInputs[3].Value()
	var deadline *time.Time
	if deadlineStr != "" {
		t, err := time.Parse("2006-01-02", deadlineStr)
		if err == nil {
			deadline = &t
		}
	}

	parentID := m.addQuestInputs[4].Value()

	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	id := hex.EncodeToString(bytes)

	newQuest := player.Quest{
		ID:        id,
		Title:     title,
		XP:        xp,
		Tags:      tags,
		Deadline:  deadline,
		ParentID:  parentID,
		Type:      m.addQuestTypes[m.addQuestTypeIdx],
		CreatedAt: time.Now(),
	}

	m.quests = append(m.quests, newQuest)
	storage.SaveAllQuests(m.quests)
	m.sortAndBuildDisplayQuests()
	m.state = stateQuests
	m.statusMessage = fmt.Sprintf("Квест '%s' добавлен!", title)
	m.addQuestInputs = nil
	return m, nil
}

func (m *Model) viewAddQuest() string {
	var b strings.Builder
	b.WriteString("📝 Новый квест\n\n")

	fields := []string{"Название", "Тип", "XP", "Теги (через запятую)", "Дедлайн (ГГГГ-ММ-ДД)", "Родительский ID"}
	for i, field := range fields {
		cursor := "  "
		if m.addQuestCursor == i {
			cursor = "> "
		}
		b.WriteString(cursor + field + "\n")

		if i == 1 {
			style := lipgloss.NewStyle()
			if m.addQuestCursor == i {
				style = style.Foreground(lipgloss.Color("205"))
			}
			b.WriteString("  " + style.Render(fmt.Sprintf("< %s >", m.addQuestTypes[m.addQuestTypeIdx])) + "\n\n")
		} else {
			inputIndex := i
			if i > 1 {
				inputIndex--
			}
			b.WriteString("  " + m.addQuestInputs[inputIndex].View() + "\n\n")
		}
	}

	saveButtonStyle := lipgloss.NewStyle().Padding(0, 1)
	if m.addQuestCursor == len(fields) {
		saveButtonStyle = saveButtonStyle.Background(lipgloss.Color("205")).Foreground(lipgloss.Color("0"))
	}
	b.WriteString(fmt.Sprintf("\n%s\n", saveButtonStyle.Render("[ Сохранить ]")))

	return b.String()
}
