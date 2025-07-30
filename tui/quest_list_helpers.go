package tui

import (
	"fmt"
	"io"
	"magus/player"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// QuestListItem - это обертка над player.Quest для управления состоянием в UI
type QuestListItem struct {
	player.Quest
	IsExpanded bool
	Depth      int
	HasKids    bool
}

// Переопределяем FilterValue, чтобы list.Model мог использовать его
func (i QuestListItem) FilterValue() string {
	return i.Quest.FilterValue()
}

type QuestDelegate struct {
	Styles *Styles
}

func NewQuestDelegate(styles *Styles) list.ItemDelegate {
	return &QuestDelegate{Styles: styles}
}

// Height рассчитывает реальную высоту элемента: 2 строки текста + 2 строки рамки
func (d QuestDelegate) Height() int                               { return 4 }
func (d QuestDelegate) Spacing() int                              { return 0 }
func (d QuestDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d QuestDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(QuestListItem)
	if !ok {
		return
	}

	isCompleted := item.Completed
	// Для ритуалов "завершение" может сбрасывать��я, поэтому нужна другая логика,
	// но пока оставим так для консистентности.
	isSelected := index == m.Index()

	cardStyle := d.Styles.QuestCardStyle
	if isSelected {
		cardStyle = d.Styles.SelectedQuestCardStyle
	} else if isCompleted {
		cardStyle = d.Styles.FaintQuestCardStyle
	}

	var questTypeStyle lipgloss.Style
	var icon string
	var info []string

	switch item.Type {
	case player.TypeGoal:
		questTypeStyle = d.Styles.MetaStyle // Используем стиль для мета-целей
		icon = d.Styles.GoalIcon
		info = append(info, questTypeStyle.Render("[Цель]"))
	case player.TypeRitual:
		questTypeStyle = d.Styles.RitualStyle // Используем стиль для ритуалов/рутины
		icon = d.Styles.RitualIcon
		info = append(info, questTypeStyle.Render(fmt.Sprintf("[%s]", item.RitualSubtype)))
	case player.TypeFocus:
		questTypeStyle = d.Styles.FocusStyle
		icon = d.Styles.FocusIcon
		info = append(info, questTypeStyle.Render("[Фокус]"))
		if item.HP > 0 {
			progress := fmt.Sprintf("HP: %d/%d", item.Progress, item.HP)
			info = append(info, d.Styles.DifficultyStyle.Render(progress))
		}
	}

	if isCompleted {
		icon = d.Styles.CompletedIcon
	}

	// Иконка раскрытия/сворачивания
	expander := " "
	if item.HasKids {
		if item.IsExpanded {
			expander = d.Styles.CollapseIconOpened
		} else {
			expander = d.Styles.CollapseIconClosed
		}
	}

	var content strings.Builder
	titleLine := fmt.Sprintf("%s %s %s", expander, icon, item.Title)
	content.WriteString(titleLine)
	content.WriteString("\n")

	if len(item.Tags) > 0 {
		info = append(info, d.Styles.TagStyle.Render("#"+strings.Join(item.Tags, " #")))
	}
	if dl := deadlineStatus(item.Deadline); dl != "" {
		info = append(info, d.Styles.DeadlineStyle.Render(dl))
	}
	content.WriteString("  " + strings.Join(info, "  "))

	indentStr := strings.Repeat(d.Styles.SubQuestIndent, item.Depth)
	cardWidth := m.Width() - lipgloss.Width(indentStr) - cardStyle.GetHorizontalFrameSize()
	cardStyle = cardStyle.Width(cardWidth)

	cardRender := cardStyle.Render(content.String())
	finalRender := lipgloss.JoinHorizontal(lipgloss.Top, indentStr, cardRender)

	fmt.Fprint(w, finalRender)
}

// BuildQuestListItems создает плоский список QuestListItem из иерархии квестов
func BuildQuestListItems(allQuests []player.Quest, existingItems []list.Item) []list.Item {
	var items []list.Item
	questMap := make(map[string][]player.Quest)
	for _, q := range allQuests {
		questMap[q.ParentID] = append(questMap[q.ParentID], q)
	}

	// Сохраняем состояние isExpanded из существующего списка
	expandedState := make(map[string]bool)
	for _, item := range existingItems {
		if qli, ok := item.(QuestListItem); ok {
			expandedState[qli.ID] = qli.IsExpanded
		}
	}

	var addChildren func(parentId string, depth int)
	addChildren = func(parentId string, depth int) {
		children, ok := questMap[parentId]
		if !ok {
			return
		}
		for _, q := range children {
			_, hasKids := questMap[q.ID]
			isExpanded := expandedState[q.ID] // Получаем сохраненное состояние

			items = append(items, QuestListItem{
				Quest:      q,
				Depth:      depth,
				HasKids:    hasKids,
				IsExpanded: isExpanded,
			})
			if isExpanded {
				addChildren(q.ID, depth+1)
			}
		}
	}

	addChildren("", 0) // Начинаем с квестов верхнего уровня
	return items
}
