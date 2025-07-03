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
	if q.Completed && q.Type != player.Daily {
		return // Не показываем выполненные квесты, кроме дейликов
	}

	var status string
	if q.Completed || (q.Type == player.Daily && isToday(q.CompletedAt)) {
		status = "✅"
	} else {
		status = "⏳"
	}

	indent := strings.Repeat("  ", indentationLevel)
	if indentationLevel > 0 {
		indent += "└─ "
	}

	fmt.Printf("%s%s [%s] %s (XP: %d) {id: %s}\n",
		indent,
		status,
		strings.ToUpper(string(q.Type)),
		q.Title,
		q.XP,
		q.ID)
}
