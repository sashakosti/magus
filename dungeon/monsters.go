package dungeon

import (
	"math/rand"
	"time"
)

type Monster struct {
	Name     string
	HP       int
	MaxHP    int
	Attack   int
	Defense  int
	XPValue  int
	AsciiArt string
}

var monsterTemplates = []Monster{
	{
		Name:     "Goblin",
		HP:       10,
		Attack:   1, // Было 2
		Defense:  1,
		XPValue:  5,
		AsciiArt: " G ",
	},
	{
		Name:     "Orc",
		HP:       20,
		Attack:   2, // Было 5
		Defense:  3,
		XPValue:  10,
		AsciiArt: " O ",
	},
	{
		Name:     "Troll",
		HP:       35,
		Attack:   4, // Было 7
		Defense:  5,
		XPValue:  20,
		AsciiArt: " T ",
	},
}

func GenerateMonster(floor int) Monster {
	rand.Seed(time.Now().UnixNano())
	template := monsterTemplates[rand.Intn(len(monsterTemplates))]

	scaleFactor := 1.0 + float64(floor)*0.2

	monster := template
	monster.MaxHP = int(float64(template.HP) * scaleFactor)
	monster.HP = monster.MaxHP
	monster.Attack = int(float64(template.Attack) * scaleFactor)
	monster.Defense = int(float64(template.Defense) * scaleFactor)
	monster.XPValue = int(float64(template.XPValue) * scaleFactor)

	return monster
}