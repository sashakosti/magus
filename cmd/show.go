package cmd

import (
	"fmt"
	"magus/player"
	"magus/rpg"
)

func Show() {
	p, err := player.LoadPlayer()
	if err != nil {
		if err == player.ErrPlayerNotFound {
			fmt.Println("🔮 Игрок не найден. Создайте его, запустив `magus` без аргументов.")
			return
		}
		fmt.Println("❌ Не удалось прочитать player.json:", err)
		return
	}

	fmt.Printf("🧙 Имя: %s\n", p.Name)
	if p.Class != player.ClassNone {
		fmt.Printf("🎖️ Класс: %s\n", p.Class)
	}
	fmt.Printf("📈 Уровень: %d\n", p.Level)
	fmt.Printf("🔋 XP: %d / %d\n", p.XP, p.NextLevelXP)
	fmt.Printf("✨ Очки навыков: %d\n", p.SkillPoints)

	if len(p.UnlockedSkills) > 0 {
		fmt.Println("🎁 Перки:")
		skillTrees, err := rpg.LoadSkillTrees(p)
		if err != nil {
			fmt.Println("  Не удалось загрузить дерево навыков:", err)
		} else {
			fmt.Println("\n--- Общие навыки ---")
			if len(skillTrees.Common) == 0 {
				fmt.Println("Нет доступных общих навыков.")
			} else {
				for _, node := range skillTrees.Common {
					unlocked := ""
					if rpg.IsSkillUnlocked(p, node.ID) {
						unlocked = "[ИЗУЧЕНО]"
					}
					fmt.Printf("- %s %s\n  %s\n", node.Name, unlocked, node.Description)
				}
			}

			fmt.Println("\n--- Классовые навыки ---")
			if len(skillTrees.Class) == 0 {
				fmt.Println("Нет доступных классовых навыков.")
			} else {
				for _, node := range skillTrees.Class {
					unlocked := ""
					if rpg.IsSkillUnlocked(p, node.ID) {
						unlocked = "[ИЗУЧЕНО]"
					}
					fmt.Printf("- %s %s\n  %s\n", node.Name, unlocked, node.Description)
				}
			}
		}
	}
}
