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
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	player          player.Player
	quests          []player.Quest
	displayQuests   []player.Quest
	perkChoices     []rpg.Perk
	skills          []rpg.Skill
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

	// Dungeon state
	dungeonState             dungeonState
	dungeonFloor             int
	dungeonRunGold           int
	dungeonRunXP             int
	dungeonDurationChoices   []string
	dungeonDurationCursor    int
	dungeonSelectedDuration  time.Duration
	dungeonStartTime         time.Time
	dungeonTicker            *time.Ticker
	dungeonCustomDurationInput textinput.Model
	currentMonster           *dungeon.Monster
	dungeonLog               []string

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
	skills, _ := rpg.LoadAllSkills()

	m := Model{
		state:                      stateHomepage,
		player:                     *p,
		quests:                     quests,
		skills:                     skills,
		cursor:                     0,
		statusMessage:              "",
		progressBar:                progress.New(progress.WithDefaultGradient(), progress.WithWidth(40), progress.WithoutPercentage()),
		collapsed:                  make(map[string]bool),
		homepageCursor:             0,
		addQuestTypes:              []player.QuestType{player.Daily, player.Arc, player.Meta, player.Epic, player.Chore},
		addQuestTypeIdx:            0,
		dungeonDurationChoices:     []string{"15", "25", "45", "Custom"},
		dungeonCustomDurationInput: textinput.New(),
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

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle dungeon ticks globally to ensure they are always processed
	// when the dungeon is running.
	if m.state == stateDungeon {
		if tick, ok := msg.(dungeonTickMsg); ok {
			return m.updateDungeon(tick)
		}
	}

	var cmd tea.Cmd
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
		if key == "q" {
			// Universal quit from most states
			if m.state == stateDungeon || m.state == stateQuests || m.state == stateCompletedQuests || m.state == stateSkills || m.state == stateClassChoice {
				if m.state == stateDungeon && m.dungeonTicker != nil {
					m.dungeonTicker.Stop()
				}
				m.state = stateHomepage
				if m.state == stateDungeon {
					m.statusMessage = "Вы сбежали из данжа."
				} else {
					m.statusMessage = ""
				}
				m.cursor = 0
				return m, nil
			}
			if m.state == stateHomepage || m.state == stateCreatePlayer {
				return m, tea.Quit
			}
			if m.state == stateAddQuest {
				m.addQuestInputs = nil
				m.state = stateHomepage
				m.statusMessage = ""
				m.cursor = 0
				return m, nil
			}
		}
		if key == "a" && m.state == stateHomepage {
			m.state = stateAddQuest
			m.addQuestCursor = 0
			m.addQuestTypeIdx = 0
			m.addQuestInputs = newAddQuestInputs()
			return m, nil
		}
	}

	switch m.state {
	case stateHomepage:
		return m.updateHomepage(msg)
	case stateQuests:
		return m.updateQuests(msg)
	case stateCompletedQuests:
		return m.updateCompletedQuests(msg)
	case stateAddQuest:
		return m.updateAddQuest(msg)
	case stateLevelUp:
		return m.updateLevelUp(msg)
	case stateSkills:
		return m.updateSkills(msg)
	case stateClassChoice:
		return m.updateClassChoice(msg)
	case stateCreatePlayer:
		return m.updateCreatePlayer(msg)
	case stateDungeonPrep:
		return m.updateDungeonPrep(msg)
	case stateDungeon:
		return m.updateDungeon(msg)
	}

	return m, cmd
}

func (m *Model) View() string {
	var s strings.Builder

	switch m.state {
	case stateHomepage:
		s.WriteString(m.viewHomepage())
	case stateQuests:
		s.WriteString(m.viewQuests())
	case stateCompletedQuests:
		s.WriteString(m.viewCompletedQuests())
	case stateAddQuest:
		s.WriteString(m.viewAddQuest())
	case stateLevelUp:
		s.WriteString(m.viewLevelUp())
	case stateSkills:
		s.WriteString(m.viewSkills())
	case stateClassChoice:
		s.WriteString(m.viewClassChoice())
	case stateCreatePlayer:
		s.WriteString(m.viewCreatePlayer())
	case stateDungeonPrep:
		s.WriteString(m.viewDungeonPrep())
	case stateDungeon:
		s.WriteString(m.viewDungeon())
	default:
		s.WriteString("Неизвестное состояние")
	}

	navText := ""
	switch m.state {
	case stateHomepage:
		navText = "Навигация: ↑/↓, Enter, 'a' - добавить, 'q' - выход."
	case stateQuests:
		navText = "Навигация: ↑/↓, Enter, [Пробел], 'a' - добавить, 'q' - назад."
	case stateAddQuest:
		navText = "Навигация: ↑/↓, ←/→, Enter, 'q' - отмена."
	case stateSkills:
		navText = "Нажмите 'enter' для улучшения, 'q' - назад."
	case stateDungeon:
		if m.dungeonState == DungeonStateFinished {
			navText = "Нажмите любую клавишу для возврата."
		} else {
			navText = "Поход в процессе... 'q' - сбежать."
		}
	}

	if navText != "" {
		s.WriteString("\n")
		s.WriteString(lipgloss.PlaceHorizontal(m.terminalWidth, lipgloss.Center, navText))
	}

	return s.String()
}
