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
		fmt.Println("Usage: magus add \"название задачи\" [--type=daily] [--xp=10]")
		return
	}
	title := os.Args[2]

	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	taskType := addCmd.String("type", "daily", "Тип квеста (daily, arc, meta)")
	xp := addCmd.Int("xp", 10, "Количество XP за квест")

	addCmd.Parse(os.Args[3:])

	newQuest := player.Quest{
		ID:        generateID(),
		Title:     title,
		Type:      player.QuestType(*taskType),
		XP:        *xp,
		Completed: false,
		CreatedAt: time.Now(),
	}

	quests, err := storage.LoadAllQuests()
	if err != nil {
		fmt.Println("❌ Ошибка з��грузки квестов:", err)
		return
	}

	quests = append(quests, newQuest)

	if err := storage.SaveAllQuests(quests); err != nil {
		fmt.Println("❌ Ошибка сохранения квеста:", err)
		return
	}

	fmt.Println("🗒️ Добавлен квест:", title)
}
