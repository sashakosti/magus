package main

import (
	"fmt"
	"log"
	"magus/cmd"
	"magus/tui"
	"math/rand"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	if len(os.Args) < 2 {
		// Запускаем TUI, если нет команд
		m := tui.InitialModel()
		p := tea.NewProgram(&m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatalf("Ошибка при запуске TUI: %v", err)
		}
		os.Exit(0)
	}

	switch os.Args[1] {
	case "add":
		cmd.Add()
	case "list":
		cmd.List()
	case "show":
		cmd.Show()
	case "why":
		cmd.Why()
	case "complete":
		cmd.Complete()
	case "roadmap":
		cmd.Roadmap()
	case "version":
		cmd.Version()
	default:
		fmt.Println("Неизвестная команда:", os.Args[1])
	}
}
