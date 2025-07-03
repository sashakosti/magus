package tui

import (
	"fmt"
	"magus/player"
	"magus/rpg"
	"magus/storage"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// isToday проверяет, является ли дата сегодняшней.
func isToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}

type state int

const (
	stateQuests state = iota
	stateLevelUp
	stateSkills
	stateClassChoice
)

type Model struct {
	state         state
	player        player.Player
	quests        []player.Quest
	perkChoices   []rpg.Perk
	skills        []rpg.Skill
	classChoices  []rpg.Class // Добавляем выбор класса
	cursor        int
	activeQuestID string
	statusMessage string
	progressBar   progress.Model
}

func InitialModel() Model {
	// Загружаем все данные при инициализации
	quests, _ := storage.LoadAllQuests()
	p, _ := player.LoadPlayer()
	skills, _ := rpg.LoadAllSkills()

	m := Model{
		state:         stateQuests,
		player:        *p,
		quests:        quests,
		skills:        skills,
		cursor:        0,
		statusMessage: "",
		progressBar:   progress.New(progress.WithWidth(40)),
	}

	// Проверяем, не пора ли выбрать класс
	if p.Level >= 3 && p.Class == player.ClassNone {
		m.state = stateClassChoice
		m.classChoices = rpg.GetAvailableClasses()
		m.cursor = 0
		return m
	}

	// Проверяем, не ждет ли нас повышение уровня
	if m.player.XP >= m.player.NextLevelXP {
		perkChoices, _ := rpg.GetPerkChoices(&m.player)
		if len(perkChoices) > 0 {
			m.state = stateLevelUp
			m.perkChoices = perkChoices
			m.cursor = 0
		} else {
			player.LevelUpPlayer("")
			p, _ := player.LoadPlayer()
			m.player = *p
			m.statusMessage = "Новый уровень! Доступных перков пока нет."
		}
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Глобальные хоткеи
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		// Выход из любого состояния в главное меню по 'q'
		if msg.String() == "q" {
			if m.state != stateQuests {
				m.state = stateQuests
				m.cursor = 0
				m.statusMessage = ""
				return m, nil
			}
			return m, tea.Quit
		}

		// Логика в зависимости от состояния
		switch m.state {
		case stateQuests:
			return updateQuests(msg, m)
		case stateLevelUp:
			return updateLevelUp(msg, m)
		case stateSkills:
			return updateSkills(msg, m)
		case stateClassChoice:
			return updateClassChoice(msg, m)
		}
	}
	return m, nil
}

func (m Model) View() string {
	switch m.state {
	case stateQuests:
		return viewQuests(m)
	case stateLevelUp:
		return viewLevelUp(m)
	case stateSkills:
		return viewSkills(m)
	case stateClassChoice:
		return viewClassChoice(m)
	default:
		return "Неизвестное состояние"
	}
}

// --- Логика для состояния просмотра квестов ---

func updateQuests(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "up" || msg.String() == "k" || msg.String() == "down" || msg.String() == "j" {
			m.activeQuestID = ""
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.quests)-1 {
				m.cursor++
			}
		case "s": // Переход к навыкам
			m.state = stateSkills
			m.cursor = 0
			m.statusMessage = "Распределите очки навыков."
			return m, nil
		case "enter":
			quest := m.quests[m.cursor]
			if (quest.Type == player.Daily && isToday(quest.CompletedAt)) || quest.Completed {
				return m, nil
			}
			if m.activeQuestID == quest.ID {
				var xpGained int
				if quest.Type == player.Daily {
					quest.CompletedAt = time.Now()
				} else {
					quest.Completed = true
				}
				xpGained = quest.XP
				m.quests[m.cursor] = quest
				storage.SaveAllQuests(m.quests)
				if xpGained > 0 {
					canLevelUp, _ := player.AddXP(xpGained)
					p, _ := player.LoadPlayer() // Загружаем обновленного игрока
					m.player = *p
					if canLevelUp {
						m.state = stateLevelUp
						m.perkChoices, _ = rpg.GetPerkChoices(&m.player)
						m.cursor = 0
					}
				}
				p, _ := player.LoadPlayer()
				m.player = *p
				m.statusMessage = fmt.Sprintf("Квест '%s' выполнен! +%d XP", quest.Title, xpGained)
				m.activeQuestID = ""
			} else {
				m.activeQuestID = quest.ID
				m.statusMessage = "Нажмите Enter еще раз для выполнения."
			}
			return m, nil
		}
	}
	return m, nil
}

