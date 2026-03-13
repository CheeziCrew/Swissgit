package screens

import (
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/CheeziCrew/swissgit/ops"
	"github.com/CheeziCrew/swissgit/tui/components"
)

// ws is a shorthand for window size message.
func ws() tea.WindowSizeMsg {
	return tea.WindowSizeMsg{Width: 120, Height: 40}
}

// --- Status: deeper Update paths ---

func TestStatusModel_UpdateReposDiscoveredEmpty(t *testing.T) {
	m := NewStatusModel()
	m, _ = m.Update(ws())
	m, _ = m.Update(statusReposDiscoveredMsg{paths: nil})
	if m.step != statusStepResults {
		t.Errorf("step = %d, want %d (should go to results on empty)", m.step, statusStepResults)
	}
}

func TestStatusModel_UpdateTaskDoneWithError(t *testing.T) {
	m := NewStatusModel()
	m, _ = m.Update(ws())
	m, _ = m.Update(statusReposDiscoveredMsg{paths: []string{"/repo1"}})

	result := ops.StatusResult{RepoName: "repo1", Error: "network error"}
	m, _ = m.Update(statusTaskDoneMsg{index: 0, result: result})

	if m.results[0].Error != "network error" {
		t.Errorf("expected error in result, got %q", m.results[0].Error)
	}
}

func TestStatusModel_UpdateResults_Esc(t *testing.T) {
	m := NewStatusModel()
	m, _ = m.Update(ws())
	m.step = statusStepResults
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from esc at results step")
	}
}

func TestStatusModel_UpdateResults_Q(t *testing.T) {
	m := NewStatusModel()
	m, _ = m.Update(ws())
	m.step = statusStepResults
	_, cmd := m.Update(qMsg())
	if cmd == nil {
		t.Error("expected cmd from q at results step")
	}
}

func TestStatusModel_UpdateProgressQ(t *testing.T) {
	m := NewStatusModel()
	m, _ = m.Update(ws())
	m.step = statusStepProgress
	_, cmd := m.Update(qMsg())
	if cmd == nil {
		t.Error("expected cmd from q at progress step")
	}
}

// --- Branches: deeper Update paths ---

func TestBranchesModel_ReposDiscoveredEmpty(t *testing.T) {
	m := NewBranchesModel()
	m, _ = m.Update(ws())
	m, _ = m.Update(branchesReposDiscoveredMsg{paths: nil})
	if m.step != branchesStepResults {
		t.Errorf("step = %d, want %d", m.step, branchesStepResults)
	}
}

func TestBranchesModel_TaskDone(t *testing.T) {
	m := NewBranchesModel()
	m, _ = m.Update(ws())
	m, _ = m.Update(branchesReposDiscoveredMsg{paths: []string{"/repo1"}})

	result := ops.BranchesResult{RepoName: "repo1", CurrentBranch: "main", DefaultBranch: "main"}
	m, _ = m.Update(branchesTaskDoneMsg{index: 0, result: result})
	if m.results[0].RepoName != "repo1" {
		t.Errorf("expected repo1, got %q", m.results[0].RepoName)
	}
}

func TestBranchesModel_TaskDoneWithError(t *testing.T) {
	m := NewBranchesModel()
	m, _ = m.Update(ws())
	m, _ = m.Update(branchesReposDiscoveredMsg{paths: []string{"/repo1"}})

	result := ops.BranchesResult{RepoName: "repo1", Error: "fetch error"}
	m, _ = m.Update(branchesTaskDoneMsg{index: 0, result: result})
	if m.results[0].Error == "" {
		t.Error("expected error in result")
	}
}

func TestBranchesModel_UpdateResults_Esc(t *testing.T) {
	m := NewBranchesModel()
	m, _ = m.Update(ws())
	m.step = branchesStepResults
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from esc at results step")
	}
}

// --- Cleanup: deeper Update paths ---

func TestCleanupModel_DropStep_EnterNo(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(ws())
	m.dropCursor = 0 // "No"
	m, _ = m.Update(enterMsg())
	if m.step != cleanupStepRepoSelect {
		t.Errorf("step = %d, want %d", m.step, cleanupStepRepoSelect)
	}
	if m.dropChanges {
		t.Error("expected dropChanges = false")
	}
}

