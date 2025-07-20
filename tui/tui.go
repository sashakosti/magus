package tui

import (
	"sort"
	"strings"
	"time"
	"magus/dungeon"
	"magus/player"
	"magus/rpg"
	"magus/storage"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
)

type dungeonState int

const (
	DungeonStateExploring dungeonState = iota
	DungeonStateInCombat
	DungeonStateFinished
)

// --- MODEL ---

type Model struct {
	state           state
	stateStack      []state
	player          player.Player
	quests          []player.Quest
	displayQuests   []player.Quest
	skillChoices    []player.SkillNode
	skillTree       map[string]player.SkillNode
	skillList       []player.SkillNode
	classChoices    []rpg.Class
	cursor          int
	activeQuestID   string
	statusMessage   string
	progressBar     progress.Model
	collapsed       map[string]bool
	homepageCursor  int
	addQuestInputs  []textinput.Model
	addQuestCursor  int
	addQuestTypes   []player.QuestType
	addQuestTypeIdx int
	createPlayerInput textinput.Model

	// Quest view
	viewport          viewport.Model
	questFilters      []string
	questFilterCursor int
	activeQuestFilter string

	// Tag management
	allTags        []string
	tagCursor      int
	renameTagInput textinput.Model

	// Dungeon state
	dungeonState             dungeonState
	dungeonFloor             int
	dungeonRunGold           int
	dungeonRunXP             int
	dungeonSelectedDuration  time.Duration
	dungeonStartTime         time.Time
	dungeonTicker            *time.Ticker
	dungeonCustomDurationInput textinput.Model
	currentMonster           *dungeon.Monster
	dungeonLog               []string

	// Quest editing
	editingQuest   player.Quest
	editInputs     []textinput.Model
	editFocusIndex int

	// Perk tree view
	perkCursorX    int
	perkCursorY    int
	cameraOffsetX  int
	cameraOffsetY  int

	terminalWidth  int
	terminalHeight int
}

func InitialModel() Model {
	p, err := player.LoadPlayer()
	if err != nil {
		return Model{
			state:             stateCreatePlayer,
			createPlayerInput: newCreatePlayerInput(),
		}
	}

	// Passive HP Regeneration
	if !p.LastSeen.IsZero() {
		minutesPassed := time.Since(p.LastSeen).Minutes()
		hpToRestore := int(minutesPassed / 5)
		if hpToRestore > 0 {
			p.HP += hpToRestore
			if p.HP > p.MaxHP {
				p.HP = p.MaxHP
			}
			player.SavePlayer(p) // Save the updated HP
		}
	}

	quests, _ := storage.LoadAllQuests()

	vp := viewport.New(100, 20) // Initial size, will be updated

	m := Model{
		state:                      stateHomepage,
		player:                     *p,
		quests:                     quests,
		cursor:                     0,
		statusMessage:              "",
		progressBar:                progress.New(progress.WithDefaultGradient(), progress.WithWidth(40), progress.WithoutPercentage()),
		collapsed:                  make(map[string]bool),
		homepageCursor:             0,
		addQuestTypes:              []player.QuestType{player.Daily, player.Arc, player.Meta, player.Epic, player.Chore},
		addQuestTypeIdx:            0,
		dungeonCustomDurationInput: textinput.New(),
		activeQuestFilter:          "Все",
		viewport:                   vp,
		renameTagInput:             textinput.New(),
	}

	m.buildQuestFilters()
	m.sortAndBuildDisplayQuests()

	if p.Level >= 3 && p.Class == player.ClassNone {
		m.state = stateClassChoice
		m.classChoices = rpg.GetAvailableClasses()
		m.cursor = 0
		return m
	}

	if m.player.XP >= m.player.NextLevelXP {
		// Загружаем все дерево навыков
		skillTree, err := rpg.LoadSkillTree(&m.player)
		if err == nil {
			var availableSkills []player.SkillNode
			for _, node := range skillTree {
				if rpg.IsSkillAvailable(&m.player, node) {
					availableSkills = append(availableSkills, node)
				}
			}

			if len(availableSkills) > 0 {
				m.state = stateLevelUp
				m.skillChoices = availableSkills
				m.cursor = 0
			} else {
				// Если доступных навыков нет, просто повышаем уровень
				player.LevelUpPlayer("") // Передаем пустой ID
				p, _ := player.LoadPlayer()
				m.player = *p
				m.statusMessage = "Новый уровень! Доступных для изучения навыков пока нет."
			}
		} else {
			// Если не удалось загрузить дерево, все равно повышаем уровень
			player.LevelUpPlayer("")
			p, _ := player.LoadPlayer()
			m.player = *p
			m.statusMessage = "Новый уровень! Ошибка загрузки дерева нав��ков."
		}
	}

	return m
}

