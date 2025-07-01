package main

import (
	"fmt"
	"log"
	"magus/cmd"
	"magus/tui"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) < 2 {
		// Запускаем TUI, если нет команд
		p := tea.NewProgram(tui.InitialModel())
		if err := p.Start(); err != nil {
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
	case "version":
		cmd.Version()
	default:
		fmt.Println("Неизвестная команда:", os.Args[1])
	}
}