func TestCleanupModel_DropStep_EnterYes(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(ws())
	m.dropCursor = 1 // "Yes"
	m, _ = m.Update(enterMsg())
	if m.step != cleanupStepRepoSelect {
		t.Errorf("step = %d, want %d", m.step, cleanupStepRepoSelect)
	}
	if !m.dropChanges {
		t.Error("expected dropChanges = true")
	}
}

func TestCleanupModel_TaskDoneSuccess(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(ws())
	m.step = cleanupStepProgress
	result := ops.CleanupResult{RepoName: "r", Success: true, PrunedBranches: 3, RemainingBranch: 1}
	m, _ = m.Update(cleanupTaskDoneMsg{index: 0, result: result})
	// verify no panic
}

func TestCleanupModel_TaskDoneFailed(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(ws())
	m.step = cleanupStepProgress
	result := ops.CleanupResult{RepoName: "r", Success: false, Error: "fail"}
	m, _ = m.Update(cleanupTaskDoneMsg{index: 0, result: result})
	// verify no panic
}

func TestCleanupModel_ResultsEsc(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(ws())
	m.step = cleanupStepResults
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from esc at results step")
	}
}

// --- Commit: deeper Update paths ---

func TestCommitModel_MessageEnterEmpty(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(ws())
	// Enter with empty message should stay at message step
	m, _ = m.Update(enterMsg())
	if m.step != commitStepMessage {
		t.Errorf("step = %d, want commitStepMessage", m.step)
	}
}

func TestCommitModel_BranchEscGoesBack(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(ws())
	m.step = commitStepBranch
	m, _ = m.Update(escMsg())
	if m.step != commitStepMessage {
		t.Errorf("step = %d, want commitStepMessage", m.step)
	}
}

func TestCommitModel_BranchEnter(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(ws())
	m.step = commitStepBranch
	m, cmd := m.Update(enterMsg())
	if m.step != commitStepRepoSelect {
		t.Errorf("step = %d, want commitStepRepoSelect", m.step)
	}
	_ = cmd
}

func TestCommitModel_HistoryBrowse(t *testing.T) {
	msgs := []string{"msg1", "msg2", "msg3"}
	m := NewCommitModel(msgs)
	m, _ = m.Update(ws())

	// Browse up
	m, _ = m.Update(upMsg())
	if m.historyCursor < 0 {
		t.Error("expected historyCursor >= 0 after up")
	}

	// Browse down
	m, _ = m.Update(downMsg())
}

func TestCommitModel_ResultsEsc(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(ws())
	m.step = commitStepResults
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from esc at results step")
	}
}

func TestCommitModel_ViewAllSteps(t *testing.T) {
	steps := []struct {
		name string
		step commitStep
	}{
		{"message", commitStepMessage},
		{"branch", commitStepBranch},
		{"reposelect", commitStepRepoSelect},
		{"progress", commitStepProgress},
		{"results", commitStepResults},
	}
	for _, tt := range steps {
		t.Run(tt.name, func(t *testing.T) {
			m := NewCommitModel(nil)
			m.viewReady = true
			m.step = tt.step
			v := m.View()
			if v == "" {
				t.Error("View() returned empty string")
			}
		})
	}
}

// --- PullRequest: deeper Update paths ---

func TestPRModel_MessageEnterEmpty(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(ws())
	m, _ = m.Update(enterMsg())
	if m.step != prStepMessage {
		t.Errorf("step = %d, want prStepMessage", m.step)
	}
}

func TestPRModel_BranchEscGoesBack(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(ws())
	m.step = prStepBranch
	m, _ = m.Update(escMsg())
	if m.step != prStepMessage {
		t.Errorf("step = %d, want prStepMessage", m.step)
	}
}

func TestPRModel_ResultsEsc(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(ws())
	m.step = prStepResults
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from esc at results")
	}
}

func TestPRModel_ViewAllSteps(t *testing.T) {
	steps := []struct {
		name string
		step prStep
	}{
		{"message", prStepMessage},
		{"branch", prStepBranch},
		{"changes", prStepChanges},
		{"breaking", prStepBreaking},
		{"reposelect", prStepRepoSelect},
		{"progress", prStepProgress},
		{"results", prStepResults},
	}
	for _, tt := range steps {
		t.Run(tt.name, func(t *testing.T) {
			m := NewPullRequestModel(nil)
			m.viewReady = true
			m.step = tt.step
			v := m.View()
			if v == "" {
				t.Error("View() returned empty string")
			}
		})
	}
}