func (m *Model) buildQuestFilters() {
	filters := []string{"Все", "Daily", "Chore", "Quest", "Завершенные", "Просроченные"}
	tagSet := make(map[string]bool)
	for _, q := range m.quests {
		for _, tag := range q.Tags {
			if !tagSet[tag] {
				tagSet[tag] = true
			}
		}
	}

	var tags []string
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	m.allTags = tags

	if len(tags) > 0 {
		filters = append(filters, "---")
		for _, tag := range tags {
			filters = append(filters, tag)
		}
	}

	filters = append(filters, "---", "[Управление тегами]")
	m.questFilters = filters
}

func (m *Model) pushState(newState state) {
	m.stateStack = append(m.stateStack, m.state)
	m.state = newState
	m.statusMessage = ""
	m.cursor = 0
}

func (m *Model) popState() (tea.Model, tea.Cmd) {
	if len(m.stateStack) > 0 {
		lastStateIndex := len(m.stateStack) - 1
		m.state = m.stateStack[lastStateIndex]
		m.stateStack = m.stateStack[:lastStateIndex]
		m.statusMessage = ""
		m.cursor = 0
		return m, nil
	}
	return m, tea.Quit // Если стек пуст, выходим из приложения
}

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.state == stateDungeon {
		if tick, ok := msg.(dungeonTickMsg); ok {
			return m.updateDungeon(tick)
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
		return m, nil
	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" {
			return m, tea.Quit
		}
		if key == "q" || key == "esc" {
			return m.popState()
		}
		if key == "a" && (m.state == stateHomepage || m.state == stateQuests || m.state == stateQuestsFilter) {
			m.pushState(stateAddQuest)
			m.initAddQuest()
			return m, textinput.Blink
		}
		if key == "e" && m.state == stateQuests && len(m.displayQuests) > 0 {
			m.pushState(stateQuestEdit)
			m.initQuestEdit()
			return m, textinput.Blink
		}
	}

	switch m.state {
	case stateHomepage:
		return m.updateHomepage(msg)
	case stateQuests, stateQuestsFilter:
		return m.updateQuests(msg)
	case stateCompletedQuests:
		return m.updateCompletedQuests(msg)
	case stateAddQuest:
		return m.updateAddQuest(msg)
	case stateQuestEdit:
		return m.updateQuestEdit(msg)
	case statePerks:
		return m.updateSkillTreeView(msg)
	case stateSkills:
		return m.updateSkills(msg)
	case stateLevelUp:
		return m.updateLevelUp(msg)
	case stateManageTags:
		return m.updateManageTags(msg)
	}
	return m, nil
}

func (m *Model) getNavigationText() string {
	bindings, ok := KeyMap[m.state]
	if !ok {
		return ""
	}

	// Специальный случай для данжа, где текст зависит от под-состояния
	if m.state == stateDungeon && m.dungeonState == DungeonStateFinished {
		return "Нажмите любую клавишу для возврата."
	}

	var parts []string
	for _, binding := range bindings {
		parts = append(parts, binding.Key+": "+binding.Description)
	}
	return strings.Join(parts, " | ")
}

func (m *Model) View() string {
	var s strings.Builder
	switch m.state {
	case stateHomepage:
		s.WriteString(m.viewHomepage())
	case stateQuests, stateQuestsFilter:
		s.WriteString(m.viewQuests())
	case stateAddQuest:
		s.WriteString(m.viewAddQuest())
	case stateQuestEdit:
		s.WriteString(m.viewQuestEdit())
	case stateSkills:
		s.WriteString(m.viewSkills())
	case statePerks:
		s.WriteString(m.viewSkillTree())
	case stateDungeon:
		s.WriteString(m.viewDungeon())
	case stateDungeonPrep:
		s.WriteString(m.viewDungeonPrep())
	default:
		s.WriteString("Неизвестное состояние")
	}

	navText := m.getNavigationText()

	if navText != "" {
		s.WriteString("\n")
		s.WriteString(lipgloss.PlaceHorizontal(m.terminalWidth, lipgloss.Center, navText))
	}

	return s.String()
}
