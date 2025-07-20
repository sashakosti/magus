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
		skillTree, err := rpg.LoadSkillTree(p)
		if err != nil {
			fmt.Println("  Не удалось загрузить дерево перков:", err)
		} else {
			for _, skillID := range p.UnlockedSkills {
				if skill, ok := skillTree[skillID]; ok {
					fmt.Printf("  • %s %s\n", skill.Icon, skill.Name)
				}
			}
		}
	}
}
