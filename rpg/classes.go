package rpg

import "magus/player"

type Class struct {
	Name        player.PlayerClass
	Description string
}

func GetAvailableClasses() []Class {
	return []Class{
		{
			Name:        player.ClassMage,
			Description: "+15% XP за квесты типа 'Arc' и 'Epic'.",
		},
		{
			Name:        player.ClassWarrior,
			Description: "+10% XP за выполнение всех 'Daily' квестов в течение дня.",
		},
		{
			Name:        player.ClassRogue,
			Description: "Получает на 1 очко навыков больше за каждый второй уровень.",
		},
	}
}
