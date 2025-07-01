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
	fmt.Printf("📈 Уровень: %d\n", p.Level)
	fmt.Printf("🔋 XP: %d / %d\n", p.XP, p.NextLevelXP)
	fmt.Println("🎁 Перки:")
	/* for _, perk := range p.Perks {
		desc := player.GetPerkDescription(perk)
		fmt.Printf("  • %s — %s\n", perk, desc)
	} */
}
