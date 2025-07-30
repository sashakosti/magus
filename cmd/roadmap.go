package cmd

import (
	"fmt"
	"magus/player"
	"magus/storage"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
)

func findSubQuests(questID string, allQuests []player.Quest) []player.Quest {
	var subQuests []player.Quest
	for _, q := range allQuests {
		if q.ParentID == questID {
			subQuests = append(subQuests, q)
			// Рекурсивно ищем под-квесты
			subQuests = append(subQuests, findSubQuests(q.ID, allQuests)...)
		}
	}
	return subQuests
}

func Roadmap() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: magus roadmap <quest_id>")
		return
	}
	questID := os.Args[2]

	allQuests, err := storage.LoadAllQuests()
	if err != nil {
		fmt.Println("❌ Ошибка загрузки квестов:", err)
		return
	}

	var targetQuest *player.Quest
	for i, q := range allQuests {
		if q.ID == questID {
			targetQuest = &allQuests[i]
			break
		}
	}

	if targetQuest == nil {
		fmt.Println("⚠️ Квест с таким ID не найден.")
		return
	}

	subQuests := findSubQuests(targetQuest.ID, allQuests)
	allRelatedQuests := append([]player.Quest{*targetQuest}, subQuests...)

	var completedCount int
	for _, q := range allRelatedQuests {
		if q.Completed {
			completedCount++
		}
	}

	totalCount := len(allRelatedQuests)
	progressPercentage := 0.0
	if totalCount > 0 {
		progressPercentage = float64(completedCount) / float64(totalCount)
	}

	// --- Визуализация ---
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	fmt.Println(titleStyle.Render(fmt.Sprintf("🗺️ Роадмап для цели: %s", targetQuest.Title)))

	// Прогресс-бар
	p := progress.New(progress.WithDefaultGradient())
	progressView := p.ViewAs(progressPercentage)
	fmt.Printf("Прогресс: %s %.0f%%\n\n", progressView, progressPercentage*100)

	// Список задач
	fmt.Println("Подзадачи:")
	for _, q := range allRelatedQuests {
		// status := "⏳"
		style := lipgloss.NewStyle()
		if q.Completed {
			// status = "✅"
			style = style.Strikethrough(true).Faint(true)
		}

		indent := ""
		if q.ParentID != "" {
			// Простой отступ для наглядности
			indent = strings.Repeat("  ", strings.Count(q.ParentID, "")-1) + "└─ "
		}

		fmt.Println(style.Render(fmt.Sprintf("%s%s", indent, q.Title)))
	}
}
