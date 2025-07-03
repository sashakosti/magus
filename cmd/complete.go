package cmd

import (
	"fmt"
	"magus/player"
	"magus/storage"
	"os"
	"strings"
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
	var xpGained int

	// Находим и завершаем квест
	for i, q := range quests {
		if q.ID == questID {
			if q.Completed || (q.Type == player.Daily && isToday(q.CompletedAt)) {
				fmt.Println("⚠️ Квест уже выполнен.")
				return
			}
			quests[i].Completed = true
			quests[i].CompletedAt = time.Now()
			xpGained = q.XP
			completedQuest = quests[i]
			found = true
			break
		}
	}

	if !found {
		fmt.Println("⚠️ Квест с таким ID не найден.")
		return
	}

	// Сохраняем изменения
	if err := storage.SaveAllQuests(quests); err != nil {
		fmt.Println("❌ Ошибка сохранения квестов:", err)
		return
	}
	fmt.Println("✅ Квест завершён!")
	addXP(xpGained, completedQuest.Type)

	// Проверяем, не нужно ли завершить родительский квест
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

	var subQuestsCompleted = true
	var parentQuestIndex = -1

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
			fmt.Printf("🎉 Все подзадачи выполнены! Родительский квест '%s' завершён!\n", parent.Title)
            addXP(parent.XP, parent.Type) // Начисляем XP за родительский квест с учетом типа

			if err := storage.SaveAllQuests(quests); err != nil {
				fmt.Println("❌ Ошибка сохранения родительского квеста:", err)
			}
		}
	}
}

// addXP начисляет опыт с учетом классовых бонусов и перков, и обрабатывает повышение уровня.
func addXP(xp int, questType player.QuestType) {
	if xp <= 0 {
		return
	}

	p, err := player.LoadPlayer()
	if err != nil {
		fmt.Println("❌ Ошибка загрузки игрока для начисления XP:", err)
		return
	}

	totalXP := xp
	bonusMessages := []string{}

	// 1. Классовые бонусы
	classBonus := 0
	switch p.Class {
	case player.ClassMage:
		if questType == player.Arc || questType == player.Epic {
			bonus := 15
			if hasPerk(p, "Магический резонанс") {
				bonus = 25
			}
			classBonus = totalXP * bonus / 100
			if classBonus > 0 {
				bonusMessages = append(bonusMessages, fmt.Sprintf("+%d бонус класса", classBonus))
			}
		}
	case player.ClassWarrior:
		if hasPerk(p, "Боевой раж") && questType == player.Daily {
			classBonus = 5
			bonusMessages = append(bonusMessages, fmt.Sprintf("+%d бонус перка", classBonus))
		}
	}
	totalXP += classBonus

	// 2. Бонусы от перков
	perkBonus := 0
	if hasPerk(p, "Фокус") && isFirstQuestOfDay(p) {
		perkBonus += 5
		bonusMessages = append(bonusMessages, "+5 Фокус")
	}
	if hasPerk(p, "Комбо-стрик") && time.Since(p.LastCompletedAt).Hours() < 1 {
		perkBonus += 5
		bonusMessages = append(bonusMessages, "+5 Комбо")
	}
	totalXP += perkBonus

	// Обновляем время последнего квеста и сохраняем
	p.LastCompletedAt = time.Now()
	player.SavePlayer(p)

	// Начисляем итоговый опыт
	canLevelUp, err := player.AddXP(totalXP)
	if err != nil {
		fmt.Println("❌ Не удалось начислить XP:", err)
	} else {
		if len(bonusMessages) > 0 {
			fmt.Printf("✨ +%d XP (%s)!\n", totalXP, strings.Join(bonusMessages, ", "))
		} else {
			fmt.Printf("✨ +%d XP!\n", totalXP)
		}
		if canLevelUp {
			fmt.Println("🔥 Поздравляем! Вы можете повысить уровень! Запустите `magus` для выбора перка или класса.")
		}
	}
}

func isFirstQuestOfDay(p *player.Player) bool {
	return !isToday(p.LastCompletedAt)
}



