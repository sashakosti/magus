package storage

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"
)

type ReflectionNote struct {
	Date     time.Time     `json:"date"`
	Duration time.Duration `json:"duration"`
	Content  string        `json:"content"`
	XPEarned int           `json:"xp_earned"`
	HPLoss   int           `json:"hp_loss"`
}

const reflectionsFile = "data/reflections.json"

func SaveReflection(note ReflectionNote) error {
	notes, err := loadReflections()
	if err != nil {
		notes = []ReflectionNote{} // Если файл не существует или пуст, создаем новый срез
	}

	notes = append(notes, note)

	data, err := json.MarshalIndent(notes, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(reflectionsFile, data, 0644)
}

func loadReflections() ([]ReflectionNote, error) {
	if _, err := os.Stat(reflectionsFile); os.IsNotExist(err) {
		return []ReflectionNote{}, nil
	}

	data, err := ioutil.ReadFile(reflectionsFile)
	if err != nil {
		return nil, err
	}

	var notes []ReflectionNote
	err = json.Unmarshal(data, &notes)
	return notes, err
}
