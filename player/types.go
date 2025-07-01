package player

import "time"

type Player struct {
	Name        string   `json:"name"`
	Level       int      `json:"level"`
	XP          int      `json:"xp"`
	NextLevelXP int      `json:"next_level_xp"`
	Perks       []string `json:"perks"`
	History     struct {
		QuestsCompleted int `json:"quests_completed"`
		XPGained        int `json:"xp_gained"`
	} `json:"history"`
}

type Stats struct {
	Intelligence int `json:"intelligence"`
	Charisma     int `json:"charisma"`
	Willpower    int `json:"willpower"`
	Discipline   int `json:"discipline"`
}

type QuestType string

const (
	Daily QuestType = "daily"
	Arc   QuestType = "arc"
	Meta  QuestType = "meta"
)

type Quest struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Type        QuestType `json:"type"`
	XP          int       `json:"xp"`
	Completed   bool      `json:"completed"` // Для квестов типа Arc и Meta
	CompletedAt time.Time `json:"completed_at"` // Для дейликов
	CreatedAt   time.Time `json:"created_at"`
}
type History struct {
	QuestsCompleted int `json:"quests_completed"`
	XPGained        int `json:"xp_gained"`
}
