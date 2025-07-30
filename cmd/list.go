package cmd

import (
	"fmt"
	"magus/player"
	"magus/storage"
	"os"
	"strings"
)

func List() {
	quests, err := storage.LoadAllQuests()
	if err != nil {
		fmt.Println("❌ Ошибка загрузки квестов:", err)
		os.Exit(1)
	}

	if len(quests) == 0 {
		fmt.Println("✨ Нет активных квестов. Время добавить новый! `magus add`")
		return
	}

	fmt.Println("📜 Список квестов:")

	// Создаем карту для быстрого доступа к квестам по ID
	questMap := make(map[string]player.Quest)
	for _, q := range quests {
		questMap[q.ID] = q
	}

	// Создаем карту для группировки подзадач по родителям
	subQuests := make(map[string][]player.Quest)
	for _, q := range quests {
		if q.ParentID != "" {
			subQuests[q.ParentID] = append(subQuests[q.ParentID], q)
		}
	}

	// Отображаем только родительские квесты
	for _, q := range quests {
		if q.ParentID != "" {
			continue // Пропускаем подзадачи, они будут отображены под родителями
		}

		printQuest(q, 0) // 0 - уровень вложенности

		// Отображаем подзадачи для текущего квеста
		if children, ok := subQuests[q.ID]; ok {
			for _, child := range children {
				printQuest(child, 1) // 1 - уровень вложенности
			}
		}
	}
}

func printQuest(q player.Quest, indentationLevel int) {
	if q.Completed {
		return // Не показываем выполненные квесты
	}

	var status string
	// Для Ritual квестов статус всегда "активен", т.к. они повторяемые
	if q.Type == player.TypeRitual {
		status = "💧"
	} else if q.Progress > 0 && q.Progress < q.HP {
		status = "⚙️" // В процессе
	} else {
		status = "⏳" // Ожидает
	}

	indent := strings.Repeat("  ", indentationLevel)
	if indentationLevel > 0 {
		indent += "└─ "
	}

	var details string
	switch q.Type {
	case player.TypeFocus:
		details = fmt.Sprintf("(HP: %d/%d, XP: %d)", q.Progress, q.HP, q.XP)
	case player.TypeGoal:
		details = fmt.Sprintf("(XP: %d)", q.XP)
	case player.TypeRitual:
		details = fmt.Sprintf("(%s)", q.RitualSubtype)
	}

	fmt.Printf("%s%s [%s] %s %s {id: %s}\n",
		indent,
		status,
		strings.ToUpper(string(q.Type)),
		q.Title,
		details,
		q.ID)
}
