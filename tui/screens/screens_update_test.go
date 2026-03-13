package screens

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/CheeziCrew/swissgit/ops"
	"github.com/CheeziCrew/swissgit/tui/components"
)

// Helper message constructors
func escMsg() tea.KeyPressMsg  { return tea.KeyPressMsg{Code: tea.KeyEscape} }
func enterMsg() tea.KeyPressMsg { return tea.KeyPressMsg{Code: tea.KeyEnter} }
func upMsg() tea.KeyPressMsg   { return tea.KeyPressMsg{Code: tea.KeyUp} }
func downMsg() tea.KeyPressMsg { return tea.KeyPressMsg{Code: tea.KeyDown} }
func tabMsg() tea.KeyPressMsg  { return tea.KeyPressMsg{Code: tea.KeyTab} }
func spaceMsg() tea.KeyPressMsg { return tea.KeyPressMsg{Code: tea.KeySpace} }
func qMsg() tea.KeyPressMsg    { return tea.KeyPressMsg{Code: 'q'} }
func wsMsg() tea.WindowSizeMsg { return tea.WindowSizeMsg{Width: 120, Height: 40} }

// --- Status ---

func TestStatusModel_UpdateWindowSize(t *testing.T) {
	m := NewStatusModel()
	m, _ = m.Update(wsMsg())
	if !m.viewReady {
		t.Error("expected viewReady = true")
	}
}

func TestStatusModel_ReposDiscovered(t *testing.T) {
	m := NewStatusModel()
	m, _ = m.Update(wsMsg())
	m, _ = m.Update(statusReposDiscoveredMsg{paths: []string{"/repo1", "/repo2"}})
	if m.step != statusStepProgress {
		t.Errorf("step = %d, want %d", m.step, statusStepProgress)
	}
}

func TestStatusModel_TaskDone(t *testing.T) {
	m := NewStatusModel()
	m, _ = m.Update(wsMsg())
	m, _ = m.Update(statusReposDiscoveredMsg{paths: []string{"/repo1"}})
	result := ops.StatusResult{RepoName: "repo1", Branch: "main", Clean: true}
	m, _ = m.Update(statusTaskDoneMsg{index: 0, result: result})
	if len(m.results) < 1 {
		t.Error("expected results to have entries")
	}
}

func TestStatusModel_EscReturnsToMenu(t *testing.T) {
	m := NewStatusModel()
	m, _ = m.Update(wsMsg())
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from escape key")
	}
}

func TestStatusModel_AllTasksDone(t *testing.T) {
	m := NewStatusModel()
	m, _ = m.Update(wsMsg())
	m.step = statusStepProgress
	m, _ = m.Update(components.AllTasksDoneMsg{})
	if m.step != statusStepResults {
		t.Errorf("step = %d, want %d", m.step, statusStepResults)
	}
}

func TestStatusModel_View(t *testing.T) {
	m := NewStatusModel()
	m.viewReady = true
	v := m.View()
	if v == "" {
		t.Error("View() empty")
	}
}

func TestStatusModel_ViewResults(t *testing.T) {
	m := NewStatusModel()
	m.viewReady = true
	m.step = statusStepResults
	m.results = []ops.StatusResult{
		{RepoName: "test", Branch: "main", Clean: true},
	}
	v := m.View()
	if v == "" {
		t.Error("View() empty at results step")
	}
}

// --- Branches ---

func TestBranchesModel_UpdateWindowSize(t *testing.T) {
	m := NewBranchesModel()
	m, _ = m.Update(wsMsg())
	if !m.viewReady {
		t.Error("expected viewReady = true")
	}
}

func TestBranchesModel_ReposDiscovered(t *testing.T) {
	m := NewBranchesModel()
	m, _ = m.Update(wsMsg())
	m, _ = m.Update(branchesReposDiscoveredMsg{paths: []string{"/repo1"}})
	if m.step != branchesStepProgress {
		t.Errorf("step = %d, want %d", m.step, branchesStepProgress)
	}
}

func TestBranchesModel_AllTasksDone(t *testing.T) {
	m := NewBranchesModel()
	m, _ = m.Update(wsMsg())
	m.step = branchesStepProgress
	m, _ = m.Update(components.AllTasksDoneMsg{})
	if m.step != branchesStepResults {
		t.Errorf("step = %d, want %d", m.step, branchesStepResults)
	}
}

func TestBranchesModel_EscReturnsToMenu(t *testing.T) {
	m := NewBranchesModel()
	m, _ = m.Update(wsMsg())
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from escape")
	}
}

func TestBranchesModel_View(t *testing.T) {
	m := NewBranchesModel()
	m.viewReady = true
	v := m.View()
	if v == "" {
		t.Error("View() empty")
	}
}

// --- Cleanup ---

func TestCleanupModel_UpdateWindowSize(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	if !m.viewReady {
		t.Error("expected viewReady = true")
	}
}

