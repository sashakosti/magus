package rpg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"magus/player"
	"strconv"
	"strings"
)

// SkillTrees содержит разделенные деревья навыков.
type SkillTrees struct {
	Common map[string]player.SkillNode
	Class  map[string]player.SkillNode
}

// LoadSkillTrees загружает и разделяет навыки на общие и классовые из skill_tree.json.
func LoadSkillTrees(p *player.Player) (SkillTrees, error) {
	trees := SkillTrees{
		Common: make(map[string]player.SkillNode),
		Class:  make(map[string]player.SkillNode),
	}

	file, err := ioutil.ReadFile("data/skill_tree.json")
	if err != nil {
		return trees, fmt.Errorf("не удалось прочитать файл data/skill_tree.json: %w", err)
	}

	var allNodes []player.SkillNode
	if err := json.Unmarshal(file, &allNodes); err != nil {
		return trees, fmt.Errorf("ошибка парсинга data/skill_tree.json: %w", err)
	}

	for _, node := range allNodes {
		// Убираем поле ClassRequirement из JSON, если оно пустое, для обратной совместимости
		if node.ClassRequirement == "" {
			trees.Common[node.ID] = node
		} else if node.ClassRequirement == string(p.Class) {
			trees.Class[node.ID] = node
		}
	}

	return trees, nil
}

// IsSkillUnlocked проверяет, разблокирован ли у игрока данный навык.
func IsSkillUnlocked(p *player.Player, skillID string) bool {
	for _, unlockedSkill := range p.UnlockedSkills {
		if unlockedSkill == skillID {
			return true
		}
	}
	return false
}

// IsSkillAvailable проверяет, доступен ли навык для изучения.
// skillTree - это объединенная карта общих и классовых навыков.
func IsSkillAvailable(p *player.Player, node player.SkillNode, skillTree map[string]player.SkillNode) bool {
	if IsSkillUnlocked(p, node.ID) {
		return false // Уже разблокирован
	}

	// Проверка на соответствие классу
	if node.ClassRequirement != "" && node.ClassRequirement != string(p.Class) {
		return false
	}

	for _, reqID := range node.Requirements {
		// Проверка требований по уровню
		if strings.HasPrefix(reqID, "level_") {
			levelStr := strings.TrimPrefix(reqID, "level_")
			requiredLevel, err := strconv.Atoi(levelStr)
			if err != nil {
				fmt.Printf("Ошибка парсинга требования к уровню: %s\n", reqID)
				return false
			}
			if p.Level < requiredLevel {
				return false
			}
			continue // Переходим к следующему требованию
		}

		// Проверка других навыков
		if !IsSkillUnlocked(p, reqID) {
			return false
		}
	}
	return true
}
