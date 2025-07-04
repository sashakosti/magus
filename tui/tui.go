package tui

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"magus/player"
	"magus/rpg"
	"magus/storage"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// --- HELPERS ---

func isToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}

func deadlineStatus(deadline *time.Time) string {
	if deadline == nil {
		return ""
	}
	remaining := time.Until(*deadline)
	if remaining < 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("(Просрочено)")
	}
	days := int(remaining.Hours() / 24)
	return fmt.Sprintf("(осталось %d д)", days)
}

// --- STATE ---

type state int

const (
	stateHomepage state = iota
	stateQuests
	stateCompletedQuests
	stateAddQuest
	stateLevelUp
	stateSkills
	stateClassChoice
	stateCreatePlayer
)

// --- MODEL ---

type Model struct {
	state             state
	player            player.Player
	quests            []player.Quest
	displayQuests     []player.Quest
	perkChoices       []rpg.Perk
	skills            []rpg.Skill
	classChoices      []rpg.Class
	cursor            int
	activeQuestID     string
	statusMessage     string
	progressBar       progress.Model
	collapsed         map[string]bool
	homepageCursor    int
	addQuestInputs    []textinput.Model
	addQuestCursor    int
	addQuestTypes     []player.QuestType
	addQuestTypeIdx   int
	createPlayerInput textinput.Model
}

func newCreatePlayerInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Имя твоего героя"
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 50
	return ti
}

func newAddQuestInputs() []textinput.Model {
	inputs := make([]textinput.Model, 5) // Title, XP, Tags, Deadline, Parent
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].CharLimit = 120
	}
	inputs[0].Placeholder = "Название квеста"
	inputs[1].Placeholder = "10"
	inputs[2].Placeholder = "работа,дом"
	inputs[3].Placeholder = "ГГГГ-ММ-ДД"
	inputs[4].Placeholder = "ID родителя (опционально)"
	inputs[0].Focus()
	return inputs
}

func InitialModel() Model {
	p, err := player.LoadPlayer()
	if err != nil {
		// Если игрок не найден, переходим в режим создания
		return Model{
			state:             stateCreatePlayer,
			createPlayerInput: newCreatePlayerInput(),
		}
	}

	quests, _ := storage.LoadAllQuests()
	skills, _ := rpg.LoadAllSkills()

	m := Model{
		state:           stateHomepage,
		player:          *p,
		quests:          quests,
		skills:          skills,
		cursor:          0,
		statusMessage:   "",
		progressBar:     progress.New(progress.WithWidth(40)),
		collapsed:       make(map[string]bool),
		homepageCursor:  0,
		addQuestTypes:   []player.QuestType{player.Daily, player.Arc, player.Meta, player.Epic, player.Chore},
		addQuestTypeIdx: 0,
	}

	m.sortAndBuildDisplayQuests()

	if p.Level >= 3 && p.Class == player.ClassNone {
		m.state = stateClassChoice
		m.classChoices = rpg.GetAvailableClasses()
		m.cursor = 0
		return m
	}

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

func (m *Model) sortAndBuildDisplayQuests() {
	sort.SliceStable(m.quests, func(i, j int) bool {
		d1 := m.quests[i].Deadline
		d2 := m.quests[j].Deadline
		if d1 != nil && d2 != nil {
			return d1.Before(*d2)
		}
		if d1 != nil && d2 == nil {
			return true
		}
		if d1 == nil && d2 != nil {
			return false
		}
		return m.quests[i].CreatedAt.After(m.quests[j].CreatedAt)
	})

	activeQuests := []player.Quest{}
	for _, q := range m.quests {
		if !q.Completed || q.Type == player.Daily {
			activeQuests = append(activeQuests, q)
		}
	}

	subQuests := make(map[string][]player.Quest)
	for _, q := range activeQuests {
		if q.ParentID != "" {
			subQuests[q.ParentID] = append(subQuests[q.ParentID], q)
		}
	}

	displayQuests := []player.Quest{}
	for _, q := range activeQuests {
		if q.ParentID != "" {
			continue
		}
		displayQuests = append(displayQuests, q)
		if children, ok := subQuests[q.ID]; ok {
			if !m.collapsed[q.ID] {
				displayQuests = append(displayQuests, children...)
			}
		}
	}
	m.displayQuests = displayQuests
}

// --- INIT & UPDATE & VIEW ---

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" {
			return m, tea.Quit
		}
		if key == "q" {
			if m.state == stateHomepage || m.state == stateCreatePlayer {
				return m, tea.Quit
			}
			if m.state == stateAddQuest {
				m.addQuestInputs = nil
			}
			m.state = stateHomepage
			m.statusMessage = ""
			m.cursor = 0
			return m, nil
		}
		if key == "a" && m.state != stateAddQuest && m.state != stateLevelUp && m.state != stateClassChoice && m.state != stateCreatePlayer {
			m.state = stateAddQuest
			m.addQuestCursor = 0
			m.addQuestTypeIdx = 0
			m.addQuestInputs = newAddQuestInputs()
			return m, nil
		}
	}

	var cmd tea.Cmd
	switch m.state {
	case stateHomepage:
		return updateHomepage(msg, m)
	case stateQuests:
		return updateQuests(msg, m)
	case stateCompletedQuests:
		return updateCompletedQuests(msg, m)
	case stateAddQuest:
		updatedModel, newCmd := updateAddQuest(msg, m)
		m = updatedModel.(Model)
		cmd = newCmd
		return m, cmd
	case stateLevelUp:
		return updateLevelUp(msg, m)
	case stateSkills:
		return updateSkills(msg, m)
	case stateClassChoice:
		return updateClassChoice(msg, m)
	case stateCreatePlayer:
		return updateCreatePlayer(msg, m)
	}

	return m, nil
}

