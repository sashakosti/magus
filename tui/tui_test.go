package tui

import (
	"magus/player"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// newTestModel создает базовую модель для тестов.
func newTestModel() *Model {
	p, _ := player.CreatePlayer("tester")
	return &Model{
		Player: p,
		Quests: []player.Quest{},
		styles: NewStyles(),
	}
}

// TestDungeonPrepFocusing проверяет переключение фокуса в dungeonPrepModel.
func TestDungeonPrepFocusing(t *testing.T) {
	m := newTestModel()
	// Начинаем с состояния подготовки к данжу
	s := NewDungeonPrepState(m).(*dungeonPrepModel)

	// Изначально фокус на выборе длительности
	if s.focused != prepFocusDuration {
		t.Errorf("expected initial focus on duration, got %v", s.focused)
	}

	// Симулируем нажатие Tab
	tabKey := tea.KeyMsg{Type: tea.KeyTab}
	_, _ = s.Update(m, tabKey)

	// Фокус должен перейти на список квестов
	if s.focused != prepFocusQuests {
		t.Errorf("expected focus to move to quests on Tab, got %v", s.focused)
	}

	// Симулируем еще одно нажатие Tab
	_, _ = s.Update(m, tabKey)

	// Фокус должен перейти на кнопку "Начать"
	if s.focused != prepFocusButton {
		t.Errorf("expected focus to move to button on Tab, got %v", s.focused)
	}

	// Симулируем еще одно нажатие Tab (цикл)
	_, _ = s.Update(m, tabKey)

	// Фокус должен вернуться на выбор длительности
	if s.focused != prepFocusDuration {
		t.Errorf("expected focus to cycle back to duration, got %v", s.focused)
	}

	// Симулируем Shift+Tab
	shiftTabKey := tea.KeyMsg{Type: tea.KeyShiftTab}
	_, _ = s.Update(m, shiftTabKey)

	// Фокус должен вернуться на кнопку
	if s.focused != prepFocusButton {
		t.Errorf("expected focus to move back to button on Shift+Tab, got %v", s.focused)
	}
}

// TestQuestCompletion checks if completing a quest correctly gives XP.
func TestQuestCompletion(t *testing.T) {
	m := newTestModel()
	initialXP := m.Player.XP
	questXP := 50

	// Добавляем квест для теста
	testQuest := player.Quest{
		ID:    "test-quest-1",
		Title: "Test Quest",
		Type:  player.TypeGoal, // Используем тип Goal, который можно завершить из меню
		XP:    questXP,
	}
	m.Quests = append(m.Quests, testQuest)

	// Создаем состояние списка квестов
	s := NewQuestsState(m)

	// Убедимся, что курсор на нашем квесте (он должен быть единственным)
	if s.list.Index() != 0 {
		t.Fatalf("expected list index to be 0, got %d", s.list.Index())
	}
	selectedItem, ok := s.list.SelectedItem().(QuestListItem)
	if !ok || selectedItem.ID != testQuest.ID {
		t.Fatalf("did not select the correct quest")
	}

	// Симулируем нажатие Enter для завершения квеста
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := s.Update(m, enterKey)

	// Проверяем, что игрок получил правильное количество опыта
	// Для этого нам нужно загрузить игрока заново, т.к. AddXP сохраняет его отдельно.
	p, err := player.LoadPlayer()
	if err != nil {
		t.Fatalf("could not load player: %v", err)
	}

	expectedXP := initialXP + questXP
	if p.XP != expectedXP {
		t.Errorf("expected player XP to be %d, got %d", expectedXP, p.XP)
	}

	// Проверяем, что квест помечен как выполненный
	questCompleted := false
	for _, q := range m.Quests {
		if q.ID == testQuest.ID {
			if q.Completed {
				questCompleted = true
			}
			break
		}
	}
	if !questCompleted {
		t.Errorf("quest was not marked as completed")
	}

	// Проверяем, что было отправлено статусное сообщение
	if cmd == nil {
		t.Fatal("expected a status message command, but got nil")
	}
}
