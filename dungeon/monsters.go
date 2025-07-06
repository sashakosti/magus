package dungeon

// Monster represents a creature in the dungeon
type Monster struct {
	Name       string
	HP         int
	MaxHP      int
	Attack     int
	Defense    int
	XPValue    int
	GoldValue  int
}

// Pre-defined monsters
var Monsters = []Monster{
	{
		Name:      "Гоблин",
		HP:        30,
		MaxHP:     30,
		Attack:    5,
		Defense:   2,
		XPValue:   10,
		GoldValue: 5,
	},
	{
		Name:      "Скелет",
		HP:        50,
		MaxHP:     50,
		Attack:    8,
		Defense:   5,
		XPValue:   15,
		GoldValue: 10,
	},
	{
		Name:      "Огр",
		HP:        100,
		MaxHP:     100,
		Attack:    15,
		Defense:   8,
		XPValue:   50,
		GoldValue: 25,
	},
}
