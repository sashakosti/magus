package tui

type state int

const (
	stateHomepage state = iota
	stateQuests
	stateQuestsFilter
	stateCompletedQuests
	stateAddQuest
	stateLevelUp
	stateSkills
	stateClassChoice
	stateCreatePlayer
	stateDungeonPrep
	stateDungeon
	stateManageTags
	stateQuestEdit
	statePerks
)
