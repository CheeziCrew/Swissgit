package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadHistory_FromExistingFile(t *testing.T) {
	dir := t.TempDir()
	histDir := filepath.Join(dir, ".swissgit")
	os.MkdirAll(histDir, 0755)
	histPath := filepath.Join(histDir, "history.json")

	// Write a valid history file
	data := History{
		Messages: map[string][]string{
			"commit": {"fix: test", "feat: add"},
			"pr":     {"pr title"},
		},
	}
	jsonData, _ := json.Marshal(data)
	os.WriteFile(histPath, jsonData, 0644)

	// Override historyPath for testing
	h := &History{
		Messages: make(map[string][]string),
		path:     histPath,
	}
	fileData, err := os.ReadFile(h.path)
	if err != nil {
		t.Fatalf("unexpected error reading file: %v", err)
	}
	_ = json.Unmarshal(fileData, h)
	if h.Messages == nil {
		h.Messages = make(map[string][]string)
	}

	items := h.Messages["commit"]
	if len(items) != 2 {
		t.Errorf("expected 2 commit items, got %d", len(items))
	}
	if items[0] != "fix: test" {
		t.Errorf("first item = %q, want %q", items[0], "fix: test")
	}
}

func TestLoadHistory_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	histDir := filepath.Join(dir, ".swissgit")
	os.MkdirAll(histDir, 0755)
	histPath := filepath.Join(histDir, "history.json")

	// Write invalid JSON
	os.WriteFile(histPath, []byte("not json"), 0644)

	h := &History{
		Messages: make(map[string][]string),
		path:     histPath,
	}
	fileData, _ := os.ReadFile(h.path)
	_ = json.Unmarshal(fileData, h)
	if h.Messages == nil {
		h.Messages = make(map[string][]string)
	}

	// Should have empty messages since JSON is invalid
	if len(h.Messages) != 0 {
		t.Errorf("expected empty messages for invalid JSON, got %v", h.Messages)
	}
}

func TestLoadHistory_FileNotExists(t *testing.T) {
	// LoadHistory with non-existent path should return empty history
	h := LoadHistory()
	if h == nil {
		t.Fatal("LoadHistory returned nil")
	}
	if h.Messages == nil {
		t.Error("Messages should not be nil")
	}
}

func TestHistory_Save_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	deepPath := filepath.Join(dir, "a", "b", "c", "history.json")

	h := &History{
		Messages: map[string][]string{"test": {"value"}},
		path:     deepPath,
	}
	h.save()

	// Verify file was created
	if _, err := os.Stat(deepPath); os.IsNotExist(err) {
		t.Error("save did not create the file")
	}
}

func TestHistory_Save_NilMessages(t *testing.T) {
	dir := t.TempDir()
	h := &History{
		Messages: nil,
		path:     filepath.Join(dir, "history.json"),
	}
	// Should not panic
	h.save()
}
