package tui

import (
	"bytes"
	"fmt"
	"magus/player"
	"magus/rpg"
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dominikbraun/graph"
)

// --- Стили ---
var (
	styleSkillUnlocked    = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // Зеленый
	styleSkillAvailable   = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Желтый
	styleSkillLocked      = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // Красный
	styleSkillUnavailable = lipgloss.NewStyle().Foreground(lipgloss.Color("242")) // Серый
	styleInfoBox          = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(1, 2)
)

type skillView int

const (
	viewCommon skillView = iota
	viewClass
)

// Хеш-функция для player.SkillNode, необходимая для библиотеки graph.
// Она позволяет графу уникально идентифицировать каждую вершину по ее ID.
func skillNodeHash(s player.SkillNode) string {
	return s.ID
}

// SkillsState представляет собой состояние экрана дерева навыков.
type SkillsState struct {
	commonSkillGraph graph.Graph[string, player.SkillNode]
	classSkillGraph  graph.Graph[string, player.SkillNode]
	commonSkillIDs   []string
	classSkillIDs    []string
	currentView      skillView
	cursorIndex      int
	statusMessage    string
}

// NewSkillsState создает новое состояние экрана навыков.
func NewSkillsState(m *Model) State {
	s := &SkillsState{
		currentView: viewCommon,
		cursorIndex: 0,
	}

	trees, err := rpg.LoadSkillTrees(m.Player)
	if err != nil {
		s.statusMessage = "Ошибка загрузки дерева навыков."
		return s
	}

	s.commonSkillGraph, s.commonSkillIDs = s.buildGraph(trees.Common)
	s.classSkillGraph, s.classSkillIDs = s.buildGraph(trees.Class)

	return s
}

// buildGraph строит граф для библиотеки dominikbraun/graph
func (s *SkillsState) buildGraph(skillMap map[string]player.SkillNode) (graph.Graph[string, player.SkillNode], []string) {
	g := graph.New(skillNodeHash, graph.Directed(), graph.PreventCycles())

	for _, skill := range skillMap {
		_ = g.AddVertex(skill)
	}

	for _, skill := range skillMap {
		for _, reqID := range skill.Requirements {
			if strings.HasPrefix(reqID, "level_") {
				continue
			}
			if _, err := g.Vertex(reqID); err == nil {
				_ = g.AddEdge(reqID, skill.ID)
			}
		}
	}

	sortedIDs, _ := graph.TopologicalSort(g)
	
	sort.SliceStable(sortedIDs, func(i, j int) bool {
		skillA, _ := g.Vertex(sortedIDs[i])
		skillB, _ := g.Vertex(sortedIDs[j])
		if skillA.Position.Y != skillB.Position.Y {
			return skillA.Position.Y < skillB.Position.Y
		}
		return skillA.Position.X < skillB.Position.X
	})

	return g, sortedIDs
}

func (s *SkillsState) Init() tea.Cmd {
	return nil
}

func (s *SkillsState) Update(m *Model, msg tea.Msg) (State, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return s, nil
	}

	s.statusMessage = ""

	switch key.String() {
	case "q", "esc":
		return PopState{}, nil
	case "tab":
		s.switchView()
	case "up", "k":
		s.navigate(-1)
	case "down", "j":
		s.navigate(1)
	case "enter":
		s.unlockSkill(m)
	}

	return s, nil
}

func (s *SkillsState) View(m *Model) string {
	var b strings.Builder

	title := s.getViewTitle(m)
	b.WriteString(m.styles.TitleStyle.Render(title) + "\n")

	treeContent := s.renderTree(m)
	infoBox := s.renderInfoBox(m)

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, treeContent, infoBox)
	b.WriteString(mainContent + "\n\n")

	help := "Навигация: ↑↓ | Tab: сменить вид | Enter: изучить | q: назад"
	if s.statusMessage != "" {
		help = s.statusMessage
	}
	b.WriteString(m.styles.StatusMessageStyle.Render(help))

	return lipgloss.NewStyle().Margin(1, 2).Render(b.String())
}

// --- Логика ---

func (s *SkillsState) switchView() {
	if s.currentView == viewCommon {
		s.currentView = viewClass
	} else {
		s.currentView = viewCommon
	}
	s.cursorIndex = 0
}

func (s *SkillsState) navigate(delta int) {
	ids := s.getActiveIDs()
	if len(ids) == 0 {
		return
	}
	s.cursorIndex = (s.cursorIndex + delta + len(ids)) % len(ids)
}

