package tui

import (
	"magus/player"
	"magus/storage"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	stateStack   []State
	currentState State

	Player         *player.Player
	Quests         []player.Quest
	TerminalWidth  int
	TerminalHeight int
	ready          bool // Флаг готовности к отрисовке
	styles         Styles
}

type Styles struct {
	TitleStyle             lipgloss.Style
	QuestCardStyle         lipgloss.Style
	SelectedQuestCardStyle lipgloss.Style
	FaintQuestCardStyle    lipgloss.Style
	StatusMessageStyle     lipgloss.Style
	MetaStyle              lipgloss.Style
	RitualStyle            lipgloss.Style
	FocusStyle             lipgloss.Style
	TagStyle               lipgloss.Style
	DifficultyStyle        lipgloss.Style
	DeadlineStyle          lipgloss.Style
	CompletedIcon          string
	GoalIcon               string
	RitualIcon             string
	FocusIcon              string
	CollapseIconOpened     string
	CollapseIconClosed     string
	SubQuestIndent         string
}

func NewStyles() Styles {
	return Styles{
		TitleStyle:             lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFDF5")).Bold(true).Padding(0, 1),
		QuestCardStyle:         lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")).Padding(0, 1),
		SelectedQuestCardStyle: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#AD58B4")).Padding(0, 1),
		FaintQuestCardStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("238")).Padding(0, 1),
		StatusMessageStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("#A89F94")).Italic(true),
		MetaStyle:              lipgloss.NewStyle().Foreground(lipgloss.Color("#FFC300")),
		RitualStyle:            lipgloss.NewStyle().Foreground(lipgloss.Color("#36A2EB")),
		FocusStyle:             lipgloss.NewStyle().Foreground(lipgloss.Color("#9B59B6")),
		TagStyle:               lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Background(lipgloss.Color("236")).Padding(0, 1),
		DifficultyStyle:        lipgloss.NewStyle().Foreground(lipgloss.Color("#DAA520")),
		DeadlineStyle:          lipgloss.NewStyle().Foreground(lipgloss.Color("#E74C3C")),
		CompletedIcon:          lipgloss.NewStyle().Foreground(lipgloss.Color("#2ECC71")).Render("✔"),
		GoalIcon:               "🏆",
		RitualIcon:             "💧",
		FocusIcon:              "🎯",
		CollapseIconOpened:     "▼",
		CollapseIconClosed:     "▶",
		SubQuestIndent:         "   ",
	}
}

func InitialModel() *Model {
	p, err := player.LoadPlayer()
	if err != nil {
		return &Model{
			currentState: NewCreatePlayerState(),
			styles:       NewStyles(),
		}
	}

	if !p.LastSeen.IsZero() {
		minutesPassed := time.Since(p.LastSeen).Minutes()
		hpToRestore := int(minutesPassed / 5)
		if hpToRestore > 0 {
			p.HP += hpToRestore
			if p.HP > p.MaxHP {
				p.HP = p.MaxHP
			}
			player.SavePlayer(p)
		}
	}

	quests, _ := storage.LoadAllQuests()

	m := &Model{
		Player: p,
		Quests: quests,
		styles: NewStyles(),
	}

	if p.XP >= p.NextLevelXP {
		levelUpState, err := NewLevelUpState(m)
		if err != nil {
			m.currentState = NewHomepageState(m)
		} else {
			m.currentState = levelUpState
		}
	} else {
		m.currentState = NewHomepageState(m)
	}

	return m
}

func (m *Model) pushState(newState State) {
	m.stateStack = append(m.stateStack, m.currentState)
	m.currentState = newState
}

func (m *Model) popState(refresh ...bool) {
	if len(m.stateStack) > 0 {
		lastStateIndex := len(m.stateStack) - 1
		m.currentState = m.stateStack[lastStateIndex]
		m.stateStack = m.stateStack[:lastStateIndex]

		// Always recreate HomepageState and QuestsState when they are popped to avoid stale data
		// or issues with nested component states (like the progress bar).
		switch m.currentState.(type) {
		case *HomepageState:
			m.currentState = NewHomepageState(m)
			return // Exit early to avoid double-processing
		case *QuestsState:
			if len(refresh) > 0 && refresh[0] {
				m.Quests, _ = storage.LoadAllQuests()
			}
			m.currentState = NewQuestsState(m)
		}
	}
}

func (m *Model) Init() tea.Cmd {
	return m.currentState.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.TerminalWidth = msg.Width
		m.TerminalHeight = msg.Height
		if !m.ready {
			m.ready = true
		}
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	newState, cmd := m.currentState.Update(m, msg)
	cmds = append(cmds, cmd)

	if newState != m.currentState {
		if pop, ok := newState.(PopState); ok {
			m.popState(pop.refreshQuests)
		} else {
			m.pushState(newState)
			cmds = append(cmds, newState.Init())
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}
	return m.currentState.View(m)
}

func Start() error {
	p := tea.NewProgram(InitialModel(), tea.WithAltScreen())
	return p.Start()
}
