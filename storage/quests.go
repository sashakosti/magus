package storage

import (
	"encoding/json"
	"io/ioutil"
	"magus/player"
	"os"
)

const questsFile = "data/quests.json"

// LoadAllQuests загружает все квесты из файла.
func LoadAllQuests() ([]player.Quest, error) {
	if _, err := os.Stat(questsFile); os.IsNotExist(err) {
		// Если файл не существует, возвращаем пустой список
		return []player.Quest{}, nil
	}

	file, err := ioutil.ReadFile(questsFile)
	if err != nil {
		return nil, err
	}

	var quests []player.Quest
	if err := json.Unmarshal(file, &quests); err != nil {
		return nil, err
	}

	return quests, nil
}

// SaveAllQuests сохраняет все квесты в файл.
func SaveAllQuests(quests []player.Quest) error {
	data, err := json.MarshalIndent(quests, "", "  ")
	if err != nil {
		return err
	}

	// Убедимся, что директория data существует
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		os.Mkdir("data", 0755)
	}

	return ioutil.WriteFile(questsFile, data, 0644)
}
