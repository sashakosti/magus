package player

import (
	"encoding/json"
	"errors"
	"os" // lol
	"time"
)

var PlayerFile = "data/player.json"

var ErrPlayerNotFound = errors.New("player file not found")

// AddXP добавляет опыт игроку и возвращает true, если можно повысить уровень.
func AddXP(xp int) (bool, error) {
	p, err := LoadPlayer()
	if err != nil {
		return false, err
	}

	p.XP += xp
	p.History.QuestsCompleted++
	p.History.XPGained += xp

	err = SavePlayer(p)
	if err != nil {
		return false, err
	}

	// Сообщаем, готов ли игрок к повышению уровня
	return p.XP >= p.NextLevelXP, nil
}

// LevelUpPlayer повышает уровень игрока, добавляет выбранный перк и начисляет очки навыков.
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
	p.SkillPoints += 10 // Начисляем 10 очков навыков за уровень

	return SavePlayer(p)
}

// CreatePlayer создает нового игрока с заданным именем.
func CreatePlayer(name string) (*Player, error) {
	if name == "" {
		name = "Magus" // Имя по умолчанию
	}
	p := &Player{
		Name:        name,
		Level:       1,
		HP:          100,
		MaxHP:       100,
		XP:          0,
		NextLevelXP: 100,
		Skills:      make(map[string]int),
		History: History{
			QuestsCompleted: 0,
			XPGained:        0,
		},
	}
	return p, SavePlayer(p)
}

// LoadPlayer загружает данные игрока из файла.
func LoadPlayer() (*Player, error) {
	if _, err := os.Stat(PlayerFile); os.IsNotExist(err) {
		return nil, ErrPlayerNotFound
	}

	file, err := os.ReadFile(PlayerFile)
	if err != nil {
		return nil, err
	}

	var p Player
	if err := json.Unmarshal(file, &p); err != nil {
		return nil, err
	}

	// Для обратной совместимости: если у старого игрока нет карты навыков, создаем ее
	if p.Skills == nil {
		p.Skills = make(map[string]int)
	}

	// ��ля обратной совместимости: если у старого игрока нет HP, устанавливаем его
	if p.MaxHP == 0 {
		p.MaxHP = 100
		p.HP = 100
	}

	return &p, nil
}

// SavePlayer сохраняет данные игрока в файл.
func SavePlayer(p *Player) error {
	p.LastSeen = time.Now()
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	// Убедимся, что директория data существует
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		os.Mkdir("data", 0755)
	}
	return os.WriteFile(PlayerFile, data, 0644)
}

// calculateNextLevelXP определяет, сколько опыта нужно для следующего уровня.
func calculateNextLevelXP(level int) int {
	return 100 * level * level
}

