package cmd

import (
	"fmt"
	"magus/player"
	"time"
)

func Version() {
	fmt.Println("🧙 Magus v0.1.0")
}

// isToday проверяет, является ли дата сегодняшней.
func isToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}

func hasPerk(p *player.Player, perkName string) bool {
	for _, ownedPerk := range p.Perks {
		if ownedPerk == perkName {
			return true
		}
	}
	return false
}
