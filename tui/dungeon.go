package tui

import (
	"fmt"
	"math/rand"
	"strings"

	"magus/dungeon"
	"magus/player"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) updateDungeon(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		// Always allow quitting
		if key.String() == "q" {
			m.state = stateHomepage
			m.statusMessage = "Вы покинули данж."
			return m, nil
		}

		switch m.dungeonState {
		case DungeonStateExploring:
			if key.String() == "e" {
				m.handleExplore()
			}
		case DungeonStateInCombat:
			switch strings.ToLower(key.String()) {
			case "a":
				m.handleCombatAttack()
			case "f":
				m.handleCombatFlee()
			}
		case DungeonStateFinished:
			// Any key press returns to the homepage
			m.state = stateHomepage
		}
	}
	return m, nil
}

func (m *Model) handleExplore() {
	m.addDungeonLog("Вы исследуете темные коридоры...")

	// 40% chance to encounter a monster
	if rand.Intn(100) < 40 {
		monster := dungeon.Monsters[rand.Intn(len(dungeon.Monsters))]
		m.currentMonster = &monster // Create a new instance for the fight
		m.dungeonState = DungeonStateInCombat
		m.addDungeonLog(fmt.Sprintf("На вашем пути появляется %s!", m.currentMonster.Name))
	} else {
		// 10% chance to find gold
		if rand.Intn(100) < 10 {
			goldFound := rand.Intn(10) + 1
			m.player.Gold += goldFound
			player.SavePlayer(&m.player)
			m.addDungeonLog(fmt.Sprintf("Вы нашли %d золота!", goldFound))
		} else {
			m.addDungeonLog("Ничего интересного не найдено.")
		}
	}
}

func (m *Model) handleCombatAttack() {
	if m.currentMonster == nil {
		return
	}

	// Player attacks monster
	playerDamage := m.player.Level*2 + rand.Intn(6) // e.g. 2-7 for level 1
	m.currentMonster.HP -= playerDamage
	m.addDungeonLog(fmt.Sprintf("Вы атаковали %s на %d урона.", m.currentMonster.Name, playerDamage))

	if m.currentMonster.HP <= 0 {
		m.addDungeonLog(fmt.Sprintf("%s побежден!", m.currentMonster.Name))
		xpGained := m.currentMonster.XPValue
		goldGained := m.currentMonster.GoldValue
		m.player.XP += xpGained
		m.player.Gold += goldGained
		player.SavePlayer(&m.player)
		m.addDungeonLog(fmt.Sprintf("Вы получили %d XP и %d золота.", xpGained, goldGained))

		m.currentMonster = nil
		m.dungeonState = DungeonStateExploring
		// Check for level up
		if m.player.XP >= m.player.NextLevelXP {
			m.state = stateLevelUp // Go to level up screen
		}
		return
	}

	// Monster attacks player
	monsterDamage := m.currentMonster.Attack - rand.Intn(m.player.Level) // Player level acts as defense
	if monsterDamage < 0 {
		monsterDamage = 0
	}
	m.player.HP -= monsterDamage
	m.addDungeonLog(fmt.Sprintf("%s атакует вас на %d урона.", m.currentMonster.Name, monsterDamage))

	if m.player.HP <= 0 {
		m.player.HP = 0
		m.addDungeonLog("Вы были побеждены... Поход окончен.")
		m.dungeonState = DungeonStateFinished
	}
	player.SavePlayer(&m.player)
}

func (m *Model) handleCombatFlee() {
	if rand.Intn(100) < 50 { // 50% chance to flee
		m.addDungeonLog("Вы успешно сбежали.")
		m.currentMonster = nil
		m.dungeonState = DungeonStateExploring
	} else {
		m.addDungeonLog("Не удалось сбежать!")
		// Monster gets a free attack
		monsterDamage := m.currentMonster.Attack - rand.Intn(m.player.Level)
		if monsterDamage < 0 {
			monsterDamage = 0
		}
		m.player.HP -= monsterDamage
		m.addDungeonLog(fmt.Sprintf("%s атакует вас на %d урона, пока вы пытались сбежать.", m.currentMonster.Name, monsterDamage))

		if m.player.HP <= 0 {
			m.player.HP = 0
			m.addDungeonLog("Вы были побеждены... Поход окончен.")
			m.dungeonState = DungeonStateFinished
		}
		player.SavePlayer(&m.player)
	}
}

func (m *Model) addDungeonLog(message string) {
	m.dungeonLog = append(m.dungeonLog, message)
	// Keep the log from getting too long
	if len(m.dungeonLog) > 10 {
		m.dungeonLog = m.dungeonLog[1:]
	}
}

func (m *Model) viewDungeon() string {
	var b strings.Builder

	b.WriteString("⚔️ Данж ⚔️\n\n")

	// Player HP
	playerHP := fmt.Sprintf("Ваше здоровье: %d / %d", m.player.HP, m.player.MaxHP)
	b.WriteString(playerHP + "\n\n")

	if m.dungeonState == DungeonStateInCombat && m.currentMonster != nil {
		// Monster Info
		monsterArt, ok := dungeon.MonsterArt[m.currentMonster.Name]
		if !ok {
			monsterArt = "???"
		}
		monsterHP := fmt.Sprintf("%s | Здоровье: %d / %d", m.currentMonster.Name, m.currentMonster.HP, m.currentMonster.MaxHP)
		monsterBox := lipgloss.JoinHorizontal(lipgloss.Top, monsterArt, monsterHP)
		b.WriteString(monsterBox + "\n\n")
	}

	// Log
	b.WriteString("--- Лог событий ---\n")
	for _, entry := range m.dungeonLog {
		b.WriteString(entry + "\n")
	}
	b.WriteString("------------------\n\n")

	// Actions
	b.WriteString("--- Действия ---\n")
	switch m.dungeonState {
	case DungeonStateExploring:
		b.WriteString("[e] - Исследовать\n")
	case DungeonStateInCombat:
		b.WriteString("[a] - Атаковать\n")
		b.WriteString("[f] - Сбежать\n")
	case DungeonStateFinished:
		b.WriteString("Поход окончен. Нажмите любую клавишу, чтобы выйти.\n")
	}
	b.WriteString("[q] - Покинуть данж\n")
	b.WriteString("----------------\n")

	return b.String()
}