func (m Model) View() string {
	switch m.state {
	case stateHomepage:
		return viewHomepage(m)
	case stateQuests:
		return viewQuests(m)
	case stateCompletedQuests:
		return viewCompletedQuests(m)
	case stateAddQuest:
		return viewAddQuest(m)
	case stateLevelUp:
		return viewLevelUp(m)
	case stateSkills:
		return viewSkills(m)
	case stateClassChoice:
		return viewClassChoice(m)
	case stateCreatePlayer:
		return viewCreatePlayer(m)
	default:
		return "Неизвестное состояние"
	}
}

// --- CREATE PLAYER ---

func updateCreatePlayer(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			playerName := m.createPlayerInput.Value()
			if playerName == "" {
				return m, nil // Не позволяем создать игрока с пустым именем
			}
			_, err := player.CreatePlayer(playerName)
			if err != nil {
				m.statusMessage = fmt.Sprintf("Ошибка создания игрока: %v", err)
				return m, nil
			}
			// Перезагружаем модель с новым игроком
			return InitialModel(), nil
		}
	}
	m.createPlayerInput, cmd = m.createPlayerInput.Update(msg)
	return m, cmd
}

func viewCreatePlayer(m Model) string {
	return fmt.Sprintf(
		"Добро пожаловать в Magus!\n\nДавай создадим твоего персонажа.\n\n%s\n\nНажми Enter, чтобы начать свое приключение.",
		m.createPlayerInput.View(),
	)
}

// --- HOMEPAGE ---

func updateHomepage(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.homepageCursor > 0 {
				m.homepageCursor--
			}
		case "down", "j":
			if m.homepageCursor < 3 {
				m.homepageCursor++
			}
		case "enter":
			switch m.homepageCursor {
			case 0:
				m.state = stateQuests
				m.sortAndBuildDisplayQuests()
			case 1:
				m.state = stateCompletedQuests
			case 2:
				m.state = stateSkills
			case 3:
				return m, tea.Quit
			}
			m.cursor = 0
			m.statusMessage = ""
		}
	}
	return m, nil
}

