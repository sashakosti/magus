package tui

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"magus/dungeon"
	"magus/player"

	"github.com/charmbracelet/bubbletea"
)

// dungeonTickMsg is sent on every game tick.
type dungeonTickMsg time.Time

func (m *Model) updateDungeon(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			m.dungeonTicker.Stop()
			// Player escapes with what they've earned so far
			m.player.Gold += m.dungeonRunGold
			m.player.XP += m.dungeonRunXP
			player.SavePlayer(&m.player)
			m.state = stateHomepage
			m.statusMessage = fmt.Sprintf("Вы сбежали, сохранив %d золота и %d XP.", m.dungeonRunGold, m.dungeonRunXP)
			return m, nil
		}

	case dungeonTickMsg:
		// Check if the run is over
		remaining := m.dungeonSelectedDuration - time.Since(m.dungeonStartTime)
		if remaining <= 0 {
			return m.handleDungeonSuccess(), nil
		}

		// If combat is over, explore again. Otherwise, do a combat turn.
		if m.dungeonState == DungeonStateExploring {
			m.handleExplore()
		} else if m.dungeonState == DungeonStateInCombat {
			m.handleAutoCombatTurn()
		}

		// Wait for the next tick
		return m, func() tea.Msg {
			return dungeonTickMsg(<-m.dungeonTicker.C)
		}
	}

	return m, nil
}

// handleExplore finds a new monster for the player to fight.
func (m *Model) handleExplore() {
	m.dungeonState = DungeonStateExploring
	monster := dungeon.GetMonsterForFloor(m.dungeonFloor)
	m.currentMonster = &monster
	m.dungeonState = DungeonStateInCombat
	m.addDungeonLog(fmt.Sprintf("На этаже %d появляется %s!", m.dungeonFloor, m.currentMonster.Name))
}

// handleAutoCombatTurn simulates one round of combat.
func (m *Model) handleAutoCombatTurn() {
	if m.currentMonster == nil || m.player.HP <= 0 {
		return
	}

	// Player attacks monster
	playerDamage := m.player.Level*2 + rand.Intn(6)
	m.currentMonster.HP -= playerDamage
	m.addDungeonLog(fmt.Sprintf("Вы атаковали %s на %d урона.", m.currentMonster.Name, playerDamage))

	if m.currentMonster.HP <= 0 {
		m.handleMonsterVictory()
		return
	}

	// Monster attacks player
	monsterDamage := m.currentMonster.Attack - rand.Intn(m.player.Level)
	if monsterDamage < 0 {
		monsterDamage = 0
	}
	m.player.HP -= monsterDamage
	m.addDungeonLog(fmt.Sprintf("%s атакует вас на %d урона.", m.currentMonster.Name, monsterDamage))

	if m.player.HP <= 0 {
		m.handleDungeonExhaustion()
	}
}

// handleMonsterVictory is called when a monster is defeated.
func (m *Model) handleMonsterVictory() {
	m.addDungeonLog(fmt.Sprintf("%s побежден!", m.currentMonster.Name))
	xpGained := m.currentMonster.XPValue
	goldGained := m.currentMonster.GoldValue
	m.dungeonRunXP += xpGained
	m.dungeonRunGold += goldGained
	m.addDungeonLog(fmt.Sprintf("Получено (в этом забеге): %d XP, %d золота.", xpGained, goldGained))

	m.currentMonster = nil
	m.dungeonFloor++
	m.dungeonState = DungeonStateExploring // Ready for the next monster
}

// handleDungeonSuccess is called when the timer runs out successfully.
func (m *Model) handleDungeonSuccess() tea.Model {
	m.dungeonTicker.Stop()
	m.dungeonState = DungeonStateFinished

	bonusXP := int(float64(m.dungeonRunXP) * 0.2)
	bonusGold := int(float64(m.dungeonRunGold) * 0.2)
	totalXP := m.dungeonRunXP + bonusXP
	totalGold := m.dungeonRunGold + bonusGold

	m.player.XP += totalXP
	m.player.Gold += totalGold
	player.SavePlayer(&m.player)

	m.addDungeonLog("Время вышло! Забег успешен!")
	m.addDungeonLog(fmt.Sprintf("Награда: %d XP, %d золота.", m.dungeonRunXP, m.dungeonRunGold))
	m.addDungeonLog(fmt.Sprintf("Бонус: +%d XP, +%d золота.", bonusXP, bonusGold))
	m.statusMessage = "Успешный забег! Вы получили бонусную награду."
	return m
}

// handleDungeonExhaustion is called when the player's HP drops to 0.
func (m *Model) handleDungeonExhaustion() {
	m.dungeonTicker.Stop()
	m.dungeonState = DungeonStateFinished
	m.player.HP = 0

	m.player.XP += m.dungeonRunXP
	m.player.Gold += m.dungeonRunGold
	player.SavePlayer(&m.player)

	m.addDungeonLog("Вы пали без сил... но сохранили добычу.")
	m.addDungeonLog("Бонус за выносливость не получен.")
	m.statusMessage = "Вы истощены, но сохранили все, что успели собрать."
}

func (m *Model) addDungeonLog(message string) {
	m.dungeonLog = append(m.dungeonLog, message)
	if len(m.dungeonLog) > 10 {
		m.dungeonLog = m.dungeonLog[1:]
	}
}

func (m *Model) viewDungeon() string {
	var b strings.Builder

	// Timer
	remaining := m.dungeonSelectedDuration - time.Since(m.dungeonStartTime)
	if remaining < 0 {
		remaining = 0
	}
	timerStr := fmt.Sprintf("⏳ Осталось: %s", formatDuration(remaining))
	title := fmt.Sprintf("⚔️ Данж (Этаж %d) ⚔️", m.dungeonFloor)
	header := fmt.Sprintf("%s\n%s", title, timerStr)
	b.WriteString(header + "\n\n")

	// Player & Monster Stats
	playerHP := fmt.Sprintf("Ваше здоровье: %d / %d", m.player.HP, m.player.MaxHP)
	b.WriteString(playerHP + "\n")
	if m.dungeonState == DungeonStateInCombat && m.currentMonster != nil {
		monsterHP := fmt.Sprintf("%s: %d / %d", m.currentMonster.Name, m.currentMonster.HP, m.currentMonster.MaxHP)
		b.WriteString(monsterHP + "\n")
	}
	b.WriteString("\n")

	// Run Stats
	runStats := fmt.Sprintf("💰 Золото в забеге: %d | ✨ XP в забеге: %d", m.dungeonRunGold, m.dungeonRunXP)
	b.WriteString(runStats + "\n\n")

	// Log
	b.WriteString("--- Лог событий ---\n")
	for _, entry := range m.dungeonLog {
		b.WriteString(entry + "\n")
	}
	b.WriteString("------------------\n\n")

	// Actions
	if m.dungeonState == DungeonStateFinished {
		b.WriteString("Забег окончен. Нажмите 'q', чтобы выйти.\n")
	} else {
		b.WriteString("Идет автоматический бой... Нажмите 'q', чтобы сбежать.\n")
	}

	return b.String()
}