func TestPRModel_HistoryBrowse(t *testing.T) {
	msgs := []string{"msg1", "msg2"}
	m := NewPullRequestModel(msgs)
	m, _ = m.Update(ws())

	m, _ = m.Update(upMsg())
	if m.historyCursor < 0 {
		t.Error("expected historyCursor >= 0")
	}
	m, _ = m.Update(downMsg())
}

// --- Clone: deeper Update paths ---

func TestCloneModel_InputTab(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(ws())
	initialFocus := m.focusIndex
	m, _ = m.Update(tabMsg())
	if m.focusIndex == initialFocus {
		t.Error("expected focusIndex to change after tab")
	}
}

func TestCloneModel_InputEsc(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(ws())
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from esc")
	}
}

func TestCloneModel_ViewAllSteps(t *testing.T) {
	steps := []struct {
		name string
		step cloneStep
	}{
		{"input", cloneStepInput},
		{"progress", cloneStepProgress},
		{"results", cloneStepResults},
	}
	for _, tt := range steps {
		t.Run(tt.name, func(t *testing.T) {
			m := NewCloneModel()
			m.viewReady = true
			m.step = tt.step
			v := m.View()
			if v == "" {
				t.Error("View() returned empty string")
			}
		})
	}
}

// --- Automerge: deeper Update paths ---

func TestAutomergeModel_TargetEsc(t *testing.T) {
	m := NewAutomergeModel()
	m, _ = m.Update(ws())
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from esc at target step")
	}
}

func TestAutomergeModel_TaskDone(t *testing.T) {
	m := NewAutomergeModel()
	m, _ = m.Update(ws())
	m.step = automergeStepProgress
	result := ops.AutomergeResult{RepoName: "r", Success: true}
	m, _ = m.Update(automergeTaskDoneMsg{index: 0, result: result})
	// verify no panic
}

func TestAutomergeModel_AllTasksDone(t *testing.T) {
	m := NewAutomergeModel()
	m, _ = m.Update(ws())
	m.step = automergeStepProgress
	m, _ = m.Update(components.AllTasksDoneMsg{})
	if m.step != automergeStepResults {
		t.Errorf("step = %d, want automergeStepResults", m.step)
	}
}

func TestAutomergeModel_ResultsEsc(t *testing.T) {
	m := NewAutomergeModel()
	m, _ = m.Update(ws())
	m.step = automergeStepResults
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from esc at results")
	}
}

func TestAutomergeModel_ViewAllSteps(t *testing.T) {
	steps := []struct {
		name string
		step automergeStep
	}{
		{"target", automergeStepTarget},
		{"progress", automergeStepProgress},
		{"results", automergeStepResults},
	}
	for _, tt := range steps {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAutomergeModel()
			m.viewReady = true
			m.step = tt.step
			v := m.View()
			if v == "" {
				t.Error("View() returned empty string")
			}
		})
	}
}

// --- MergePRs: deeper Update paths ---

func TestMergePRsModel_InputEsc(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(ws())
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from esc")
	}
}

func TestMergePRsModel_InputEmptyEnter(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(ws())
	// Enter with default value should proceed
	m, cmd := m.Update(enterMsg())
	_ = cmd
	// May or may not change step depending on input value
}

func TestMergePRsModel_ViewAllSteps(t *testing.T) {
	steps := []struct {
		name string
		step mergePRsStep
	}{
		{"input", mergePRsStepInput},
		{"fetching", mergePRsStepFetching},
		{"progress", mergePRsStepProgress},
		{"waiting", mergePRsStepWaiting},
		{"results", mergePRsStepResults},
	}
	for _, tt := range steps {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMergePRsModel()
			m.viewReady = true
			m.step = tt.step
			v := m.View()
			if v == "" {
				t.Error("View() returned empty string")
			}
		})
	}
}

// --- EnableWorkflows: deeper Update paths ---

