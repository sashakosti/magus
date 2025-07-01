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
		fmt.Println("❌ ��шибка загрузки квестов:", err)
		os.Exit(1)
	}

	if len(quests) == 0 {
		fmt.Println("✨ Нет активных квестов. Время добавить новый! `magus add`")
		return
	}

	fmt.Println("📜 Список квестов:")
	for _, q := range quests {
		var status string
		if q.Type == player.Daily {
			if isToday(q.CompletedAt) {
				status = "✅"
			} else {
				status = "⏳"
			}
		} else {
			if q.Completed {
				continue // Не показываем выполненные сюжетные квесты
			}
			status = "⏳"
		}

		fmt.Printf("  %s [%s] %s (XP: %d) {id: %s}\n",
			status,
			strings.ToUpper(string(q.Type)),
			q.Title,
			q.XP,
			q.ID)
	}
}
