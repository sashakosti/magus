package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// RenderProgressBar создает строку с прогресс-баром.
func RenderProgressBar(current, max, width int) string {
	// Обеспечиваем, чтобы current не выходил за пределы [0, max]
	if current < 0 {
		current = 0
	}
	if current > max {
		current = max
	}

	percent := float64(current) / float64(max)
	filledWidth := int(percent * float64(width))

	// Выбираем цвет в зависимости от процента
	var barColor lipgloss.Color
	if percent > 0.5 {
		barColor = lipgloss.Color("10") // Зеленый
	} else if percent > 0.25 {
		barColor = lipgloss.Color("11") // Желтый
	} else {
		barColor = lipgloss.Color("9") // Красный
	}

	filled := strings.Repeat("█", filledWidth)
	empty := strings.Repeat("─", width-filledWidth)

	bar := filled + empty

	return lipgloss.NewStyle().Foreground(barColor).Render(bar)
}


func isToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}

func isYesterday(t, now time.Time) bool {
	yesterday := now.AddDate(0, 0, -1)
	return t.Year() == yesterday.Year() && t.Month() == yesterday.Month() && t.Day() == yesterday.Day()
}


func interpolate(a, b uint8, t float64) uint8 {
	return uint8(float64(a) + t*(float64(b)-float64(a)))
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func ansiGradient(text string, startRGB, endRGB [3]uint8) string {
	lines := strings.Split(text, "\n")
	n := len(lines)
	var builder strings.Builder

	for i, line := range lines {
		t := float64(i) / float64(Max(n-1, 1))
		r := interpolate(startRGB[0], endRGB[0], t)
		g := interpolate(startRGB[1], endRGB[1], t)
		b := interpolate(startRGB[2], endRGB[2], t)

		fmt.Fprintf(&builder, "\x1b[38;2;%d;%d;%dm%s\x1b[0m\n", r, g, b, line)
	}

	return builder.String()
}

func deadlineStatus(deadline *time.Time) string {
	if deadline == nil {
		return ""
	}
	remaining := time.Until(*deadline)
	if remaining < 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("(Просрочено)")
	}
	days := int(remaining.Hours() / 24)
	return fmt.Sprintf("(осталось %d д)", days)
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