func (s *SkillsState) unlockSkill(m *Model) {
	ids := s.getActiveIDs()
	if len(ids) == 0 || s.cursorIndex >= len(ids) {
		return
	}
	skillID := ids[s.cursorIndex]
	g := s.getActiveGraph()
	node, _ := g.Vertex(skillID)

	// Создаем плоскую карту всех навыков в текущем графе для IsSkillAvailable
	allSkillsInGraph := make(map[string]player.SkillNode)
	adjMap, _ := g.AdjacencyMap()
	for id := range adjMap {
		v, _ := g.Vertex(id)
		allSkillsInGraph[id] = v
	}

	if rpg.IsSkillUnlocked(m.Player, node.ID) {
		s.statusMessage = "✅ Навык уже изучен."
		return
	}

	if !rpg.IsSkillAvailable(m.Player, node, allSkillsInGraph) {
		if m.Player.SkillPoints <= 0 {
			s.statusMessage = "❗ Недостаточно очков навыков."
		} else {
			s.statusMessage = fmt.Sprintf("🔒 Требования для '%s' не выполнены.", node.Name)
		}
		return
	}

	if m.Player.SkillPoints <= 0 {
		s.statusMessage = "❗ Недостаточно очков навыков."
		return
	}

	m.Player.UnlockedSkills = append(m.Player.UnlockedSkills, node.ID)
	m.Player.SkillPoints--
	player.SavePlayer(m.Player)
	s.statusMessage = fmt.Sprintf("✨ Навык '%s' изучен!", node.Name)
}

// --- Хелперы для View ---

func (s *SkillsState) getViewTitle(m *Model) string {
	viewName := "Общие навыки"
	if s.currentView == viewClass {
		viewName = fmt.Sprintf("Навыки класса: %s", m.Player.Class)
	}
	return fmt.Sprintf("🧠 Дерево навыков (%s) | Очки: %d", viewName, m.Player.SkillPoints)
}

func (s *SkillsState) renderTree(m *Model) string {
	g := s.getActiveGraph()
	ids := s.getActiveIDs()
	if len(ids) == 0 {
		return styleInfoBox.Render("Нет доступных навыков в этой категории.")
	}

	var b bytes.Buffer
	
	var roots []string
	adjMap, _ := g.AdjacencyMap()
	for id := range adjMap {
		// Узел является корневым, если на него никто не ссылается.
		// Проверяем, есть ли он в качестве цели в каком-либо ребре.
		isRoot := true
		for _, edges := range adjMap {
			if _, ok := edges[id]; ok {
				isRoot = false
				break
			}
		}
		if isRoot {
			roots = append(roots, id)
		}
	}

	// Сортируем корневые узлы для стабильного отображения
	sort.Slice(roots, func(i, j int) bool {
		sA, _ := g.Vertex(roots[i])
		sB, _ := g.Vertex(roots[j])
		if sA.Position.Y != sB.Position.Y {
			return sA.Position.Y < sB.Position.Y
		}
		return sA.Position.X < sB.Position.X
	})


	for _, rootID := range roots {
		s.drawNode(&b, m, rootID, "", true)
	}

	return b.String()
}

func (s *SkillsState) drawNode(b *bytes.Buffer, m *Model, skillID, prefix string, isLast bool) {
	g := s.getActiveGraph()
	skill, _ := g.Vertex(skillID)

	isCursor := skillID == s.getActiveIDs()[s.cursorIndex]
	isUnlocked := rpg.IsSkillUnlocked(m.Player, skill.ID)
	
	allSkillsInGraph := make(map[string]player.SkillNode)
	adjMap, _ := g.AdjacencyMap()
	for id := range adjMap {
		v, _ := g.Vertex(id)
		allSkillsInGraph[id] = v
	}
	isAvailable := rpg.IsSkillAvailable(m.Player, skill, allSkillsInGraph)

	var style lipgloss.Style
	var icon string

	switch {
	case isUnlocked:
		style = styleSkillUnlocked
		icon = "✓"
	case isAvailable && m.Player.SkillPoints > 0:
		style = styleSkillAvailable
		icon = "+"
	case isAvailable && m.Player.SkillPoints <= 0:
		style = styleSkillLocked
		icon = "!"
	default:
		style = styleSkillUnavailable
		icon = " "
	}

	nodeStr := fmt.Sprintf("[%s] %s %s", icon, skill.Icon, skill.Name)
	if isCursor {
		nodeStr = lipgloss.NewStyle().Background(lipgloss.Color("237")).Render(nodeStr)
	} else {
		nodeStr = style.Render(nodeStr)
	}

	b.WriteString(prefix)
	if isLast {
		b.WriteString("└─ ")
		prefix += "   "
	} else {
		b.WriteString("├─ ")
		prefix += "│  "
	}
	b.WriteString(nodeStr)
	b.WriteString("\n")

	childrenMap := adjMap[skillID]
	
	var sortedChildren []string
	for childID := range childrenMap {
		sortedChildren = append(sortedChildren, childID)
	}
	
	sort.Slice(sortedChildren, func(i, j int) bool {
		sA, _ := g.Vertex(sortedChildren[i])
		sB, _ := g.Vertex(sortedChildren[j])
		if sA.Position.Y != sB.Position.Y {
			return sA.Position.Y < sB.Position.Y
		}
		return sA.Position.X < sB.Position.X
	})

	for i, childID := range sortedChildren {
		s.drawNode(b, m, childID, prefix, i == len(sortedChildren)-1)
	}
}