func viewHomepage(m Model) string {
	// --- Player Info Box ---
	var playerInfoLines []string
	playerInfoLines = append(playerInfoLines, fmt.Sprintf("🧙 %s (Уровень: %d)", m.player.Name, m.player.Level))
	if m.player.Class != player.ClassNone {
		playerInfoLines = append(playerInfoLines, fmt.Sprintf("🎖️  Класс: %s", m.player.Class))
	}
	playerInfoLines = append(playerInfoLines, fmt.Sprintf("🔋 XP: %d / %d", m.player.XP, m.player.NextLevelXP))
	playerInfoLines = append(playerInfoLines, m.progressBar.ViewAs(float64(m.player.XP)/float64(m.player.NextLevelXP)))
	if len(m.player.Perks) > 0 {
		playerInfoLines = append(playerInfoLines, "🎁 Перки: "+strings.Join(m.player.Perks, ", "))
	}
	if m.player.SkillPoints > 0 {
		playerInfoLines = append(playerInfoLines, fmt.Sprintf("✨ Очки навыков: %d", m.player.SkillPoints))
	}
	playerInfoContent := strings.Join(playerInfoLines, "\n")
	playerInfoBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Render(playerInfoContent)

	// --- Menu Box ---
	var menuLines []string
	menuLines = append(menuLines, "Меню")
	menuLines = append(menuLines, "") // Spacer
	menuItems := []string{"Активные квесты", "Завершенные квесты", "Навыки", "Выход"}
	for i, item := range menuItems {
		cursor := " "
		if m.homepageCursor == i {
			cursor = ">"
		}
		menuLines = append(menuLines, fmt.Sprintf("%s %s", cursor, item))
	}
	menuContent := strings.Join(menuLines, "\n")
	menuBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Render(menuContent)

	// --- Final Assembly using lipgloss.Place ---
	// This approach is more robust for aligning independently sized blocks.
	ui := lipgloss.JoinHorizontal(lipgloss.Top, playerInfoBox, menuBox)

	art := ` `
	artBox := lipgloss.NewStyle().
		Align(lipgloss.Center).
		PaddingTop(1).
		Render(art)

	// Place the UI centrally.
	// You might need to get terminal width for more advanced centering.
	// For now, this structure is cleaner.
	content := lipgloss.JoinVertical(lipgloss.Left,
		artBox,
		ui,
		lipgloss.NewStyle().Padding(1, 2).Render(m.statusMessage),
		"\nНавигация: ↑/↓, Enter для выбора, 'a' для добавления, 'q' для выхода.",
	)

	return content
}

// --- QUESTS ---

func updateQuests(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.displayQuests)-1 {
				m.cursor++
			}
		case " ":
			if m.cursor < len(m.displayQuests) {
				quest := m.displayQuests[m.cursor]
				if quest.ParentID == "" {
					m.collapsed[quest.ID] = !m.collapsed[quest.ID]
					m.sortAndBuildDisplayQuests()
				}
			}
		case "enter":
			if m.cursor >= len(m.displayQuests) {
				return m, nil
			}
			quest := m.displayQuests[m.cursor]
			if (quest.Type == player.Daily && isToday(quest.CompletedAt)) || quest.Completed {
				return m, nil
			}

			var xpGained int
			for i, q := range m.quests {
				if q.ID == quest.ID {
					if m.quests[i].Type == player.Daily {
						m.quests[i].CompletedAt = time.Now()
					} else {
						m.quests[i].Completed = true
					}
					xpGained = m.quests[i].XP
					break
				}
			}

			storage.SaveAllQuests(m.quests)
			m.sortAndBuildDisplayQuests()

			if xpGained > 0 {
				canLevelUp, _ := player.AddXP(xpGained)
				p, _ := player.LoadPlayer()
				m.player = *p
				if canLevelUp {
					perkChoices, _ := rpg.GetPerkChoices(&m.player)
					if len(perkChoices) > 0 {
						m.state = stateLevelUp
						m.perkChoices = perkChoices
					} else {
						player.LevelUpPlayer("")
						p, _ := player.LoadPlayer()
						m.player = *p
						m.statusMessage = "Новый уровень! Доступных перков пока нет."
					}
					m.cursor = 0
				}
			}
			p, _ := player.LoadPlayer()
			m.player = *p
			m.statusMessage = fmt.Sprintf("Квест '%s' выполнен! +%d XP", quest.Title, xpGained)
		}
	}
	return m, nil
}

