package storage

import (
	"io/ioutil"
	"magus/player"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestSaveAndLoadQuests(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "quests_test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Temporarily override the quest file path for this test
	originalPath := QuestFile
	QuestFile = tmpfile.Name()
	defer func() { QuestFile = originalPath }()

	quests := []player.Quest{
		{ID: "q1", Title: "Test Quest 1", CreatedAt: time.Now()},
		{ID: "q2", Title: "Test Quest 2", CreatedAt: time.Now()},
	}

	err = SaveAllQuests(quests)
	if err != nil {
		t.Fatalf("SaveAllQuests() failed: %v", err)
	}

	// Test loading
	loadedQuests, err := LoadAllQuests()
	if err != nil {
		t.Fatalf("LoadAllQuests() failed: %v", err)
	}

	// Normalize time for comparison
	for i := range quests {
		quests[i].CreatedAt = quests[i].CreatedAt.Truncate(time.Second)
		loadedQuests[i].CreatedAt = loadedQuests[i].CreatedAt.Truncate(time.Second)
	}

	if !reflect.DeepEqual(quests, loadedQuests) {
		t.Errorf("Loaded quests do not match saved quests.\nSaved: %+v\nLoaded: %+v", quests, loadedQuests)
	}
}