func TestCleanupModel_DropStep_Navigation(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	// Navigate down/up in drop selection
	m, _ = m.Update(downMsg())
	if m.dropCursor != 1 {
		t.Errorf("dropCursor = %d, want 1", m.dropCursor)
	}
	m, _ = m.Update(upMsg())
	if m.dropCursor != 0 {
		t.Errorf("dropCursor = %d, want 0", m.dropCursor)
	}
}

func TestCleanupModel_TaskDone(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	m.step = cleanupStepProgress
	result := ops.CleanupResult{RepoName: "test", Success: true, PrunedBranches: 2}
	m, _ = m.Update(cleanupTaskDoneMsg{index: 0, result: result})
	// Results are collected in the results model
}

func TestCleanupModel_AllTasksDone(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	m.step = cleanupStepProgress
	m, _ = m.Update(components.AllTasksDoneMsg{})
	if m.step != cleanupStepResults {
		t.Errorf("step = %d, want %d", m.step, cleanupStepResults)
	}
}

func TestCleanupModel_EscReturnsToMenu(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from escape")
	}
}

func TestCleanupModel_View(t *testing.T) {
	m := NewCleanupModel()
	m.viewReady = true
	v := m.View()
	if v == "" {
		t.Error("View() empty")
	}
}

// --- Commit ---

func TestCommitModel_UpdateWindowSize(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	if !m.viewReady {
		t.Error("expected viewReady = true")
	}
}

func TestCommitModel_EscReturnsToMenu(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from escape")
	}
}

func TestCommitModel_View(t *testing.T) {
	m := NewCommitModel(nil)
	m.viewReady = true
	v := m.View()
	if v == "" {
		t.Error("View() empty")
	}
}

// --- PullRequest ---

func TestPRModel_UpdateWindowSize(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	if !m.viewReady {
		t.Error("expected viewReady = true")
	}
}

func TestPRModel_EscReturnsToMenu(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from escape")
	}
}

func TestPRModel_View(t *testing.T) {
	m := NewPullRequestModel(nil)
	m.viewReady = true
	v := m.View()
	if v == "" {
		t.Error("View() empty")
	}
}

// --- Clone ---

func TestCloneModel_UpdateWindowSize(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	if !m.viewReady {
		t.Error("expected viewReady = true")
	}
}

func TestCloneModel_EscReturnsToMenu(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from escape")
	}
}

func TestCloneModel_View(t *testing.T) {
	m := NewCloneModel()
	m.viewReady = true
	v := m.View()
	if v == "" {
		t.Error("View() empty")
	}
}

// --- Automerge ---

func TestAutomergeModel_UpdateWindowSize(t *testing.T) {
	m := NewAutomergeModel()
	m, _ = m.Update(wsMsg())
	if !m.viewReady {
		t.Error("expected viewReady = true")
	}
}

func TestAutomergeModel_EscReturnsToMenu(t *testing.T) {
	m := NewAutomergeModel()
	m, _ = m.Update(wsMsg())
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from escape")
	}
}

func TestAutomergeModel_View(t *testing.T) {
	m := NewAutomergeModel()
	m.viewReady = true
	v := m.View()
	if v == "" {
		t.Error("View() empty")
	}
}

// --- MergePRs ---

func TestMergePRsModel_UpdateWindowSize(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	if !m.viewReady {
		t.Error("expected viewReady = true")
	}
}

func TestMergePRsModel_EscReturnsToMenu(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from escape")
	}
}

func TestMergePRsModel_View(t *testing.T) {
	m := NewMergePRsModel()
	m.viewReady = true
	v := m.View()
	if v == "" {
		t.Error("View() empty")
	}
}

// --- EnableWorkflows ---

func TestEnableWFModel_UpdateWindowSize(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	if !m.viewReady {
		t.Error("expected viewReady = true")
	}
}

func TestEnableWFModel_View(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m.viewReady = true
	v := m.View()
	if v == "" {
		t.Error("View() empty")
	}
}

// --- TeamPRs ---

func TestTeamPRsModel_UpdateWindowSize(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	if !m.viewReady {
		t.Error("expected viewReady = true")
	}
}

func TestTeamPRsModel_EscReturnsToMenu(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from escape")
	}
}

func TestTeamPRsModel_View(t *testing.T) {
	m := NewTeamPRsModel()
	m.viewReady = true
	v := m.View()
	if v == "" {
		t.Error("View() empty")
	}
}

// --- MyPRs ---

func TestMyPRsModel_UpdateWindowSize(t *testing.T) {
	m := NewMyPRsModel()
	m, _ = m.Update(wsMsg())
	if !m.viewReady {
		t.Error("expected viewReady = true")
	}
}

func TestMyPRsModel_EscReturnsToMenu(t *testing.T) {
	m := NewMyPRsModel()
	m, _ = m.Update(wsMsg())
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from escape")
	}
}

func TestMyPRsModel_View(t *testing.T) {
	m := NewMyPRsModel()
	m.viewReady = true
	v := m.View()
	if v == "" {
		t.Error("View() empty")
	}
}
