package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/CheeziCrew/swissgit/tui/screens"
)

func TestNew(t *testing.T) {
	m := New()
	if m.current != ScreenMenu {
		t.Errorf("initial screen = %d, want ScreenMenu (%d)", m.current, ScreenMenu)
	}
}

func TestNewView(t *testing.T) {
	m := New()
	v := m.View()
	if v.Content == "" {
		t.Error("New().View().Content returned empty string")
	}
}

func TestNewInit(t *testing.T) {
	m := New()
	cmd := m.Init()
	if cmd != nil {
		t.Error("expected Init() to return nil cmd")
	}
}

func TestUpdateWindowSize(t *testing.T) {
	m := New()
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(msg)
	model := updated.(Model)
	if model.width != 120 {
		t.Errorf("width = %d, want 120", model.width)
	}
	if model.height != 40 {
		t.Errorf("height = %d, want 40", model.height)
	}
}

func TestUpdateNavigateMsg(t *testing.T) {
	m := New()
	msg := NavigateMsg{Screen: ScreenStatus}
	updated, _ := m.Update(msg)
	model := updated.(Model)
	if model.current != ScreenStatus {
		t.Errorf("current screen = %d, want ScreenStatus (%d)", model.current, ScreenStatus)
	}
}

func TestUpdateBackToMenu(t *testing.T) {
	m := New()
	// Navigate away first.
	m.current = ScreenStatus
	updated, _ := m.Update(screens.BackToMenuMsg{})
	model := updated.(Model)
	if model.current != ScreenMenu {
		t.Errorf("current screen = %d, want ScreenMenu (%d)", model.current, ScreenMenu)
	}
}
