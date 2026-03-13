package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/CheeziCrew/swissgit/tui/screens"
)

// Test ScreenRepoSelect in View (covers uncovered branch in View)
func TestView_RepoSelectScreen(t *testing.T) {
	m := New()
	m.current = ScreenRepoSelect
	m.repoSelect = screens.NewRepoSelectModel("test", ".", 5, 30)
	v := m.View()
	if v.Content == "" {
		t.Error("View().Content is empty for ScreenRepoSelect")
	}
}

// Test Update routing to ScreenRepoSelect
func TestUpdateRouting_RepoSelect(t *testing.T) {
	m := New()
	m.current = ScreenRepoSelect
	m.repoSelect = screens.NewRepoSelectModel("test", ".", 5, 30)
	wsm := tea.WindowSizeMsg{Width: 80, Height: 24}
	_, _ = m.Update(wsm)
}

// Test NavigateMsg
func TestNavigateMsg_ToClone(t *testing.T) {
	m := New()
	msg := NavigateMsg{Screen: ScreenClone}
	updated, _ := m.Update(msg)
	model := updated.(Model)
	if model.current != ScreenClone {
		t.Errorf("current = %d, want ScreenClone", model.current)
	}
}

// Test Update with regular key on menu (not q)
func TestUpdate_RegularKeyOnMenu(t *testing.T) {
	m := New()
	m.current = ScreenMenu
	msg := tea.KeyPressMsg{Code: 'a'}
	_, _ = m.Update(msg)
}

// Test WindowSizeMsg stores dimensions
func TestUpdate_WindowSizeMsg(t *testing.T) {
	m := New()
	wsm := tea.WindowSizeMsg{Width: 100, Height: 50}
	updated, _ := m.Update(wsm)
	model := updated.(Model)
	if model.width != 100 {
		t.Errorf("width = %d, want 100", model.width)
	}
	if model.height != 50 {
		t.Errorf("height = %d, want 50", model.height)
	}
}
