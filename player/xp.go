package player

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

const playerFile = "data/player.json"

// AddXP добавляет опыт игроку и возвращает true, если можно повысить уровень.
func AddXP(xp int) (bool, error) {
	p, err := LoadPlayer()
	if err != nil {
		return false, err
	}

	p.XP += xp
	p.History.QuestsCompleted++
	p.History.XPGained += xp

	err = savePlayer(p)
	if err != nil {
		return false, err
	}

	// Сообщаем, готов ли игрок к повышению уровня
	return p.XP >= p.NextLevelXP, nil
}

// LevelUpPlayer повышает уровень игрока и добавляет выбранный перк.
func LevelUpPlayer(chosenPerkName string) error {
	p, err := LoadPlayer()
	if err != nil {
		return err
	}

	if p.XP < p.NextLevelXP {
		// На всякий случай, если функция будет вызвана по ошибке
		return nil
	}

	p.Level++
	p.XP -= p.NextLevelXP
	p.NextLevelXP = calculateNextLevelXP(p.Level)
	if chosenPerkName != "" {
		p.Perks = append(p.Perks, chosenPerkName)
	}

	return savePlayer(p)
}

// LoadPlayer загружает данные игрока из файла.
func LoadPlayer() (*Player, error) {
	if _, err := os.Stat(playerFile); os.IsNotExist(err) {
		p := &Player{
			Name:        "Magus",
			Level:       1,
			XP:          0,
			NextLevelXP: 100,
		}
		return p, savePlayer(p)
	}

	file, err := ioutil.ReadFile(playerFile)
	if err != nil {
		return nil, err
	}

	var p Player
	if err := json.Unmarshal(file, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// savePlayer сохраняет данные игрока в файл.
func savePlayer(p *Player) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	// Убедимся, что директория data существует
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		os.Mkdir("data", 0755)
	}
	return ioutil.WriteFile(playerFile, data, 0644)
}

// calculateNextLevelXP определяет, сколько опыта нужно для следующего уровня.
func calculateNextLevelXP(level int) int {
	return 100 * level * level
}
