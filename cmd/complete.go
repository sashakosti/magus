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

	var completedQuest player.Quest
	var found bool

	for i, q := range quests {
		if q.ID == questID {
			if q.Completed {
				fmt.Println("⚠️ Квест уже выполнен.")
				return
			}

			switch q.Type {
			case player.TypeGoal:
				fmt.Println("⚠️ Цели (Goal) нельзя завершить напрямую. Завершите все подзадачи.")
				return
			case player.TypeRitual:
				fmt.Println("💧 Ритуал выполнен. Мана восстановлена (в TUI).")
				// Логика начисления маны находится в TUI, здесь просто сообщение
			case player.TypeFocus:
				quests[i].Completed = true
				quests[i].CompletedAt = time.Now()
				quests[i].Progress = q.HP // Считаем выполненным
				addXP(q.XP)
				fmt.Println("✅ Квест завершён!")
			}

			completedQuest = quests[i]
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

	if completedQuest.ParentID != "" {
		checkAndCompleteParent(completedQuest.ParentID)
	}
}

// checkAndCompleteParent проверяет, все ли дочерние квесты выполнены, и завершает родительский.
func checkAndCompleteParent(parentID string) {
	quests, err := storage.LoadAllQuests()
	if err != nil {
		fmt.Println("❌ Ошибка загрузки квестов для проверки родительского:", err)
		return
	}

	subQuestsCompleted := true
	parentQuestIndex := -1

	for i, q := range quests {
		if q.ID == parentID {
			parentQuestIndex = i
			continue
		}
		if q.ParentID == parentID && !q.Completed {
			subQuestsCompleted = false
			break
		}
	}

	if parentQuestIndex != -1 && subQuestsCompleted {
		parent := &quests[parentQuestIndex]
		if !parent.Completed {
			parent.Completed = true
			parent.CompletedAt = time.Now()
			fmt.Printf("🎉 Все подзадачи выполнены! Родительский квест '%s' завершён!", parent.Title)
			addXP(parent.XP)

			if err := storage.SaveAllQuests(quests); err != nil {
				fmt.Println("❌ Ошибка сохранения родительского квеста:", err)
			}
		}
	}
}

// addXP начисляет опыт и обрабатывает повышение уровня.
func addXP(xp int) {
	if xp <= 0 {
		return
	}

	canLevelUp, err := player.AddXP(xp)
	if err != nil {
		fmt.Println("❌ Не удалось начислить XP:", err)
	} else {
		fmt.Printf("✨ +%d XP!\n", xp)
		if canLevelUp {
			fmt.Println("🔥 Поздравляем! Вы можете повысить уровень! Запустите `magus` для выбора перка или класса.")
		}
	}
}

func isFirstQuestOfDay(p *player.Player) bool {
	return !isToday(p.LastCompletedAt)
}
