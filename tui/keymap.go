package tui

// KeyBinding представляет собой привязку клавиши к действию с описанием.
type KeyBinding struct {
	Key         string
	Description string
}

// KeyMap сопоставляет каждое состояние TUI со списком доступных действий.
var KeyMap = map[state][]KeyBinding{
	stateHomepage: {
		{Key: "↑/↓", Description: "навигация"},
		{Key: "enter", Description: "выбрать"},
		{Key: "a", Description: "добавить квест"},
		{Key: "q", Description: "выход"},
	},
	stateQuests: {
		{Key: "↑/↓", Description: "навигация"},
		{Key: "enter", Description: "выполнить"},
		{Key: "space", Description: "свернуть"},
		{Key: "e", Description: "редактировать"},
		{Key: "←", Description: "к фильтрам"},
		{Key: "a", Description: "добавить"},
		{Key: "q", Description: "назад"},
	},
	stateQuestsFilter: {
		{Key: "↑/↓", Description: "навигация"},
		{Key: "enter/→", Description: "выбрать"},
		{Key: "a", Description: "добавить"},
		{Key: "q", Description: "назад"},
	},
	stateAddQuest: {
		{Key: "tab", Description: "переключить поле"},
		{Key: "enter", Description: "сохранить"},
		{Key: "esc/q", Description: "отмена"},
	},
	stateQuestEdit: {
		{Key: "tab", Description: "переключить поле"},
		{Key: "enter", Description: "сохранить"},
		{Key: "esc/q", Description: "отмена"},
	},
	stateSkills: {
		{Key: "↑/↓", Description: "навигация"},
		{Key: "enter", Description: "улучшить"},
		{Key: "q", Description: "назад"},
	},
	statePerks: {
		{Key: "↑/↓/←/→", Description: "навигация"},
		{Key: "enter", Description: "изучить"},
		{Key: "q", Description: "назад"},
	},
	stateLevelUp: {
		{Key: "↑/↓", Description: "навигация"},
		{Key: "enter", Description: "выбрать"},
	},
	stateDungeonPrep: {
		{Key: "↑/↓", Description: "навигация"},
		{Key: "enter", Description: "выбрать"},
		{Key: "q", Description: "назад"},
	},
	stateDungeon: {
		{Key: "q", Description: "сбежать"},
	},
	stateManageTags: {
		{Key: "↑/↓", Description: "навигация"},
		{Key: "d", Description: "удалить"},
		{Key: "r", Description: "переименовать"},
		{Key: "q", Description: "назад"},
	},
}
