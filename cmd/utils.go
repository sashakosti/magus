package cmd

import (
	"fmt"
	"time"
)

func Version() {
	fmt.Println("🧙 Magus v0.1.0")
}

// isToday проверяет, является ли дата сегодняшней.
func isToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}