func (s *SkillsState) renderInfoBox(m *Model) string {
	ids := s.getActiveIDs()
	if len(ids) == 0 || s.cursorIndex >= len(ids) {
		return styleInfoBox.Width(50).Render("Выберите навык...")
	}
	skillID := ids[s.cursorIndex]
	g := s.getActiveGraph()
	node, _ := g.Vertex(skillID)

	title := fmt.Sprintf("%s %s", node.Icon, node.Name)

	var status, statusStyled string
	isUnlocked := rpg.IsSkillUnlocked(m.Player, node.ID)
	
	allSkillsInGraph := make(map[string]player.SkillNode)
	adjMap, _ := g.AdjacencyMap()
	for id := range adjMap {
		v, _ := g.Vertex(id)
		allSkillsInGraph[id] = v
	}
	isAvailable := rpg.IsSkillAvailable(m.Player, node, allSkillsInGraph)

	switch {
	case isUnlocked:
		status = "ИЗУЧЕНО"
		statusStyled = styleSkillUnlocked.Render(fmt.Sprintf("[%s]", status))
	case isAvailable && m.Player.SkillPoints > 0:
		status = "ДОСТУПНО"
		statusStyled = styleSkillAvailable.Render(fmt.Sprintf("[%s]", status))
	case isAvailable && m.Player.SkillPoints <= 0:
		status = "НЕ ХВАТАЕТ ОЧКОВ"
		statusStyled = styleSkillLocked.Render(fmt.Sprintf("[%s]", status))
	default:
		status = "ЗАБЛОКИРОВАНО"
		statusStyled = styleSkillUnavailable.Render(fmt.Sprintf("[%s]", status))
	}

	reqs := s.buildRequirementsString(node, m.Player)

	content := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, title, " ", statusStyled),
		"",
		node.Description,
		"",
		reqs,
	)

	return styleInfoBox.Width(50).Padding(1, 2).Render(content)
}

func (s *SkillsState) buildRequirementsString(node player.SkillNode, p *player.Player) string {
	if len(node.Requirements) == 0 {
		return "Требования: нет"
	}

	var reqs []string
	g := s.getActiveGraph()
	
	for _, reqID := range node.Requirements {
		var reqStr string
		var style lipgloss.Style

		if level, found := strings.CutPrefix(reqID, "level_"); found {
			reqLevel, _ := strconv.Atoi(level)
			reqStr = fmt.Sprintf("Уровень %s", level)
			if p.Level >= reqLevel {
				style = styleSkillUnlocked
			} else {
				style = styleSkillLocked
			}
		} else if reqNode, err := g.Vertex(reqID); err == nil {
			reqStr = reqNode.Name
			if rpg.IsSkillUnlocked(p, reqID) {
				style = styleSkillUnlocked
			} else {
				style = styleSkillLocked
			}
		} else {
			continue
		}

		reqs = append(reqs, style.Render(reqStr))
	}
	if len(reqs) == 0 {
		return "Требования: нет"
	}
	return "Требует: " + strings.Join(reqs, ", ")
}

// --- Вспомогательные функции ---

func (s *SkillsState) getActiveGraph() graph.Graph[string, player.SkillNode] {
	if s.currentView == viewClass {
		return s.classSkillGraph
	}
	return s.commonSkillGraph
}

func (s *SkillsState) getActiveIDs() []string {
	if s.currentView == viewClass {
		return s.classSkillIDs
	}
	return s.commonSkillIDs
}