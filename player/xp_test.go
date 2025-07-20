package player

import (
	"os"
	"testing"
)

// TestMain позволяет нам управлять setup и teardown для всех тестов в пакете.
func TestMain(m *testing.M) {
	// Setup: используем временный файл для тестов
	tmpfile, err := os.CreateTemp("", "player_test_*.json")
	if err != nil {
		panic("Failed to create temp file for testing")
	}
	originalPlayerFile := PlayerFile
	PlayerFile = tmpfile.Name()

	// Запускаем тесты
	code := m.Run()

	// Teardown: очищаем временный файл и восстанавливаем путь
	os.Remove(PlayerFile)
	PlayerFile = originalPlayerFile

	os.Exit(code)
}

func TestAddXP(t *testing.T) {
	// Создаем чистого игрока для этого теста
	p, err := CreatePlayer("TestXP")
	if err != nil {
		t.Fatalf("CreatePlayer failed: %v", err)
	}
	SavePlayer(p)

	ready, err := AddXP(50)
	if err != nil {
		t.Fatalf("AddXP failed: %v", err)
	}
	if ready {
		t.Errorf("Player should not be ready to level up, but they are")
	}

	loadedPlayer, _ := LoadPlayer()
	if loadedPlayer.XP != 50 {
		t.Errorf("Expected XP to be 50, but got %d", loadedPlayer.XP)
	}

	ready, err = AddXP(60)
	if err != nil {
		t.Fatalf("AddXP failed: %v", err)
	}
	if !ready {
		t.Errorf("Player should be ready to level up, but they are not")
	}

	loadedPlayer, _ = LoadPlayer()
	if loadedPlayer.XP != 110 {
		t.Errorf("Expected XP to be 110, but got %d", loadedPlayer.XP)
	}
}
