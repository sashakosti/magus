package tui

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"magus/player"
	"magus/rpg"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

const (
	skillGridWidth  = 60
	skillGridHeight = 15
	skillCellWidth  = 4
)

type skillNavDirection int

const (
	navUp skillNavDirection = iota
	navDown
	navLeft
	navRight
)

func (m *Model) updateSkillTreeView(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch key.String() {
	case "q", "esc":
		m.state = stateHomepage
		m.statusMessage = ""
		return m, nil
	case "up", "k":
		m.navigateSkillTree(navUp)
	case "down", "j":
		m.navigateSkillTree(navDown)
	case "left", "h":
		m.navigateSkillTree(navLeft)
	case "right", "l":
		m.navigateSkillTree(navRight)
	case "enter":
		skillTree, _ := rpg.LoadSkillTree(&m.player)
		selectedNode, ok := findNodeAt(m.perkCursorX, m.perkCursorY, skillTree)
		if !ok {
			return m, nil
		}
		err := m.unlockSkill(selectedNode.ID)
		if err != nil {
			m.statusMessage = fmt.Sprintf("❗ %v", err)
		} else {
			m.statusMessage = fmt.Sprintf("✨ Навык '%s' разблокирован!", selectedNode.Name)
		}
	}

	return m, nil
}

func (m *Model) navigateSkillTree(dir skillNavDirection) {
	skillTree, _ := rpg.LoadSkillTree(&m.player)
	if len(skillTree) == 0 {
		return
	}

	nodes := mapToSlice(skillTree)
	currentPos := player.Position{X: m.perkCursorX, Y: m.perkCursorY}
	var bestCandidate player.SkillNode
	minDist := math.MaxFloat64

	for _, node := range nodes {
		if node.Position == currentPos {
			continue
		}

		dist := distance(currentPos, node.Position)
		isDirectionMatch := false

		switch dir {
		case navUp:
			isDirectionMatch = node.Position.Y < currentPos.Y
		case navDown:
			isDirectionMatch = node.Position.Y > currentPos.Y
		case navLeft:
			isDirectionMatch = node.Position.X < currentPos.X
		case navRight:
			isDirectionMatch = node.Position.X > currentPos.X
		}

		if isDirectionMatch && dist < minDist {
			minDist = dist
			bestCandidate = node
		}
	}

	if bestCandidate.ID != "" {
		m.perkCursorX = bestCandidate.Position.X
		m.perkCursorY = bestCandidate.Position.Y
	}
}

func (m *Model) viewSkillTree() string {
	title := titleStyle.Render(fmt.Sprintf("🌳 Дерево перков (Очки: %d)", m.player.SkillPoints))

	skillTree, _ := rpg.LoadSkillTree(&m.player)
	nodes := mapToSlice(skillTree)

	if m.perkCursorX == 0 && m.perkCursorY == 0 && len(nodes) > 0 {
		sort.Slice(nodes, func(i, j int) bool {
			if nodes[i].Position.Y != nodes[j].Position.Y {
				return nodes[i].Position.Y < nodes[j].Position.Y
			}
			return nodes[i].Position.X < nodes[j].Position.X
		})
		m.perkCursorX = nodes[0].Position.X
		m.perkCursorY = nodes[0].Position.Y
	}

	grid := make([][]string, skillGridHeight)
	for i := range grid {
		grid[i] = make([]string, skillGridWidth)
		for j := range grid[i] {
			grid[i][j] = strings.Repeat(" ", skillCellWidth)
		}
	}

	drawSkillConnections(grid, skillTree)
	drawSkillNodes(grid, nodes, &m.player, m.perkCursorX, m.perkCursorY)

	var gridView strings.Builder
	for _, row := range grid {
		gridView.WriteString(strings.Join(row, ""))
		gridView.WriteString("\n")
	}

	perkBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Render(gridView.String())

	infoBox := createSkillInfoBox(m.perkCursorX, m.perkCursorY, skillTree, &m.player, lipgloss.Width(perkBox))

	mainContent := lipgloss.JoinVertical(lipgloss.Top, perkBox, infoBox)
	finalView := lipgloss.JoinVertical(lipgloss.Top, title, mainContent, statusMessageStyle.Render(m.statusMessage))

	return docStyle.Render(finalView)
}

