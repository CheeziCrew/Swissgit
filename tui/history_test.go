package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func newTestHistory(t *testing.T) *History {
	t.Helper()
	dir := t.TempDir()
	return &History{
		Messages: make(map[string][]string),
		path:     filepath.Join(dir, ".swissgit", "history.json"),
	}
}

func TestHistory_Add_Basic(t *testing.T) {
	h := newTestHistory(t)
	h.Add("commit", "initial commit")

	items := h.Get("commit")
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0] != "initial commit" {
		t.Errorf("got %q, want %q", items[0], "initial commit")
	}
}

func TestHistory_Add_Prepends(t *testing.T) {
	h := newTestHistory(t)
	h.Add("commit", "first")
	h.Add("commit", "second")

	items := h.Get("commit")
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0] != "second" {
		t.Errorf("first item: got %q, want %q", items[0], "second")
	}
	if items[1] != "first" {
		t.Errorf("second item: got %q, want %q", items[1], "first")
	}
}

func TestHistory_Add_Dedup(t *testing.T) {
	h := newTestHistory(t)
	h.Add("commit", "fix bug")
	h.Add("commit", "add feature")
	h.Add("commit", "fix bug") // duplicate, should move to front

	items := h.Get("commit")
	if len(items) != 2 {
		t.Fatalf("expected 2 items after dedup, got %d: %v", len(items), items)
	}
	if items[0] != "fix bug" {
		t.Errorf("first item: got %q, want %q", items[0], "fix bug")
	}
	if items[1] != "add feature" {
		t.Errorf("second item: got %q, want %q", items[1], "add feature")
	}
}

func TestHistory_Add_CapsAtMax(t *testing.T) {
	h := newTestHistory(t)
	for i := 0; i < 10; i++ {
		h.Add("commit", string(rune('a'+i)))
	}

	items := h.Get("commit")
	if len(items) != maxHistoryItems {
		t.Errorf("expected %d items (max), got %d", maxHistoryItems, len(items))
	}
	// Most recent should be first
	if items[0] != "j" {
		t.Errorf("first item: got %q, want %q", items[0], "j")
	}
}

func TestHistory_Get_EmptyCategory(t *testing.T) {
	h := newTestHistory(t)
	items := h.Get("nonexistent")
	if items != nil {
		t.Errorf("expected nil for empty category, got %v", items)
	}
}

func TestHistory_Get_MultipleCategories(t *testing.T) {
	h := newTestHistory(t)
	h.Add("commit", "commit msg")
	h.Add("pr", "pr title")

	commits := h.Get("commit")
	prs := h.Get("pr")

	if len(commits) != 1 || commits[0] != "commit msg" {
		t.Errorf("commits: %v", commits)
	}
	if len(prs) != 1 || prs[0] != "pr title" {
		t.Errorf("prs: %v", prs)
	}
}

func TestHistory_Persistence(t *testing.T) {
	dir := t.TempDir()
	histPath := filepath.Join(dir, ".swissgit", "history.json")

	// Create and save
	h := &History{
		Messages: make(map[string][]string),
		path:     histPath,
	}
	h.Add("commit", "persisted msg")

	// Read back from file
	data, err := os.ReadFile(histPath)
	if err != nil {
		t.Fatalf("failed to read history file: %v", err)
	}

	var loaded History
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	items := loaded.Messages["commit"]
	if len(items) != 1 || items[0] != "persisted msg" {
		t.Errorf("persisted data mismatch: %v", items)
	}
}