func TestEnableWFModel_InputTab(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(ws())
	initialFocus := m.focusIndex
	m, _ = m.Update(tabMsg())
	if m.focusIndex == initialFocus {
		t.Error("expected focusIndex to change after tab")
	}
}

func TestEnableWFModel_InputEsc(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(ws())
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from esc")
	}
}

func TestEnableWFModel_ViewAllSteps(t *testing.T) {
	steps := []struct {
		name string
		step enableWFStep
	}{
		{"input", enableWFStepInput},
		{"fetching", enableWFStepFetching},
		{"progress", enableWFStepProgress},
		{"results", enableWFStepResults},
	}
	for _, tt := range steps {
		t.Run(tt.name, func(t *testing.T) {
			m := NewEnableWorkflowsModel()
			m.viewReady = true
			m.step = tt.step
			v := m.View()
			if v == "" {
				t.Error("View() returned empty string")
			}
		})
	}
}

// --- TeamPRs: deeper Update paths ---

func TestTeamPRsModel_InputEmptyEnter(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(ws())
	// Enter with default pre-filled values
	m, cmd := m.Update(enterMsg())
	_ = cmd
}

func TestTeamPRsModel_InputEsc(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(ws())
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from esc")
	}
}

func TestTeamPRsModel_FetchedPRsEmpty(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(ws())
	m.step = teamPRsStepFetching
	m, _ = m.Update(teamPRsFetchedMsg{prs: nil})
	if m.step != teamPRsStepResults {
		t.Errorf("step = %d, want teamPRsStepResults", m.step)
	}
}

func TestTeamPRsModel_FetchedPRsError(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(ws())
	m.step = teamPRsStepFetching
	m, _ = m.Update(teamPRsFetchedMsg{err: fmt.Errorf("network error")})
	v := m.View()
	if v == "" {
		t.Error("expected non-empty view after error")
	}
}

func TestTeamPRsModel_ResultsEsc(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(ws())
	m.step = teamPRsStepResults
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from esc at results")
	}
}

func TestTeamPRsModel_ViewAllSteps(t *testing.T) {
	steps := []struct {
		name string
		step teamPRsStep
	}{
		{"input", teamPRsStepInput},
		{"fetching", teamPRsStepFetching},
		{"results", teamPRsStepResults},
	}
	for _, tt := range steps {
		t.Run(tt.name, func(t *testing.T) {
			m := NewTeamPRsModel()
			m.viewReady = true
			m.step = tt.step
			v := m.View()
			if v == "" {
				t.Error("View() returned empty string")
			}
		})
	}
}

// --- MyPRs: deeper Update paths ---

func TestMyPRsModel_FetchedSuccess(t *testing.T) {
	m := NewMyPRsModel()
	m, _ = m.Update(ws())
	prs := []ops.MyPR{
		{Repo: "org/repo", Number: 1, Title: "fix"},
	}
	m, _ = m.Update(myPRsFetchedMsg{prs: prs})
	if m.step != myPRsStepResults {
		t.Errorf("step = %d, want myPRsStepResults", m.step)
	}
}

func TestMyPRsModel_FetchedError(t *testing.T) {
	m := NewMyPRsModel()
	m, _ = m.Update(ws())
	m, _ = m.Update(myPRsFetchedMsg{err: fmt.Errorf("auth error")})
	v := m.View()
	if v == "" {
		t.Error("expected non-empty view after error")
	}
}

func TestMyPRsModel_FetchingEsc(t *testing.T) {
	m := NewMyPRsModel()
	m, _ = m.Update(ws())
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from esc at fetching step")
	}
}

func TestMyPRsModel_ResultsEsc(t *testing.T) {
	m := NewMyPRsModel()
	m, _ = m.Update(ws())
	m.step = myPRsStepResults
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from esc at results")
	}
}

func TestMyPRsModel_ViewAllSteps(t *testing.T) {
	steps := []struct {
		name string
		step myPRsStep
	}{
		{"fetching", myPRsStepFetching},
		{"results", myPRsStepResults},
	}
	for _, tt := range steps {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMyPRsModel()
			m.viewReady = true
			m.step = tt.step
			v := m.View()
			if v == "" {
				t.Error("View() returned empty string")
			}
		})
	}
}
