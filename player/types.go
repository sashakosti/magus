package player

import "time"

// Position defines screen coordinates for layout
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type PlayerClass string

const (
	ClassNone      PlayerClass = ""
	ClassMage      PlayerClass = "Маг"
	ClassWarrior   PlayerClass = "Воин"
	ClassRogue     PlayerClass = "Разбойник"
)

type Player struct {
	Name            string          `json:"name"`
	Class           PlayerClass     `json:"class,omitempty"`
	Level           int             `json:"level"`
	HP              int             `json:"hp"`
	MaxHP           int             `json:"max_hp"`
	XP              int             `json:"xp"`
	NextLevelXP     int             `json:"next_level_xp"`
	Gold            int             `json:"gold"`
	Mana            int             `json:"mana"`
	MaxMana         int             `json:"max_mana"`
	UnlockedSkills  []string        `json:"unlocked_skills"` // Changed from Perks
	SkillPoints     int             `json:"skill_points"`
	Skills          map[string]int  `json:"skills"`
	LastCompletedAt time.Time       `json:"last_completed_at,omitempty"`
	History         struct {
		QuestsCompleted int `json:"quests_completed"`
		XPGained        int `json:"xp_gained"`
	} `json:"history"`
	LastSeen time.Time `json:"last_seen,omitempty"`
}

// ... (rest of the file)

type SkillType string

const (
	TypeActive  SkillType = "ACTIVE"
	TypePassive SkillType = "PASSIVE"
	TypeStat    SkillType = "STAT"
)

type SkillNode struct {
	ID               string             `json:"id"`
	Name             string             `json:"name"`
	Icon             string             `json:"icon"`
	Description      string             `json:"description"`
	Type             SkillType          `json:"type"`
	ClassRequirement string             `json:"class_requirement,omitempty"`
	Requirements     []string           `json:"requirements"`
	Position         Position           `json:"position"`
	Effects          map[string]float64 `json:"effects,omitempty"`
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
	Epic  QuestType = "epic"
	Chore QuestType = "chore"
)

type Quest struct {
	ID          string     `json:"id"`
	ParentID    string     `json:"parent_id,omitempty"` // ID родительского квеста
	Title       string     `json:"title"`
	Type        QuestType  `json:"type"`
	XP          int        `json:"xp"`
	Tags        []string   `json:"tags,omitempty"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	Completed   bool       `json:"completed"`
	CompletedAt time.Time  `json:"completed_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type History struct {
	QuestsCompleted int `json:"quests_completed"`
	XPGained        int `json:"xp_gained"`
}
