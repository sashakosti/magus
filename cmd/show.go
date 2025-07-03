package cmd

import (
	"encoding/json"
	"fmt"
	"magus/player"
	"os"
)

func Show() {
	file, err := os.ReadFile("data/player.json")
	if err != nil {
		fmt.Println("❌ Не удалось прочитать player.json:", err)
		return
	}

	var p player.Player
	err = json.Unmarshal(file, &p)
	if err != nil {
		fmt.Println("❌ Ошибка разбора JSON:", err)
		return
	}

	fmt.Printf("🧙 Имя: %s\n", p.Name)
    if p.Class != player.ClassNone {
        fmt.Printf("🎖️ Класс: %s\n", p.Class)
    }
    fmt.Printf("📈 Уровень: %d\n", p.Level)
	fmt.Printf("🔋 XP: %d / %d\n", p.XP, p.NextLevelXP)
	fmt.Printf("✨ Очки навыков: %d\n", p.SkillPoints)

	if len(p.Perks) > 0 {
		fmt.Println("🎁 Перки:")
		for _, perk := range p.Perks {
			fmt.Printf("  • %s\n", perk)
		}
	}

	if len(p.Skills) > 0 {
		fmt.Println("🧠 Навыки:")
		for skill, level := range p.Skills {
			fmt.Printf("  • %s: %d\n", skill, level)
		}
	}
}
