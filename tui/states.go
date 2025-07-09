package tui

type state int

const (
	stateHomepage state = iota
	stateQuests
	stateCompletedQuests
	stateAddQuest
	stateLevelUp
	stateSkills
	stateClassChoice
	stateCreatePlayer
	stateDungeonPrep
	stateDungeon
)
