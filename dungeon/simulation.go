package dungeon

import (
	"fmt"
	"magus/player"
	"math/rand"
	"time"
)

// SimulationResult holds the outcome of a dungeon run
type SimulationResult struct {
	Events   []DungeonEvent
	XPGained int
	GoldGained int
	Completed bool // True if the timer finished, false if escaped
}

// RunSimulation simulates a dungeon run for a given duration.
func RunSimulation(p *player.Player, duration time.Duration) SimulationResult {
	rand.Seed(time.Now().UnixNano())
	
	var result SimulationResult
	endTime := time.Now().Add(duration)

	// Basic simulation loop
	// In a real implementation, this would be a more complex turn-based system.
	// For now, we'll generate some random events.

	result.Events = append(result.Events, DungeonEvent{
		Timestamp: time.Now(),
		Type:      EventTypeMessage,
		Message:   "Вы входите в темное подземелье...",
	})

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for time.Now().Before(endTime) {
		select {
		case t := <-ticker.C:
			monster := Monsters[rand.Intn(len(Monsters))] // Fight a random monster
			
			// Player attacks monster
			playerDamage := 10 + p.Skills["Сила Стихий"] * 2 // Example calculation
			result.Events = append(result.Events, DungeonEvent{
				Timestamp: t,
				Type:      EventTypePlayerAttack,
				Message:   fmt.Sprintf("Вы атакуете %s и наносите %d урона!", monster.Name, playerDamage),
			})

			// Monster attacks player
			monsterDamage := monster.Attack
			result.Events = append(result.Events, DungeonEvent{
				Timestamp: t,
				Type:      EventTypeMonsterAttack,
				Message:   fmt.Sprintf("%s атакует вас и наносит %d урона!", monster.Name, monsterDamage),
			})

			// Simulate gaining loot
			if rand.Intn(100) < 20 { // 20% chance to find loot
				xp := monster.XPValue
				gold := monster.GoldValue
				result.XPGained += xp
				result.GoldGained += gold
				result.Events = append(result.Events, DungeonEvent{
					Timestamp: t,
					Type:      EventTypeLoot,
					Message:   fmt.Sprintf("Вы победили %s и получили %d XP и %d золота!", monster.Name, xp, gold),
				})
			}
		}
	}

	result.Completed = true
	result.Events = append(result.Events, DungeonEvent{
		Timestamp: time.Now(),
		Type:      EventTypeMessage,
		Message:   "Вы у��пешно зачистили этаж!",
	})

	return result
}
