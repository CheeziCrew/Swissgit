package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const maxHistoryItems = 5

// History stores recent inputs per category (e.g. "commit_message", "pr_message").
type History struct {
	Messages map[string][]string `json:"messages"`
	path     string
}

func historyPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".swissgit", "history.json")
}

// LoadHistory reads the history file or returns an empty history.
func LoadHistory() *History {
	h := &History{
		Messages: make(map[string][]string),
		path:     historyPath(),
	}

	data, err := os.ReadFile(h.path)
	if err != nil {
		return h
	}

	_ = json.Unmarshal(data, h)
	if h.Messages == nil {
		h.Messages = make(map[string][]string)
	}
	return h
}

// Get returns the recent items for a category.
func (h *History) Get(category string) []string {
	return h.Messages[category]
}

// Add prepends an item to a category, deduplicates, and caps at maxHistoryItems.
func (h *History) Add(category, item string) {
	items := h.Messages[category]

	// Remove duplicates
	var filtered []string
	for _, existing := range items {
		if existing != item {
			filtered = append(filtered, existing)
		}
	}

	// Prepend new item
	result := append([]string{item}, filtered...)
	if len(result) > maxHistoryItems {
		result = result[:maxHistoryItems]
	}

	h.Messages[category] = result
	h.save()
}

func (h *History) save() {
	dir := filepath.Dir(h.path)
	_ = os.MkdirAll(dir, 0o755)

	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(h.path, data, 0o644)
}