// viewQuests отображает квесты с иерархией.
func viewQuests(m Model) string {
	playerInfoStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Width(40)
	questListStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Width(40)
	statusMessageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		PaddingLeft(2)

	// Информация об игроке
	playerInfo := fmt.Sprintf("🧙 %s (Уровень: %d)\n", m.player.Name, m.player.Level)
	if m.player.Class != player.ClassNone {
		playerInfo += fmt.Sprintf("🎖️ Класс: %s\n", m.player.Class)
	}
	playerInfo += fmt.Sprintf("🔋 XP: %d / %d\n", m.player.XP, m.player.NextLevelXP)
	progress := float64(m.player.XP) / float64(m.player.NextLevelXP)
	playerInfo += m.progressBar.ViewAs(progress)
	if len(m.player.Perks) > 0 {
		playerInfo += "\n🎁 Перки: " + strings.Join(m.player.Perks, ", ")
	}
	if m.player.SkillPoints > 0 {
		playerInfo += fmt.Sprintf("\n✨ Очки навыков: %d", m.player.SkillPoints)
	}

	// Группируем подзадачи
	subQuests := make(map[string][]player.Quest)
	for _, q := range m.quests {
		if q.ParentID != "" {
			subQuests[q.ParentID] = append(subQuests[q.ParentID], q)
		}
	}

	// Формируем плоский список для отображения с учетом иерархии
	displayQuests := []player.Quest{}
	for _, q := range m.quests {
		if q.ParentID != "" {
			continue
		}
		displayQuests = append(displayQuests, q)
		if children, ok := subQuests[q.ID]; ok {
			displayQuests = append(displayQuests, children...)
		}
	}
	m.quests = displayQuests // Обновляем порядок в модели для корректной работы курсора

	// Рендерим список
	questList := "📜 Список квестов:\n\n"
	for i, quest := range m.quests {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		icon := "⏳"
		style := lipgloss.NewStyle()
		isCompleted := quest.Completed || (quest.Type == player.Daily && isToday(quest.CompletedAt))

		if isCompleted {
			icon = "✅"
			style = style.Strikethrough(true).Faint(true)
		} else if m.activeQuestID == quest.ID {
			icon = "⌛️"
			style = style.Bold(true)
		}

		indent := ""
		if quest.ParentID != "" {
			indent = "  └─ "
		}

		questList += style.Render(fmt.Sprintf("%s %s%s [%s] %s {id: %s}", cursor, indent, icon, quest.Type, quest.Title, quest.ID)) + "\n"
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		playerInfoStyle.Render(playerInfo),
		questListStyle.Render(questList),
		statusMessageStyle.Render(m.statusMessage),
		"\nНажмите 's' для навыков, 'q' для выхода.\n",
	)
}

// --- Логика для состояния повышения уровня ---

func updateLevelUp(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.perkChoices)-1 {
				m.cursor++
			}
		case "enter":
			chosenPerk := m.perkChoices[m.cursor]
			player.LevelUpPlayer(chosenPerk.Name)
			m.state = stateQuests
			p, _ := player.LoadPlayer()
			m.player = *p
			m.cursor = 0
			m.statusMessage = fmt.Sprintf("Вы выучили перк: %s! И получили 10 очков навыков.", chosenPerk.Name)
			return m, nil
		}
	}
	return m, nil
}

func viewLevelUp(m Model) string {
	s := "🔥 Поздравляем! Новый уровень!\n\n"
	s += "Выберите новый перк:\n\n"
	for i, perk := range m.perkChoices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s: %s\n", cursor, perk.Name, perk.Description)
	}
	s += "\nНажмите 'enter' для выбора.\n"
	return s
}

// --- Логика для состояния навыков ---

func updateSkills(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.skills)-1 {
				m.cursor++
			}
		case "enter":
			if m.player.SkillPoints > 0 {
				skillToIncrease := m.skills[m.cursor]
				err := rpg.IncreaseSkill(&m.player, skillToIncrease.Name)
				if err != nil {
					m.statusMessage = fmt.Sprintf("Ошибка: %v", err)
				} else {
					// Обновляем данные игрока в модели после успешного сохранения
					p, _ := player.LoadPlayer()
					m.player = *p
					m.statusMessage = fmt.Sprintf("Навык '%s' увеличен!", skillToIncrease.Name)
				}
			} else {
				m.statusMessage = "Недостаточно очков навыков."
			}
			return m, nil
		}
	}
	return m, nil
}

func viewSkills(m Model) string {
	s := fmt.Sprintf("🧠 Навыки (Очки: %d)\n\n", m.player.SkillPoints)
	for i, skill := range m.skills {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		level := m.player.Skills[skill.Name]
		s += fmt.Sprintf("%s %s: %d\n  %s\n\n", cursor, skill.Name, level, skill.Description)
	}
	s += "\nНажмите 'enter' для улучшения, 'q' для возврата.\n"
	return s
}

// --- Логика для состояния выбора класса ---

func updateClassChoice(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.classChoices)-1 {
				m.cursor++
			}
		case "enter":
			chosenClass := m.classChoices[m.cursor]
			m.player.Class = chosenClass.Name
			player.SavePlayer(&m.player) // Сохраняем выбор

			m.state = stateQuests
			m.cursor = 0
			m.statusMessage = fmt.Sprintf("Вы выбрали класс: %s!", chosenClass.Name)
			return m, nil
		}
	}
	return m, nil
}

func viewClassChoice(m Model) string {
	s := "⚔️ Пришло время выбрать свой путь!\n\n"
	s += "Выберите класс:\n\n"
	for i, class := range m.classChoices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s: %s\n", cursor, class.Name, class.Description)
	}
	s += "\nНажмите 'enter' для выбора. Этот выбор нельзя будет изменить.\n"
	return s
}
