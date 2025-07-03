package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"magus/player"
	"magus/storage"
	"os"
	"time"
)

func generateID() string {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func Add() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: magus add \"название задачи\" [--type=daily] [--xp=10] [--parent=ID]")
		return
	}
	title := os.Args[2]

	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	taskType := addCmd.String("type", "daily", "Тип квеста (daily, arc, meta, epic, chore)")
	xp := addCmd.Int("xp", 10, "Количество XP за квест")
	parentID := addCmd.String("parent", "", "ID родительского квеста")

	addCmd.Parse(os.Args[3:])

	newQuest := player.Quest{
		ID:        generateID(),
		ParentID:  *parentID,
		Title:     title,
		Type:      player.QuestType(*taskType),
		XP:        *xp,
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
		if err == nil && hasPerk(p, "Планирование") {
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
