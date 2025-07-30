package tui

import (
	"time"

	"github.com/charmbracelet/bubbletea"
)

// State представляет собой самодостаточный модуль TUI (например, экран).
// Каждый State управляет своей собственной логикой, состоянием и отображением.
type State interface {
	// Init вызывается один раз при создании состояния для выполнения
	// начальных команд, таких как мигание курсора.
	Init() tea.Cmd

	// Update обрабатывает входящие сообщения (ввод пользователя, таймеры и т.д.).
	// Он может вернуть новое состояние для перехода (например, на другой экран)
	// или вернуть себя же, чтобы остаться на текущем экране.
	Update(m *Model, msg tea.Msg) (State, tea.Cmd)

	// View генерирует строковое представление для отображения на экране.
	View(m *Model) string
}

// PopState - это специальный тип-сигнал. Когда Update возвращает этот объект,
// главный цикл понимает, что нужно вернуться к предыдущему состоянию в стеке.
type PopState struct {
	refreshQuests bool
}

func (p PopState) Init() tea.Cmd                                 { return nil }
func (p PopState) Update(m *Model, msg tea.Msg) (State, tea.Cmd) { return p, nil }
func (p PopState) View(m *Model) string                          { return "" }

// DungeonResult содержит итоги фокус-сессии.
type DungeonResult struct {
	Duration           time.Duration
	DistractionAttacks int
	Success            bool
	XPEarned           int // Добавлено для случая досрочного выхода из старой версии
}