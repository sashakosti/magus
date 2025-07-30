package tui

import (
	"fmt"
	"magus/player"
	"magus/storage"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type QuestsState struct {
	list          list.Model
	allQuests     []player.Quest // Мастер-список всех квестов
	statusMessage string
}

func NewQuestsState(m *Model) *QuestsState {
	s := &QuestsState{allQuests: m.Quests}

	delegate := NewQuestDelegate(&m.styles)
	questList := list.New(nil, delegate, 0, 0) // Start with an empty list
	questList.Title = "Активные квесты"
	questList.Styles.Title = m.styles.TitleStyle
	questList.SetShowStatusBar(false)
	questList.SetShowHelp(true)
	questList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "добавить")),
			key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "удалить")),
			key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "развернуть")),
		}
	}
	questList.AdditionalFullHelpKeys = func() []key.Binding {
		return questList.AdditionalShortHelpKeys()
	}

	s.list = questList
	// Инициализируем список с самого начала
	s.list.SetItems(BuildQuestListItems(s.allQuests, s.list.Items()))
	return s
}

func (s *QuestsState) Init() tea.Cmd {
	return nil
}

func (s *QuestsState) Update(m *Model, msg tea.Msg) (State, tea.Cmd) {
	var cmds []tea.Cmd

	// tea.WindowSizeMsg is handled by the main model
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if s.list.FilterState() == list.Filtering {
			break
		}
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("a"))):
			// Новое состояние для добавления квеста
			return NewAddQuestState(m), nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			return s.completeQuest(m)
		case key.Matches(msg, key.NewBinding(key.WithKeys("d", "delete"))):
			return s.deleteQuest(m)
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			s.toggleQuestExpansion()
			return s, nil
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "esc"))):
			return PopState{}, nil
		}
	}

	newListModel, cmd := s.list.Update(msg)
	s.list = newListModel
	cmds = append(cmds, cmd)

	return s, tea.Batch(cmds...)
}

func (s *QuestsState) View(m *Model) string {
	// Set the size of the list before rendering
	h, v := lipgloss.NewStyle().Margin(1, 2).GetFrameSize()
	s.list.SetSize(m.TerminalWidth-h, m.TerminalHeight-v)
	return lipgloss.NewStyle().Margin(1, 2).Render(s.list.View())
}

func (s *QuestsState) toggleQuestExpansion() {
	selectedItem, ok := s.list.SelectedItem().(QuestListItem)
	if !ok || !selectedItem.HasKids {
		return
	}

	// Найдем и обновим состояние в `s.list.Items()`
	for i, item := range s.list.Items() {
		qli := item.(QuestListItem)
		if qli.ID == selectedItem.ID {
			qli.IsExpanded = !qli.IsExpanded
			s.list.SetItem(i, qli) // Обновляем элемент в списке
			break
		}
	}

	// Перестраиваем список на основе обновленных состояний isExpanded
	s.list.SetItems(BuildQuestListItems(s.allQuests, s.list.Items()))
}

func (s *QuestsState) deleteQuest(m *Model) (State, tea.Cmd) {
	selectedItem, ok := s.list.SelectedItem().(QuestListItem)
	if !ok {
		return s, nil
	}

	// Найти все ID для удаления (выбранный квест + все дочерние)
	idsToDelete := make(map[string]struct{})
	var findChildren func(parentID string)
	findChildren = func(parentID string) {
		idsToDelete[parentID] = struct{}{}
		for _, q := range s.allQuests {
			if q.ParentID == parentID {
				findChildren(q.ID)
			}
		}
	}
	findChildren(selectedItem.ID)

	// Создать новый срез без удаленных квестов
	var updatedQuests []player.Quest
	for _, q := range s.allQuests {
		if _, found := idsToDelete[q.ID]; !found {
			updatedQuests = append(updatedQuests, q)
		}
	}

	s.allQuests = updatedQuests
	m.Quests = updatedQuests // Обновляем мастер-список в главной модели
	storage.SaveAllQuests(m.Quests)

	// Обновляем UI
	s.list.SetItems(BuildQuestListItems(s.allQuests, s.list.Items()))
	// Перемещаем курсор, если он был на последнем элементе, который удалили
	if s.list.Index() >= len(s.list.Items()) && len(s.list.Items()) > 0 {
		s.list.Select(len(s.list.Items()) - 1)
	}

	statusMsg := fmt.Sprintf("🗑️ Квест '%s' и все подзадачи удалены.", selectedItem.Title)
	return s, s.list.NewStatusMessage(statusMsg)
}

