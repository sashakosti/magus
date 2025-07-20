package rpg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"magus/player"
	"strconv"
	"strings"
)

// loadSkillsFromFile загружает и декодирует навыки из указанного файла.
func loadSkillsFromFile(filePath string) ([]player.SkillNode, error) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл %s: %w", filePath, err)
	}

	var nodes []player.SkillNode
	if err := json.Unmarshal(file, &nodes); err != nil {
		return nil, fmt.Errorf("ошибка парсинга %s: %w", filePath, err)
	}
	return nodes, nil
}

// LoadSkillTree загружает дерево навыков для конкретного игрока.
func LoadSkillTree(p *player.Player) (map[string]player.SkillNode, error) {
	skillTree := make(map[string]player.SkillNode)

	// 1. Загружаем общие навыки
	commonSkills, err := loadSkillsFromFile("data/common_skills.json")
	if err != nil {
		return nil, err
	}
	for _, node := range commonSkills {
		skillTree[node.ID] = node
	}

	// 2. Загружаем классовые навыки
	var classSkillFile string
	switch p.Class {
	case player.ClassMage:
		classSkillFile = "data/mage_skills.json"
	case player.ClassWarrior:
		classSkillFile = "data/warrior_skills.json"
	case player.ClassRogue:
		classSkillFile = "data/rogue_skills.json"
	default:
		// Если класс не выбран или неизвестен, загружаем только общие навыки
		return skillTree, nil
	}

	classSkills, err := loadSkillsFromFile(classSkillFile)
	if err != nil {
		return nil, err
	}
	for _, node := range classSkills {
		// Проверяем на случай дублирования ID
		if _, exists := skillTree[node.ID]; exists {
			fmt.Printf("Внимание: Дубликат ID навыка '%s' в файле %s\n", node.ID, classSkillFile)
		}
		skillTree[node.ID] = node
	}

	return skillTree, nil
}

// IsSkillUnlocked проверяет, разблокирован ли у игрока данный перк.
func IsSkillUnlocked(p *player.Player, skillID string) bool {
	for _, unlockedSkill := range p.UnlockedSkills {
		if unlockedSkill == skillID {
			return true
		}
	}
	return false
}

// IsSkillAvailable проверяет, доступен ли перк для изучения.
func IsSkillAvailable(p *player.Player, node player.SkillNode) bool {
	if IsSkillUnlocked(p, node.ID) {
		return false // Уже разблокирован
	}

	for _, reqID := range node.Requirements {
		// Проверка требований по уровню
		if strings.HasPrefix(reqID, "level_") {
			levelStr := strings.TrimPrefix(reqID, "level_")
			requiredLevel, err := strconv.Atoi(levelStr)
			if err != nil {
				// Обработка ошибки, если формат требования уровня некорректен
                fmt.Printf("Ошибка парсинга требования к уровню: %s\n", reqID)
                return false
            }
			if p.Level < requiredLevel {
				return false
			}
			continue // Переходим к следующему требованию
		}

		// Проверка других перков
		if !IsSkillUnlocked(p, reqID) {
			return false
		}
	}
	return true
}
