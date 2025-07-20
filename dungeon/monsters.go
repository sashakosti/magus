package dungeon

import (
	"fmt"
	"math/rand"
)

// Monster represents a creature in the dungeon
type Monster struct {
	Name      string
	HP        int
	MaxHP     int
	Attack    int
	Defense   int
	XPValue   int
	GoldValue int
	IsBoss    bool
}

// Pre-defined monsters
var Monsters = []Monster{
	{Name: "Гоблин", HP: 30, MaxHP: 30, Attack: 5, Defense: 2, XPValue: 10, GoldValue: 5},
	{Name: "Скелет", HP: 50, MaxHP: 50, Attack: 8, Defense: 5, XPValue: 15, GoldValue: 10},
	{Name: "Огр", HP: 100, MaxHP: 100, Attack: 15, Defense: 8, XPValue: 50, GoldValue: 25},
}

// Pre-defined bosses
var Bosses = []Monster{
	{Name: "Король Гоблинов", HP: 200, MaxHP: 200, Attack: 25, Defense: 15, XPValue: 200, GoldValue: 100, IsBoss: true},
	{Name: "Лич", HP: 350, MaxHP: 350, Attack: 40, Defense: 20, XPValue: 500, GoldValue: 250, IsBoss: true},
}

// getScaledMonster creates a monster instance and scales its stats.
func getScaledMonster(baseMonster Monster, floor int) Monster {
	monster := baseMonster

	// Linear scaling factor. The difficulty ramps up steadily.
	// e.g., floor 1: 1.0, floor 2: 1.15, floor 3: 1.30, etc.
	scaleFactor := 1.0 + 0.15*float64(floor-1)

	monster.HP = int(float64(baseMonster.HP) * scaleFactor)
	monster.MaxHP = int(float64(baseMonster.MaxHP) * scaleFactor)
	monster.Attack = int(float64(baseMonster.Attack) * scaleFactor)
	monster.Defense = int(float64(baseMonster.Defense) * scaleFactor)
	monster.XPValue = int(float64(baseMonster.XPValue) * scaleFactor)
	monster.GoldValue = int(float64(baseMonster.GoldValue) * scaleFactor)

	// Add floor level to name for clarity
	if monster.IsBoss {
		monster.Name = fmt.Sprintf("%s (%d этаж)", baseMonster.Name, floor)
	} else {
		monster.Name = fmt.Sprintf("%s (%d этаж)", baseMonster.Name, floor)
	}

	return monster
}

// GetMonsterForFloor selects a random monster or a boss and scales its stats.
func GetMonsterForFloor(floor int) Monster {
	// Every 5th floor is a boss floor
	if floor%5 == 0 {
		// Select a boss. We can cycle through them or pick randomly.
		// For now, let's pick one based on the floor number.
		bossIndex := (floor/5 - 1) % len(Bosses)
		baseBoss := Bosses[bossIndex]
		return getScaledMonster(baseBoss, floor)
	}

	// Otherwise, return a regular monster
	baseMonster := Monsters[rand.Intn(len(Monsters))]
	return getScaledMonster(baseMonster, floor)
}