func viewQuests(m Model) string {
	s := "📜 Активные квесты\n\n"
	for i, quest := range m.displayQuests {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		icon := "⏳"
		style := lipgloss.NewStyle()
		if quest.Completed || (quest.Type == player.Daily && isToday(quest.CompletedAt)) {
			icon = "✅"
			style = style.Strikethrough(true).Faint(true)
		}

		indent := ""
		collapseIcon := " "
		if quest.ParentID == "" {
			isParent := false
			for _, q := range m.quests {
				if q.ParentID == quest.ID {
					isParent = true
					break
				}
			}
			if isParent {
				collapseIcon = "⊖"
				if m.collapsed[quest.ID] {
					collapseIcon = "⊕"
				}
			}
		} else {
			indent = "  └─ "
		}

		tags := ""
		for _, tag := range quest.Tags {
			tags += fmt.Sprintf(" [#%s]", tag)
		}

		// Correctly handle padding with runewidth
		iconWithPadding := icon + strings.Repeat(" ", 2-runewidth.StringWidth(icon))
		collapseIconWithPadding := collapseIcon + strings.Repeat(" ", 2-runewidth.StringWidth(collapseIcon))

		s += style.Render(fmt.Sprintf("%s %s%s%s[%s] %s%s %s", cursor, indent, iconWithPadding, collapseIconWithPadding, quest.Type, quest.Title, tags, deadlineStatus(quest.Deadline))) + "\n"
	}

	s += fmt.Sprintf("\n%s\n", m.statusMessage)
	s += "\nНавигация: ↑/↓, Enter, [Пробел], 'a' для добавления, 'q' для возврата."
	return s
}

// --- COMPLETED QUESTS ---

func updateCompletedQuests(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	return m, nil
}

func viewCompletedQuests(m Model) string {
	s := "✅ Завершенные квесты\n\n"
	found := false
	for _, quest := range m.quests {
		if quest.Completed && quest.Type != player.Daily {
			s += fmt.Sprintf("  - %s [%s] (XP: %d)\n", quest.Title, quest.Type, quest.XP)
			found = true
		}
	}
	if !found {
		s += "Пока нет завершенных квестов."
	}
	s += "\nНажмите 'q' для возврата."
	return s
}

// --- ADD QUEST ---

func updateAddQuest(msg tea.Msg, m Model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			m.addQuestCursor = (m.addQuestCursor + 1) % 6 // 6 fields now
		case "shift+tab", "up":
			m.addQuestCursor--
			if m.addQuestCursor < 0 {
				m.addQuestCursor = 5
			}
		case "left":
			if m.addQuestCursor == 1 { // Type field
				m.addQuestTypeIdx--
				if m.addQuestTypeIdx < 0 {
					m.addQuestTypeIdx = len(m.addQuestTypes) - 1
				}
			}
		case "right":
			if m.addQuestCursor == 1 { // Type field
				m.addQuestTypeIdx = (m.addQuestTypeIdx + 1) % len(m.addQuestTypes)
			}
		case "enter":
			if m.addQuestCursor == 5 { // Last field, save
				title := m.addQuestInputs[0].Value()
				if title == "" {
					m.statusMessage = "Название квеста не может быть пустым."
					return m, nil
				}

				xp, _ := strconv.Atoi(m.addQuestInputs[1].Value())
				if xp == 0 {
					xp = 10
				}

				tagsStr := m.addQuestInputs[2].Value()
				var tags []string
				if tagsStr != "" {
					tags = strings.Split(tagsStr, ",")
				}

				deadlineStr := m.addQuestInputs[3].Value()
				var deadline *time.Time
				if deadlineStr != "" {
					t, err := time.Parse("2006-01-02", deadlineStr)
					if err == nil {
						deadline = &t
					}
				}

				parentID := m.addQuestInputs[4].Value()

				bytes := make([]byte, 4)
				if _, err := rand.Read(bytes); err != nil {
					panic(err)
				}
				id := hex.EncodeToString(bytes)

				newQuest := player.Quest{
					ID:        id,
					Title:     title,
					XP:        xp,
					Tags:      tags,
					Deadline:  deadline,
					ParentID:  parentID,
					Type:      m.addQuestTypes[m.addQuestTypeIdx],
					CreatedAt: time.Now(),
				}

				m.quests = append(m.quests, newQuest)
				storage.SaveAllQuests(m.quests)
				m.sortAndBuildDisplayQuests()
				m.state = stateQuests
				m.statusMessage = fmt.Sprintf("Квест '%s' добавлен!", title)
				m.addQuestInputs = nil
				return m, nil
			}
			m.addQuestCursor = (m.addQuestCursor + 1) % 6
		}

		for i := 0; i < len(m.addQuestInputs); i++ {
			if i == m.addQuestCursor-1 || (m.addQuestCursor == 0 && i == 1) { // Blur previous/next on text inputs
				m.addQuestInputs[i].Blur()
			}
		}
		if m.addQuestCursor > 1 {
			m.addQuestInputs[m.addQuestCursor-1].Focus()
		} else if m.addQuestCursor == 0 {
			m.addQuestInputs[0].Focus()
		}
	}

	for i := range m.addQuestInputs {
		m.addQuestInputs[i], cmd = m.addQuestInputs[i].Update(msg)
	}

	return m, cmd
}

