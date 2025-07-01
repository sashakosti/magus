package cmd

import (
	"fmt"
	"magus/player"
	"magus/storage"
	"os"
	"time"
)

func Complete() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: magus complete <quest_id>")
		return
	}
	questID := os.Args[2]

	quests, err := storage.LoadAllQuests()
	if err != nil {
		fmt.Println("❌ Ошибка загрузки квестов:", err)
		return
	}

	var found bool
	var xpGained int
	for i, q := range quests {
		if q.ID == questID {
			if q.Type == player.Daily {
				if isToday(q.CompletedAt) {
					fmt.Println("⚠️ Этот дейлик уже выполнен сегодня.")
					return
				}
				quests[i].CompletedAt = time.Now()
			} else {
				if q.Completed {
					fmt.Println("⚠️ Квест уже был выполнен ранее.")
					return
				}
				quests[i].Completed = true
			}

			xpGained = q.XP
			found = true
			break
		}
	}

	if !found {
		fmt.Println("⚠️ Квест с таким ID не найден.")
		return
	}

	if err := storage.SaveAllQuests(quests); err != nil {
		fmt.Println("❌ Ошибка сохранения квестов:", err)
		return
	}

	fmt.Println("✅ Квест завершён!")

	if xpGained > 0 {
		canLevelUp, err := player.AddXP(xpGained)
		if err != nil {
			fmt.Println("❌ Не удалось начислить XP:", err)
		} else {
			fmt.Printf("✨ +%d XP!\n", xpGained)
			if canLevelUp {
				fmt.Println("🔥 Поздравляем! Вы можете повысить уровень! Запустите `magus` для выбора перка.")
			}
		}
	}
}