func drawSkillConnections(grid [][]string, skillTree map[string]player.SkillNode) {
	for _, node := range skillTree {
		if node.Position.X >= skillGridWidth || node.Position.Y >= skillGridHeight {
			continue // Не рисовать для узлов за пределами сетки
		}
		endPos := node.Position
		for _, reqID := range node.Requirements {
			startNode, ok := skillTree[reqID]
			if !ok {
				continue
			}
			if startNode.Position.X >= skillGridWidth || startNode.Position.Y >= skillGridHeight {
				continue // Не рисовать от узлов за пределами сетки
			}
			startPos := startNode.Position

			x1, y1 := startPos.X, startPos.Y
			x2, y2 := endPos.X, endPos.Y

			// Вертикальная линия
			for y := y1 + 1; y < y2; y++ {
				setGridChar(grid, x1, y, "│")
			}

			// Горизонтальная линия
			for x := Min(x1, x2); x <= Max(x1, x2); x++ {
				if x != x1 {
					setGridChar(grid, x, y2, "─")
				}
			}

			// Символ-коннектор
			if y1 < y2 {
				currentVal, _ := getGridChar(grid, x1, y2)
				if currentVal == padCell("│") {
					if x1 < x2 {
						setGridChar(grid, x1, y2, "├")
					} else if x1 > x2 {
						setGridChar(grid, x1, y2, "┤")
					}
				} else {
					if x1 < x2 {
						setGridChar(grid, x1, y2, "└")
					} else if x1 > x2 {
						setGridChar(grid, x1, y2, "┘")
					} else {
						setGridChar(grid, x1, y2, "│")
					}
				}
			}
		}
	}
}

func drawSkillNodes(grid [][]string, nodes []player.SkillNode, p *player.Player, cursorX, cursorY int) {
	for _, node := range nodes {
		if node.Position.X >= skillGridWidth || node.Position.Y >= skillGridHeight {
			continue
		}
		isAvailable := rpg.IsSkillAvailable(p, node)
		isUnlocked := rpg.IsSkillUnlocked(p, node.ID)
		isSelected := cursorX == node.Position.X && cursorY == node.Position.Y
		isClassMismatch := node.ClassRequirement != player.ClassNone && p.Class != node.ClassRequirement

		style := lipgloss.NewStyle().Align(lipgloss.Center)
		
		if isSelected {
			style = style.Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("212"))
		} else if isUnlocked {
			style = style.Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("10")).Foreground(lipgloss.Color("10"))
		} else if isClassMismatch {
			style = style.Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("238")).Foreground(lipgloss.Color("242"))
		} else if isAvailable {
			style = style.Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("11")).Foreground(lipgloss.Color("11"))
		} else {
			style = style.Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("242")).Foreground(lipgloss.Color("242"))
		}
		
		contentWidth := runewidth.StringWidth(node.Icon)
		padding := (skillCellWidth - contentWidth) / 2
		boxContent := strings.Repeat(" ", padding) + node.Icon + strings.Repeat(" ", skillCellWidth-contentWidth-padding)
		
		grid[node.Position.Y][node.Position.X] = style.Render(boxContent)
	}
}

func createSkillInfoBox(cursorX, cursorY int, skillTree map[string]player.SkillNode, p *player.Player, width int) string {
	var infoBoxContent string
	selectedNode, ok := findNodeAt(cursorX, cursorY, skillTree)

	if ok {
		infoTitle := fmt.Sprintf("%s %s", selectedNode.Icon, selectedNode.Name)
		var status string
		if rpg.IsSkillUnlocked(p, selectedNode.ID) {
			status = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("[ИЗУЧЕНО]")
		} else if selectedNode.ClassRequirement != player.ClassNone && p.Class != selectedNode.ClassRequirement {
			status = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(fmt.Sprintf("[ТОЛЬКО ДЛЯ %s]", selectedNode.ClassRequirement))
		} else if rpg.IsSkillAvailable(p, selectedNode) {
			status = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render("[ДОСТУПНО]")
		} else {
			status = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("[ЗАБЛОКИРОВАНО]")
		}
		infoBoxContent = fmt.Sprintf("%s %s\n\n%s", infoTitle, status, selectedNode.Description)
	} else {
		infoBoxContent = "Выберите перк, чтобы увидеть информацию о нем."
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Width(width - 4).
		Render(infoBoxContent)
}

// --- Helpers ---

func findNodeAt(x, y int, tree map[string]player.SkillNode) (player.SkillNode, bool) {
	for _, node := range tree {
		if node.Position.X == x && node.Position.Y == y {
			return node, true
		}
	}
	return player.SkillNode{}, false
}

func mapToSlice(tree map[string]player.SkillNode) []player.SkillNode {
	nodes := make([]player.SkillNode, 0, len(tree))
	for _, node := range tree {
		nodes = append(nodes, node)
	}
	return nodes
}

func getGridChar(grid [][]string, x, y int) (string, bool) {
	if y < 0 || y >= len(grid) || x < 0 || x >= len(grid[0]) {
		return "", false
	}
	return grid[y][x], true
}

func setGridChar(grid [][]string, x, y int, char string) {
	if y < 0 || y >= len(grid) || x < 0 || x >= len(grid[0]) {
		return
	}
	grid[y][x] = padCell(char)
}

func padCell(s string) string {
	return s + strings.Repeat(" ", skillCellWidth-runewidth.StringWidth(s))
}

func distance(p1, p2 player.Position) float64 {
	return math.Sqrt(math.Pow(float64(p2.X-p1.X), 2) + math.Pow(float64(p2.Y-p1.Y), 2))
}
