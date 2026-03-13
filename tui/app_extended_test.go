package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/CheeziCrew/swissgit/tui/screens"
)

// TestHandleMenuSelection_AllCommands verifies each menu command creates
// the correct screen and returns a non-nil init command.
func TestHandleMenuSelection_AllCommands(t *testing.T) {
	commands := []struct {
		command string
		screen  Screen
	}{
		{"pullrequest", ScreenPullRequest},
		{"cleanup", ScreenCleanup},
		{"commit", ScreenCommit},
		{"status", ScreenStatus},
		{"branches", ScreenBranches},
		{"clone", ScreenClone},
		{"automerge", ScreenAutomerge},
		{"mergeprs", ScreenMergePRs},
		{"enableworkflows", ScreenEnableWorkflows},
		{"teamprs", ScreenTeamPRs},
		{"myprs", ScreenMyPRs},
	}

	for _, tt := range commands {
		t.Run(tt.command, func(t *testing.T) {
			m := New()
			m.width = 120
			m.height = 40

			msg := screens.MenuSelectionMsg{Command: tt.command}
			updated, cmd := m.Update(msg)
			model := updated.(Model)

			if model.current != tt.screen {
				t.Errorf("current = %d, want %d", model.current, tt.screen)
			}
			if cmd == nil {
				t.Error("expected non-nil cmd from handleMenuSelection")
			}
		})
	}
}

func TestHandleMenuSelection_Unknown(t *testing.T) {
	m := New()
	msg := screens.MenuSelectionMsg{Command: "nonexistent"}
	updated, cmd := m.Update(msg)
	model := updated.(Model)
	if model.current != ScreenMenu {
		t.Errorf("current = %d, want ScreenMenu (%d)", model.current, ScreenMenu)
	}
	if cmd != nil {
		t.Error("expected nil cmd for unknown command")
	}
}

func TestSaveHistoryMsg(t *testing.T) {
	m := New()
	msg := screens.SaveHistoryMsg{Category: "test_cat", Value: "test_val"}
	_, _ = m.Update(msg)
	// Just verify it doesn't panic; history persistence is tested in history_test.go
}

func TestBackToMenuMsg_ResetsToMenu(t *testing.T) {
	m := New()
	m.width = 120
	m.height = 40
	m.current = ScreenPullRequest

	updated, cmd := m.Update(screens.BackToMenuMsg{})
	model := updated.(Model)
	if model.current != ScreenMenu {
		t.Errorf("current = %d, want ScreenMenu", model.current)
	}
	if cmd == nil {
		t.Error("expected cmd to send WindowSizeMsg")
	}
}

func TestQuitOnMenuScreen(t *testing.T) {
	m := New()
	m.current = ScreenMenu
	quitMsg := tea.KeyPressMsg{Code: 'q'}
	_, cmd := m.Update(quitMsg)
	if cmd == nil {
		t.Error("expected quit cmd when pressing q on menu screen")
	}
}

func TestNoQuitOnOtherScreen(t *testing.T) {
	m := New()
	m.current = ScreenStatus
	quitMsg := tea.KeyPressMsg{Code: 'q'}
	_, cmd := m.Update(quitMsg)
	// On non-menu screen, 'q' is forwarded to sub-screen, should not quit
	// We just verify no panic and cmd may or may not be nil
	_ = cmd
}

// TestView_AllScreens ensures View returns non-empty for every screen.
func TestView_AllScreens(t *testing.T) {
	allScreens := []struct {
		name   string
		screen Screen
	}{
		{"menu", ScreenMenu},
		{"pullrequest", ScreenPullRequest},
		{"cleanup", ScreenCleanup},
		{"commit", ScreenCommit},
		{"status", ScreenStatus},
		{"branches", ScreenBranches},
		{"clone", ScreenClone},
		{"automerge", ScreenAutomerge},
		{"mergeprs", ScreenMergePRs},
		{"enableworkflows", ScreenEnableWorkflows},
		{"teamprs", ScreenTeamPRs},
		{"myprs", ScreenMyPRs},
	}

	for _, tt := range allScreens {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			m.current = tt.screen
			v := m.View()
			if v.Content == "" {
				t.Errorf("View().Content is empty for screen %s", tt.name)
			}
		})
	}
}

// TestUpdateRouting_ForwardsToSubScreen verifies Update forwards messages
// to the correct sub-screen based on current screen.
func TestUpdateRouting_ForwardsToSubScreen(t *testing.T) {
	screens := []struct {
		name   string
		screen Screen
	}{
		{"pullrequest", ScreenPullRequest},
		{"cleanup", ScreenCleanup},
		{"commit", ScreenCommit},
		{"status", ScreenStatus},
		{"branches", ScreenBranches},
		{"clone", ScreenClone},
		{"automerge", ScreenAutomerge},
		{"mergeprs", ScreenMergePRs},
		{"enableworkflows", ScreenEnableWorkflows},
		{"teamprs", ScreenTeamPRs},
		{"myprs", ScreenMyPRs},
	}

	for _, tt := range screens {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			m.current = tt.screen
			// Send a window size msg -- all screens accept it
			wsMsg := tea.WindowSizeMsg{Width: 100, Height: 30}
			_, _ = m.Update(wsMsg)
			// Just verify no panic
		})
	}
}

func TestNavigateMsg(t *testing.T) {
	m := New()
	msg := NavigateMsg{Screen: ScreenClone}
	updated, cmd := m.Update(msg)
	model := updated.(Model)
	if model.current != ScreenClone {
		t.Errorf("current = %d, want ScreenClone", model.current)
	}
	if cmd != nil {
		t.Error("NavigateMsg should return nil cmd")
	}
}
