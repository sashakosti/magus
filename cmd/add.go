package cmd

import (
	"flag"
	"fmt"
	"magus/player"
	"magus/storage"
	"magus/utils"
	"os"
	"strings"
	"time"
)

func Add() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: magus add \"название задачи\" [--type=daily] [--xp=10] [--parent=ID] [--tags=\"tag1,tag2\"] [--deadline=\"YYYY-MM-DD\"]")
		return
	}
	title := os.Args[2]

	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	taskType := addCmd.String("type", "daily", "Тип квеста (daily, arc, meta, epic, chore)")
	xp := addCmd.Int("xp", 10, "Количество XP за квест")
	parentID := addCmd.String("parent", "", "ID родительского квеста")
	tagsStr := addCmd.String("tags", "", "Теги через запятую (e.g., \"работа,дом\")")
	deadlineStr := addCmd.String("deadline", "", "Дедлайн в формате YYYY-MM-DD")

	addCmd.Parse(os.Args[3:])

	var tags []string
	if *tagsStr != "" {
		tags = strings.Split(*tagsStr, ",")
	}

	var deadline *time.Time
	if *deadlineStr != "" {
		t, err := time.Parse("2006-01-02", *deadlineStr)
		if err != nil {
			fmt.Println("❌ Ошибка парсинга дедлайна. Используйте формат YYYY-MM-DD:", err)
			return
		}
		deadline = &t
	}

	newQuest := player.Quest{
		ID:        utils.GenerateID(),
		ParentID:  *parentID,
		Title:     title,
		Type:      player.QuestType(*taskType),
		XP:        *xp,
		Tags:      tags,
		Deadline:  deadline,
		Completed: false,
		CreatedAt: time.Now(),
	}

	quests, err := storage.LoadAllQuests()
	if err != nil {
		fmt.Println("❌ Ошибка загрузки квестов:", err)
		return
	}

	quests = append(quests, newQuest)

	// Применяем перк "Планирование"
	if *parentID != "" {
		p, err := player.LoadPlayer()
		if err != nil {
			if err == player.ErrPlayerNotFound {
				// Игнорируем ошибку, если игрок не найден, перк просто не применяется
			} else {
				fmt.Println("❌ Ошибка загрузки игрока для применения перка:", err)
			}
		} else if hasPerk(p, "Планирование") {
			for i, q := range quests {
				if q.ID == *parentID {
					bonusXP := q.XP * 20 / 100
					quests[i].XP += bonusXP
					fmt.Printf("✨ Перк 'Планирование': +%d XP к родительскому квесту!\n", bonusXP)
					break
				}
			}
		}
	}

	if err := storage.SaveAllQuests(quests); err != nil {
		fmt.Println("❌ Ошибка сохранения квеста:", err)
		return
	}

	fmt.Println("🗒️ Добавлен квест:", title)
	if *parentID != "" {
		fmt.Printf("   (Подзадача для квеста %s)\n", *parentID)
	}
}
