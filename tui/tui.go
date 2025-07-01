package tui

import (
	"fmt"
	"magus/player"
	"magus/rpg"
	"magus/storage"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/progress"
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
)

type Model struct {
	state         state
	player        player.Player
	quests        []player.Quest
	perkChoices   []rpg.Perk
	cursor        int
	activeQuestID string
	statusMessage string
	progressBar   progress.Model
}

func InitialModel() Model {
	// Загружаем все данные при инициализации
	quests, _ := storage.LoadAllQuests()
	p, _ := player.LoadPlayer()

	m := Model{
		state:         stateQuests,
		player:        *p,
		quests:        quests,
		cursor:        0,
		statusMessage: "",
		progressBar:   progress.New(progress.WithWidth(40)), // Инициализация прогресс-бара
	}

	// Сразу проверяем, не ждет ли нас повышение уровня
	if m.player.XP >= m.player.NextLevelXP {
		perkChoices, _ := rpg.GetPerkChoices(&m.player)
		if len(perkChoices) > 0 {
			m.state = stateLevelUp
			m.perkChoices = perkChoices
			m.cursor = 0
		} else {
			// Если перков для выбора нет, просто повышаем уровень
			player.LevelUpPlayer("") // Пустая строка вместо перка
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
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

		// Логика в зависимости от состояния
		switch m.state {
		case stateQuests:
			return updateQuests(msg, m)
		case stateLevelUp:
			return updateLevelUp(msg, m)
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
	default:
		return "Неизвестное состояние"
	}
}

// --- Логика для состояния просмотра квестов ---

func updateQuests(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Сбрасываем активный квест, если пользователь двигается
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
		case "enter":
			quest := m.quests[m.cursor]

			// Если квест уже выполнен, ничего не делаем
			if (quest.Type == player.Daily && isToday(quest.CompletedAt)) || quest.Completed {
				return m, nil
			}

			// Если этот квест уже активен, выполняем его
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
					if canLevelUp {
						m.state = stateLevelUp
						m.perkChoices, _ = rpg.GetPerkChoices(&m.player)
						m.cursor = 0
					}
				}

				p, _ := player.LoadPlayer()
				m.player = *p
				m.statusMessage = fmt.Sprintf("Квест '%s' выполнен! +%d XP", quest.Title, xpGained)
				m.activeQuestID = "" // Сбрасываем активный квест
			} else {
				// Иначе делаем его активным
				m.activeQuestID = quest.ID
				m.statusMessage = "Нажмите Enter еще раз для выполнения."
			}
			return m, nil
		}
	}
	return m, nil
}

func viewQuests(m Model) string {
	// Стили для Lipgloss
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
	playerInfo += fmt.Sprintf("🔋 XP: %d / %d\n", m.player.XP, m.player.NextLevelXP)

	// Прогресс-бар
	progress := float64(m.player.XP) / float64(m.player.NextLevelXP)
	playerInfo += m.progressBar.ViewAs(progress)

	// Перки
	if len(m.player.Perks) > 0 {
		playerInfo += "\n🎁 Перки: " + strings.Join(m.player.Perks, ", ")
	}

	// Сортируем квесты: невыполненные вверху, выполненные внизу
	sort.SliceStable(m.quests, func(i, j int) bool {
		isCompletedI := (m.quests[i].Type == player.Daily && isToday(m.quests[i].CompletedAt)) || m.quests[i].Completed
		isCompletedJ := (m.quests[j].Type == player.Daily && isToday(m.quests[j].CompletedAt)) || m.quests[j].Completed
		return !isCompletedI && isCompletedJ
	})

	// Список квестов
	questList := "📜 Список квестов:\n\n"
	for i, quest := range m.quests {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		// Иконки и стили
		icon := "⏳"
		style := lipgloss.NewStyle()
		isCompleted := (quest.Type == player.Daily && isToday(quest.CompletedAt)) || quest.Completed

		if isCompleted {
			icon = "✅"
			style = style.Strikethrough(true).Faint(true)
		} else if m.activeQuestID == quest.ID {
			icon = "⌛️"
			style = style.Bold(true)
		}

		questList += style.Render(fmt.Sprintf("%s %s [%s] %s {id: %s}", cursor, icon, quest.Type, quest.Title, quest.ID)) + "\n"
	}

	// Собираем все части
	return lipgloss.JoinVertical(lipgloss.Left,
		playerInfoStyle.Render(playerInfo),
		questListStyle.Render(questList),
		statusMessageStyle.Render(m.statusMessage),
		"\nНажмите 'q' для выхода.\n",
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
			// Повышаем уровень с выбранным перком
			chosenPerk := m.perkChoices[m.cursor]
			player.LevelUpPlayer(chosenPerk.Name)

			// Возвращаемся в состояние просмотра квестов
			m.state = stateQuests
			p, _ := player.LoadPlayer()
			m.player = *p // Обновляем данные игрока
			m.cursor = 0
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