func viewAddQuest(m Model) string {
	var b strings.Builder
	b.WriteString("📝 Новый квест\n\n")

	// Title
	b.WriteString("Название\n" + m.addQuestInputs[0].View() + "\n\n")

	// Type
	typeStyle := lipgloss.NewStyle()
	if m.addQuestCursor == 1 {
		typeStyle = typeStyle.Foreground(lipgloss.Color("205"))
	}
	b.WriteString("Тип\n" + typeStyle.Render(fmt.Sprintf("< %s >", m.addQuestTypes[m.addQuestTypeIdx])) + "\n\n")

	// XP
	b.WriteString("XP\n" + m.addQuestInputs[1].View() + "\n\n")
	// Tags
	b.WriteString("Теги (через запятую)\n" + m.addQuestInputs[2].View() + "\n\n")
	// Deadline
	b.WriteString("Дедлайн (ГГГГ-ММ-ДД)\n" + m.addQuestInputs[3].View() + "\n\n")
	// Parent
	b.WriteString("Родительский ID\n" + m.addQuestInputs[4].View() + "\n\n")

	b.WriteString("\nНавигация: ↑/↓, ←/→ для типа, Enter, 'q' для отмены.")
	return b.String()
}

// --- LEVEL UP, SKILLS, CLASS CHOICE ---
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
			p, _ := player.LoadPlayer()
			m.player = *p

			if m.player.Level >= 3 && m.player.Class == player.ClassNone {
				m.state = stateClassChoice
				m.classChoices = rpg.GetAvailableClasses()
				m.cursor = 0
			} else {
				m.state = stateHomepage
				m.statusMessage = fmt.Sprintf("Вы выучили перк: %s! И получили 10 очков навыков.", chosenPerk.Name)
			}
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
	return lipgloss.NewStyle().Border(lipgloss.DoubleBorder(), true).Padding(2).Render(s)
}

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
					p, _ := player.LoadPlayer()
					m.player = *p
					m.statusMessage = fmt.Sprintf("Навык '%s' увеличен!", skillToIncrease.Name)
				}
			} else {
				m.statusMessage = "Недостаточно очков навыков."
			}
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
	s += fmt.Sprintf("\n%s\n", m.statusMessage)
	s += "\nНажмите 'enter' для улучшения, 'q' для возврата.\n"
	return s
}

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
			player.SavePlayer(&m.player)

			m.state = stateHomepage
			m.statusMessage = fmt.Sprintf("Вы выбрали класс: %s!", chosenClass.Name)
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
	return lipgloss.NewStyle().Border(lipgloss.DoubleBorder(), true).Padding(2).Render(s)
}