func (s *QuestsState) completeQuest(m *Model) (State, tea.Cmd) {
	selectedItem, ok := s.list.SelectedItem().(QuestListItem)
	if !ok {
		return s, nil
	}

	// Нельзя завершить уже завершенный квест
	if selectedItem.Completed {
		return s, s.list.NewStatusMessage("✅ Квест уже выполнен")
	}

	// Нельзя завершать фокус-квесты из этого меню
	if selectedItem.Type == player.TypeFocus {
		return s, s.list.NewStatusMessage("❗ Этот квест выполняется в фокус-сессии (подземелье)")
	}

	// Нельзя завершить цель напрямую, если у нее есть незавершенные подзадачи
	if selectedItem.Type == player.TypeGoal {
		hasIncompleteChildren := false
		for _, q := range s.allQuests {
			if q.ParentID == selectedItem.ID && !q.Completed {
				hasIncompleteChildren = true
				break
			}
		}
		if hasIncompleteChildren {
			statusMsg := fmt.Sprintf("❗ Сначала завершите все подзадачи для цели '%s'", selectedItem.Title)
			return s, s.list.NewStatusMessage(statusMsg)
		}
	}

	var xpGained int
	var manaGained int
	questCompleted := false

	// Обновляем квест в мастер-списке
	for i, q := range s.allQuests {
		if q.ID == selectedItem.ID {
			switch q.Type {
			case player.TypeRitual:
				// Ритуалы восстанавливают ману, не дают XP и не "завершаются" навсегда
				manaGained = 5 // Примерное значение, можно вынести в конфиг
				s.statusMessage = fmt.Sprintf("💧 +%d маны за ритуал '%s'", manaGained, q.Title)
				// Можно добавить кулдаун, но пока просто восстанавливаем ману
			case player.TypeFocus, player.TypeGoal:
				// Фокус-квесты и цели завершаются, дают XP
				s.allQuests[i].Completed = true
				s.allQuests[i].CompletedAt = time.Now()
				s.allQuests[i].Progress = s.allQuests[i].HP // Заполняем прогресс при завершении
				xpGained = s.allQuests[i].XP
				questCompleted = true
				s.statusMessage = fmt.Sprintf("✨ +%d XP за квест '%s'!", xpGained, q.Title)
			}
			m.Quests[i] = s.allQuests[i] // Обновляем квест в главной модели
			break
		}
	}

	// Обновляем данные игрока
	p, _ := player.LoadPlayer()
	p.Mana += manaGained
	if p.Mana > p.MaxMana {
		p.Mana = p.MaxMana
	}
	m.Player = p // Обновляем игрока в текущей модели
	player.SavePlayer(p)

	// Сохраняем и обновляем список
	storage.SaveAllQuests(m.Quests)
	s.list.SetItems(BuildQuestListItems(s.allQuests, s.list.Items()))

	// Проверка на повышение уровня, если был получен опыт
	if xpGained > 0 {
		canLevelUp, _ := player.AddXP(xpGained)
		if canLevelUp {
			levelUpState, err := NewLevelUpState(m)
			if err != nil {
				s.statusMessage = "🔮 Новый уровень! Доступных для изучения навыков пока нет."
				player.LevelUpPlayer("")
				p, _ := player.LoadPlayer()
				m.Player = p
			} else {
				return levelUpState, nil
			}
		}
	}

	if !questCompleted && manaGained == 0 {
		return s, nil // Ничего не произошло
	}

	return s, s.list.NewStatusMessage(s.statusMessage)
}