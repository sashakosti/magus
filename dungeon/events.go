package dungeon

import "time"

// EventType defines the type of a dungeon event
type EventType int

const (
	EventTypeAttack EventType = iota
	EventTypePlayerAttack
	EventTypeMonsterAttack
	EventTypeLoot
	EventTypeMessage
)

// DungeonEvent represents a single event that occurs in the dungeon log.
type DungeonEvent struct {
	Timestamp time.Time
	Type      EventType
	Message   string
}
