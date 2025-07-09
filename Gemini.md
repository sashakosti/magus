# Magus Project Context

## 1. Core Vision & Goal
Magus is a **gamified Pomodoro timer**. Its primary goal is to help users focus on real-world tasks by providing a non-distracting, rewarding background activity. The core principle is **"Focus Support, Not Distraction"**.

## 2. Key Mechanics
- **Dungeon as a "Focus Mode":** The dungeon is an auto-battler that runs for a user-defined duration. The player's goal is to prepare their character to be as effective as possible during this automated session.
- **No Punishment for Inattention:** The player never loses progress for being defeated during a run. They simply miss out on a performance bonus. This encourages setting the timer and focusing on the real task.
- **Gameplay Loop:**
  1.  **Prepare:** The player improves their character outside the dungeon (skills, perks).
  2.  **Focus:** The player starts a timed dungeon run to coincide with a real-world task.
  3.  **Reward:** The player is rewarded with XP and Gold based on their character's performance.
  4.  **Recover:** The player restores HP by completing real-life chores (marked as `Chore` quests) and taking breaks from the app.

## 3. Development Info
- **Run Command:** `go run .`
- **Test Command:** `go test ./...`
- **Main Data File:** `data/player.json`
- **Core Logic Directories:**
  - `/tui`: UI components and state management.
  - `/rpg`: Core game mechanics (classes, skills, perks).
  - `/dungeon`: Monster definitions and dungeon-specific logic.
  - `/player`: Player data structures and management.
