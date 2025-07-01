package rpg

import (
	"fmt"
	"magus/player"
)

// XPPerLevel определяет XP, необходимый для перехода на каждый уровень
func XPPerLevel(level int) int {
	return 100 + (level-1)*50 // например: 100, 150, 200, ...
}

// AddXP начисляет XP, проверяет левелап, добавляет перки
func AddXP(p *player.Player, amount int) (leveledUp bool, gainedPerks []string) {
	fmt.Printf("+%d XP!\n", amount)
	p.XP += amount

	// проверка на левелап
	leveledUp = false
	for p.XP >= p.NextLevelXP {
		p.XP -= p.NextLevelXP
		p.Level++
		leveledUp = true
		p.NextLevelXP = XPPerLevel(p.Level)

		fmt.Printf("🔥 Уровень повышен! Текущий уровень: %d\n", p.Level)

		
	}

	return
}
