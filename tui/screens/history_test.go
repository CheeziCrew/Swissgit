package screens

import (
	"testing"

	"charm.land/bubbles/v2/textinput"
)

func newTestInput(val string) textinput.Model {
	ti := textinput.New()
	ti.SetValue(val)
	return ti
}

func TestBrowseUp_SavesTypedValue(t *testing.T) {
	h := NewHistoryBrowser([]string{"old1", "old2"})
	input := newTestInput("my typing")

	h.BrowseUp(&input)

	if h.typedValue != "my typing" {
		t.Errorf("typedValue = %q, want %q", h.typedValue, "my typing")
	}
	if input.Value() != "old1" {
		t.Errorf("input = %q, want %q", input.Value(), "old1")
	}
	if h.cursor != 0 {
		t.Errorf("cursor = %d, want 0", h.cursor)
	}
}

func TestBrowseUp_ClampsAtEnd(t *testing.T) {
	h := NewHistoryBrowser([]string{"a", "b"})
	input := newTestInput("")

	h.BrowseUp(&input) // cursor=0
	h.BrowseUp(&input) // cursor=1
	h.BrowseUp(&input) // should stay at 1

	if h.cursor != 1 {
		t.Errorf("cursor = %d, want 1 (clamped)", h.cursor)
	}
	if input.Value() != "b" {
		t.Errorf("input = %q, want %q", input.Value(), "b")
	}
}

func TestBrowseDown_RestoresTypedValue(t *testing.T) {
	h := NewHistoryBrowser([]string{"old1", "old2"})
	input := newTestInput("original")

	h.BrowseUp(&input) // saves "original", shows "old1"
	h.BrowseUp(&input) // shows "old2"
	h.BrowseDown(&input) // shows "old1"

	if input.Value() != "old1" {
		t.Errorf("input = %q, want %q", input.Value(), "old1")
	}

	h.BrowseDown(&input) // restores "original"

	if input.Value() != "original" {
		t.Errorf("input = %q, want %q", input.Value(), "original")
	}
	if h.cursor != -1 {
		t.Errorf("cursor = %d, want -1", h.cursor)
	}
}

func TestBrowseDown_WhenNotActive(t *testing.T) {
	h := NewHistoryBrowser([]string{"a"})
	input := newTestInput("untouched")

	h.BrowseDown(&input)

	if input.Value() != "untouched" {
		t.Errorf("input changed to %q, should stay %q", input.Value(), "untouched")
	}
	if h.cursor != -1 {
		t.Errorf("cursor = %d, want -1", h.cursor)
	}
}

func TestReset(t *testing.T) {
	h := NewHistoryBrowser([]string{"a"})
	input := newTestInput("")

	h.BrowseUp(&input)
	if !h.IsActive() {
		t.Error("expected active after BrowseUp")
	}

	h.Reset()
	if h.IsActive() {
		t.Error("expected inactive after Reset")
	}
}

func TestEmptyHistory(t *testing.T) {
	h := NewHistoryBrowser(nil)
	input := newTestInput("typed")

	h.BrowseUp(&input)

	if input.Value() != "typed" {
		t.Errorf("input changed to %q, should stay %q", input.Value(), "typed")
	}
	if h.IsActive() {
		t.Error("expected inactive with empty history")
	}
	if h.Len() != 0 {
		t.Errorf("Len() = %d, want 0", h.Len())
	}
}

func TestLen(t *testing.T) {
	h := NewHistoryBrowser([]string{"a", "b", "c"})
	if h.Len() != 3 {
		t.Errorf("Len() = %d, want 3", h.Len())
	}
}
