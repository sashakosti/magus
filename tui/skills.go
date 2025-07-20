package tui

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"magus/player"
	"magus/rpg"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// unlockSkill пытается разблокировать навык для игрока.
func (m *Model) unlockSkill(skillID string) error {
	if m.player.SkillPoints <= 0 {
		return fmt.Errorf("недостаточно очков навыков")
	}

	skillTree, err := rpg.LoadSkillTree(&m.player)
	if err != nil {
		return fmt.Errorf("не удалось загрузить дерево навыков: %w", err)
	}

	skillNode, ok := skillTree[skillID]
	if !ok {
		return fmt.Errorf("навык с ID '%s' не найден", skillID)
	}

	if !rpg.IsSkillAvailable(&m.player, skillNode) {
		return fmt.Errorf("требования для изучения навыка '%s' не выполнены", skillNode.Name)
	}

	// Добавляем навык и списываем очко
	m.player.UnlockedSkills = append(m.player.UnlockedSkills, skillID)
	m.player.SkillPoints--

	// Сохраняем изменения
	return player.SavePlayer(&m.player)
}

func (m *Model) updateSkills(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Загружаем дерево навыков один раз при инициализации
	if m.skillTree == nil {
		tree, err := rpg.LoadSkillTree(&m.player)
		if err != nil {
			m.statusMessage = "Ошибка загрузки дерева навыков."
			return m, nil
		}
		m.skillTree = tree

		// Конвертируем map в slice для навигации
		m.skillList = make([]player.SkillNode, 0, len(m.skillTree))
		for _, node := range m.skillTree {
			m.skillList = append(m.skillList, node)
		}
		// Сортируем по Y, затем по X для предсказуемого порядка
		sort.Slice(m.skillList, func(i, j int) bool {
			if m.skillList[i].Position.Y == m.skillList[j].Position.Y {
				return m.skillList[i].Position.X < m.skillList[j].Position.X
			}
			return m.skillList[i].Position.Y < m.skillList[j].Position.Y
		})
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.skillList)-1 {
				m.cursor++
			}
		case "enter":
			if m.cursor >= 0 && m.cursor < len(m.skillList) {
				selectedSkill := m.skillList[m.cursor]
				err := m.unlockSkill(selectedSkill.ID)
				if err != nil {
					m.statusMessage = fmt.Sprintf("❗ %v", err)
				} else {
					m.statusMessage = fmt.Sprintf("✨ Навык '%s' разблокирован!", selectedSkill.Name)
				}
			}
		}
	}
	return m, nil
}

func (m *Model) viewSkills() string {
	if m.skillTree == nil {
		return "Загрузка дерева навыков..."
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("🧠 Дерево навыков (Очки: %d)", m.player.SkillPoints)) + "\n\n")

	for i, node := range m.skillList {
		isUnlocked := rpg.IsSkillUnlocked(&m.player, node.ID)
		isAvailable := rpg.IsSkillAvailable(&m.player, node)
		isSelected := m.cursor == i

		// --- Styling ---
		style := questCardStyle.Copy().PaddingLeft(2)
		switch {
		case isSelected:
			style = selectedQuestCardStyle.Copy().PaddingLeft(2)
		case isUnlocked:
			style = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				Foreground(lipgloss.Color("250")).
				PaddingLeft(2)
		case !isAvailable:
			style = faintQuestCardStyle.Copy().PaddingLeft(2)
		}

		// --- Content ---
		var content strings.Builder
		icon := "  "
		if isUnlocked {
			icon = "✅"
		} else if isAvailable && m.player.SkillPoints > 0 {
			icon = "✨"
		}

		titleLine := fmt.Sprintf("%s %s", node.Name, icon)
		content.WriteString(titleLine + "\n")
		content.WriteString(lipgloss.NewStyle().Faint(true).Render(node.Description))

		// --- Requirements ---
		if !isUnlocked && len(node.Requirements) > 0 {
			reqsStr := m.buildRequirementsString(node)
			content.WriteString("\n" + reqsStr)
		}

		b.WriteString(style.Render(content.String()) + "\n\n")
	}

	b.WriteString("\n" + statusMessageStyle.Render(m.statusMessage) + "\n")
	return docStyle.Render(b.String())
}

// buildRequirementsString создает строку с требованиями для навыка.
func (m *Model) buildRequirementsString(node player.SkillNode) string {
	var reqs []string
	
	for _, reqID := range node.Requirements {
		// Обработка требований по уровню
		if strings.HasPrefix(reqID, "level_") {
			level := strings.TrimPrefix(reqID, "level_")
			reqStr := fmt.Sprintf("Уровень %s", level)
			
			// Проверка, выполнены ли требования по уровню
			reqLevel, _ := strconv.Atoi(level)
			if m.player.Level >= reqLevel {
				reqs = append(reqs, lipgloss.NewStyle().Strikethrough(true).Render(reqStr))
			} else {
				reqs = append(reqs, reqStr)
			}
			continue
		}

		// Обработка требований по другим навыкам
		if reqNode, ok := m.skillTree[reqID]; ok {
			reqStr := reqNode.Name
			if rpg.IsSkillUnlocked(&m.player, reqID) {
				reqs = append(reqs, lipgloss.NewStyle().Strikethrough(true).Render(reqStr))
			} else {
				reqs = append(reqs, reqStr)
			}
		}
	}
	
	if len(reqs) > 0 {
		return "Требует: " + strings.Join(reqs, ", ")
	}
	return ""
}