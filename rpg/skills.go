package rpg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"magus/player"
)

// Skill определяет структуру навыка
type Skill struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// LoadAllSkills загружает все навыки из JSON-файла.
func LoadAllSkills() ([]Skill, error) {
	file, err := ioutil.ReadFile("data/skills.json")
	if err != nil {
		return nil, fmt.Errorf("Не удалось прочитать skills.json: %w", err)
	}

	var skills []Skill
	if err := json.Unmarshal(file, &skills); err != nil {
		return nil, fmt.Errorf("Ошибка парсинга skills.json: %w", err)
	}

	return skills, nil
}

// IncreaseSkill увеличивает уровень навыка игрока и сохраняет изменения.
func IncreaseSkill(p *player.Player, skillName string) error {
	if p.SkillPoints <= 0 {
		return fmt.Errorf("недостаточно очков навыков")
	}

	// Увеличиваем навык и списыва��м очко
	p.Skills[skillName]++
	p.SkillPoints--

	// Сохраняем изменения в файле
	return player.SavePlayer(p)
}
