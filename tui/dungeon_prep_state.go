package tui

import (
	"fmt"
	"io"
	"magus/player"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- list items ---

type durationItem struct {
	duration time.Duration
}

func (i durationItem) Title() string       { return fmt.Sprintf("%d минут", int(i.duration.Minutes())) }
func (i durationItem) Description() string { return "Длительность фокус-сессии" }
func (i durationItem) FilterValue() string { return i.Title() }

// --- delegate for quest list ---

type prepQuestDelegate struct {
	selectedQuests *map[string]struct{}
	styles         list.DefaultItemStyles
}

func (d prepQuestDelegate) Height() int                               { return 1 }
func (d prepQuestDelegate) Spacing() int                              { return 0 }
func (d prepQuestDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d prepQuestDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(QuestListItem)
	if !ok {
		return
	}

	check := "[ ]"
	if _, selected := (*d.selectedQuests)[item.ID]; selected && item.Type == player.TypeFocus {
		check = "[x]"
	}

	indent := lipgloss.NewStyle().PaddingLeft(item.Depth * 2).String()

	var title string
	if item.Type == player.TypeGoal {
		title = item.Title
	} else {
		title = fmt.Sprintf("%s %s", check, item.Title)
	}

	var style lipgloss.Style
	if index == m.Index() {
		style = d.styles.SelectedTitle
	} else {
		style = d.styles.NormalTitle
	}

	fmt.Fprint(w, style.Render(indent+title))
}

// --- model ---

type prepFocusable int

const (
	prepFocusDuration prepFocusable = iota
	prepFocusQuests
	prepFocusButton
)

type dungeonPrepModel struct {
	durationList   list.Model
	questList      list.Model
	allQuests      []player.Quest
	selectedQuests map[string]struct{}
	focused        prepFocusable
	statusMessage  string
}

func NewDungeonPrepState(m *Model) State {
	selectedQuests := make(map[string]struct{})

	durations := []list.Item{
		durationItem{duration: 15 * time.Minute},
		durationItem{duration: 25 * time.Minute},
		durationItem{duration: 45 * time.Minute},
	}

	durationDelegate := list.NewDefaultDelegate()
	durationDelegate.Styles.SelectedTitle.Foreground(lipgloss.Color("205"))
	durationList := list.New(durations, durationDelegate, 0, 0)
	durationList.Title = "Длительность"
	durationList.SetShowHelp(false)
	durationList.SetShowStatusBar(true)

	questDelegateStyles := list.NewDefaultItemStyles()
	questDelegateStyles.SelectedTitle.Foreground(lipgloss.Color("205"))
	questDelegate := prepQuestDelegate{
		selectedQuests: &selectedQuests,
		styles:         questDelegateStyles,
	}

	questList := list.New(nil, questDelegate, 0, 0)
	questList.Title = "Квесты для сессии"
	questList.SetShowHelp(false)
	questList.SetShowStatusBar(true)

	s := &dungeonPrepModel{
		durationList:   durationList,
		questList:      questList,
		allQuests:      m.Quests,
		selectedQuests: selectedQuests,
		focused:        prepFocusDuration,
	}
	s.questList.SetItems(BuildQuestListItems(s.allQuests, s.questList.Items()))
	return s
}

func (s *dungeonPrepModel) Init() tea.Cmd {
	return nil
}

func (s *dungeonPrepModel) Update(m *Model, msg tea.Msg) (State, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// tea.WindowSizeMsg is handled by the main model.
	// We'll resize components in the View function.

	if msg, ok := msg.(tea.KeyMsg); ok {
		s.statusMessage = ""
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))):
			return &PopState{}, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			s.focused = (s.focused + 1) % 3
		case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab"))):
			s.focused = (s.focused - 1 + 3) % 3
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if s.focused == prepFocusButton {
				selectedDuration := s.durationList.SelectedItem().(durationItem).duration
				manaCost := int(selectedDuration.Minutes() / 5)
				if m.Player.Mana < manaCost {
					s.statusMessage = fmt.Sprintf("Недостаточно маны! Нужно %d, у вас %d.", manaCost, m.Player.Mana)
					return s, nil
				}
				m.Player.Mana -= manaCost
				player.SavePlayer(m.Player)
				return NewDungeonState(m, selectedDuration), nil
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys(" "))):
			if s.focused == prepFocusQuests {
				item, ok := s.questList.SelectedItem().(QuestListItem)
				if ok && item.Type == player.TypeFocus {
					if _, exists := s.selectedQuests[item.ID]; exists {
						delete(s.selectedQuests, item.ID)
					} else {
						s.selectedQuests[item.ID] = struct{}{}
					}
				}
			}
		}
	}

	if s.focused == prepFocusDuration {
		s.durationList, cmd = s.durationList.Update(msg)
		cmds = append(cmds, cmd)
	} else if s.focused == prepFocusQuests {
		s.questList, cmd = s.questList.Update(msg)
		cmds = append(cmds, cmd)
	}

	return s, tea.Batch(cmds...)
}

func (s *dungeonPrepModel) View(m *Model) string {
	// Set the size of the lists before rendering
	availableWidth := m.TerminalWidth - 4  // Account for padding
	availableHeight := m.TerminalHeight - 8 // Account for titles, button, help text
	listWidth := availableWidth / 2
	hFrame, vFrame := m.styles.QuestCardStyle.GetFrameSize()
	s.durationList.SetSize(listWidth-hFrame, availableHeight-vFrame)
	s.questList.SetSize(listWidth-hFrame, availableHeight-vFrame)

	var durationStyle, questStyle lipgloss.Style

	if s.focused == prepFocusDuration {
		durationStyle = m.styles.QuestCardStyle.Copy().BorderForeground(lipgloss.Color("205"))
	} else {
		durationStyle = m.styles.QuestCardStyle
	}

	if s.focused == prepFocusQuests {
		questStyle = m.styles.QuestCardStyle.Copy().BorderForeground(lipgloss.Color("205"))
	} else {
		questStyle = m.styles.QuestCardStyle
	}

	button := "[ Начать ]"
	if s.focused == prepFocusButton {
		button = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("> Начать <")
	}

	lists := lipgloss.JoinHorizontal(lipgloss.Top,
		durationStyle.Render(s.durationList.View()),
		questStyle.Render(s.questList.View()),
	)

	var bottomContent string
	if s.statusMessage != "" {
		bottomContent = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(s.statusMessage)
	} else {
		bottomContent = m.styles.StatusMessageStyle.Render("tab: сменить фокус | space: выбрать квест | enter: начать | q/esc: назад")
	}

	mainContent := lipgloss.JoinVertical(lipgloss.Center,
		lists,
		button,
		bottomContent,
	)

	return lipgloss.Place(m.TerminalWidth, m.TerminalHeight, lipgloss.Center, lipgloss.Center, mainContent)
}