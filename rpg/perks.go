package rpg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"magus/player"
)

// Perk определяет структуру перка
type Perk struct {
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	UnlockLevel   int                `json:"unlock_level"`
	RequiredClass player.PlayerClass `json:"required_class,omitempty"`
}

// LoadAllPerks загружает все перки из JSON-файла.
func LoadAllPerks() ([]Perk, error) {
	file, err := ioutil.ReadFile("data/perks.json")
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать perks.json: %w", err)
	}

	var perks []Perk
	if err := json.Unmarshal(file, &perks); err != nil {
		return nil, fmt.Errorf("ошибка парсинга perks.json: %w", err)
	}

	return perks, nil
}

// GetPerkChoices возвращает до 3 перков, доступных для выбора на данном уровне.
func GetPerkChoices(p *player.Player) ([]Perk, error) {
	allPerks, err := LoadAllPerks()
	if err != nil {
		return nil, err
	}

	var availablePerks []Perk
	for _, perk := range allPerks {
		// Проверяем, что перк доступен на этом уровне и еще не получен игроком
		if perk.UnlockLevel > p.Level {
			continue
		}
		if hasPerk(p, perk.Name) {
			continue
		}

		// Проверяем классовые ограничения
		if perk.RequiredClass != player.ClassNone && perk.RequiredClass != p.Class {
			continue
		}

		availablePerks = append(availablePerks, perk)
	}

	// Возвращаем до 3 перков (можно добавить случайный выбор)
	if len(availablePerks) > 3 {
		return availablePerks[:3], nil
	}

	return availablePerks, nil
}

// hasPerk проверяет, есть ли у игрока уже данный перк.
func hasPerk(p *player.Player, perkName string) bool {
	for _, ownedPerk := range p.Perks {
		if ownedPerk == perkName {
			return true
		}
	}
	return false
}
