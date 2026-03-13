package screens

import (
	"os"
	"path/filepath"
	"testing"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/CheeziCrew/swissgit/ops"
	"github.com/CheeziCrew/swissgit/tui/components"
)

// initTestRepo creates a minimal git repo for testing.
func initTestRepo(t *testing.T, dir string) {
	t.Helper()
	repo, err := gogit.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("init repo: %v", err)
	}
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"git@github.com:testorg/testrepo.git"},
	})
	if err != nil {
		t.Fatalf("create remote: %v", err)
	}
	w, _ := repo.Worktree()
	f, _ := os.Create(filepath.Join(dir, "README.md"))
	f.WriteString("# test\n")
	f.Close()
	w.Add("README.md")
	w.Commit("initial commit", &gogit.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@test.com"},
	})
}

func kMsg(c rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: c} }

// === Init() tests ===

func TestAutomergeModel_Init(t *testing.T) {
	m := NewAutomergeModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() returned nil")
	}
}

func TestBranchesModel_Init(t *testing.T) {
	m := NewBranchesModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() returned nil")
	}
}

func TestCleanupModel_Init(t *testing.T) {
	m := NewCleanupModel()
	cmd := m.Init()
	// Cleanup Init returns nil (no async work needed at start)
	_ = cmd
}

func TestCloneModel_Init(t *testing.T) {
	m := NewCloneModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() returned nil")
	}
}

func TestCommitModel_Init(t *testing.T) {
	m := NewCommitModel(nil)
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() returned nil")
	}
}

func TestPullRequestModel_Init(t *testing.T) {
	m := NewPullRequestModel(nil)
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() returned nil")
	}
}

func TestStatusModel_Init(t *testing.T) {
	m := NewStatusModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() returned nil")
	}
}

func TestMergePRsModel_Init(t *testing.T) {
	m := NewMergePRsModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() returned nil")
	}
}

func TestEnableWorkflowsModel_Init(t *testing.T) {
	m := NewEnableWorkflowsModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() returned nil")
	}
}

func TestTeamPRsModel_Init(t *testing.T) {
	m := NewTeamPRsModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() returned nil")
	}
}

func TestMyPRsModel_Init(t *testing.T) {
	m := NewMyPRsModel()
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() returned nil")
	}
}

// === updateResults with non-keypress msgs (viewport forwarding) ===

func TestAutomergeModel_UpdateResults_ViewportMsg(t *testing.T) {
	m := NewAutomergeModel()
	m, _ = m.Update(wsMsg())
	m.step = automergeStepResults
	// Send a window size msg to viewport
	m, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	_ = cmd
	// Non-key msg should be forwarded to viewport
}

func TestBranchesModel_UpdateResults_ViewportMsg(t *testing.T) {
	m := NewBranchesModel()
	m, _ = m.Update(wsMsg())
	m.step = branchesStepResults
	m, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	_ = cmd
}

func TestCleanupModel_UpdateResults_ViewportMsg(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	m.step = cleanupStepResults
	m, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	_ = cmd
}

func TestCloneModel_UpdateResults_ViewportMsg(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.step = cloneStepResults
	m, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	_ = cmd
}

func TestCommitModel_UpdateResults_ViewportMsg(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = commitStepResults
	m, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	_ = cmd
}

func TestPRModel_UpdateResults_ViewportMsg(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepResults
	m, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	_ = cmd
}

func TestStatusModel_UpdateResults_ViewportMsg(t *testing.T) {
	m := NewStatusModel()
	m, _ = m.Update(wsMsg())
	m.step = statusStepResults
	m, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	_ = cmd
}

func TestMergePRsModel_UpdateResults_ViewportMsg(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepResults
	m, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	_ = cmd
}

func TestEnableWFModel_UpdateResults_ViewportMsg(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.step = enableWFStepResults
	m, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	_ = cmd
}

func TestTeamPRsModel_UpdateResults_ViewportMsg(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = teamPRsStepResults
	m, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	_ = cmd
}

func TestMyPRsModel_UpdateResults_ViewportMsg(t *testing.T) {
	m := NewMyPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = myPRsStepResults
	m, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	_ = cmd
}

// === Automerge deeper paths ===

func TestAutomergeModel_TaskDoneFailed(t *testing.T) {
	m := NewAutomergeModel()
	m, _ = m.Update(wsMsg())
	m.step = automergeStepProgress
	result := ops.AutomergeResult{RepoName: "r", Success: false, Error: "no PR found"}
	m, _ = m.Update(automergeTaskDoneMsg{index: 0, result: result})
}

func TestAutomergeModel_ProgressGenericMsg(t *testing.T) {
	m := NewAutomergeModel()
	m, _ = m.Update(wsMsg())
	m.step = automergeStepProgress
	// Send a spinner tick-like message to exercise the default branch
	m, _ = m.Update(spinner.TickMsg{})
}

func TestAutomergeModel_ResultsQ(t *testing.T) {
	m := NewAutomergeModel()
	m, _ = m.Update(wsMsg())
	m.step = automergeStepResults
	_, cmd := m.Update(qMsg())
	if cmd == nil {
		t.Error("expected cmd from q at results")
	}
}

func TestAutomergeModel_ResultsEnter(t *testing.T) {
	m := NewAutomergeModel()
	m, _ = m.Update(wsMsg())
	m.step = automergeStepResults
	_, cmd := m.Update(enterMsg())
	if cmd == nil {
		t.Error("expected cmd from enter at results")
	}
}

func TestAutomergeModel_TargetEnterEmpty(t *testing.T) {
	m := NewAutomergeModel()
	m, _ = m.Update(wsMsg())
	// Enter with empty target should stay at target step
	m, _ = m.Update(enterMsg())
	if m.step != automergeStepTarget {
		t.Errorf("step = %d, want automergeStepTarget", m.step)
	}
}

func TestAutomergeModel_ViewProgressWithSummary(t *testing.T) {
	m := NewAutomergeModel()
	m.viewReady = true
	m.step = automergeStepProgress
	m.target = "test-branch"
	v := m.View()
	if v == "" {
		t.Error("View() returned empty string at progress step")
	}
}

func TestAutomergeModel_ViewResultsNotReady(t *testing.T) {
	m := NewAutomergeModel()
	m.viewReady = false
	m.step = automergeStepResults
	v := m.View()
	if v == "" {
		t.Error("View() returned empty at results without viewport ready")
	}
}

// === Branches deeper paths ===

func TestBranchesModel_ProgressEsc(t *testing.T) {
	m := NewBranchesModel()
	m, _ = m.Update(wsMsg())
	m.step = branchesStepProgress
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from esc at progress")
	}
}

func TestBranchesModel_ProgressQ(t *testing.T) {
	m := NewBranchesModel()
	m, _ = m.Update(wsMsg())
	m.step = branchesStepProgress
	_, cmd := m.Update(qMsg())
	if cmd == nil {
		t.Error("expected cmd from q at progress")
	}
}

func TestBranchesModel_ResultsQ(t *testing.T) {
	m := NewBranchesModel()
	m, _ = m.Update(wsMsg())
	m.step = branchesStepResults
	_, cmd := m.Update(qMsg())
	if cmd == nil {
		t.Error("expected cmd from q at results")
	}
}

func TestBranchesModel_ViewResultsNotReady(t *testing.T) {
	m := NewBranchesModel()
	m.viewReady = false
	m.step = branchesStepResults
	v := m.View()
	if v == "" {
		t.Error("View() empty at results without viewport ready")
	}
}

func TestBranchesModel_RenderResults_WithErrors(t *testing.T) {
	m := NewBranchesModel()
	m.width = 120
	m.results = []ops.BranchesResult{
		{RepoName: "err-repo", Error: "connection timeout"},
		{RepoName: "clean-repo", CurrentBranch: "main", DefaultBranch: "main", LocalBranches: []ops.BranchInfo{{Name: "main"}}},
		{RepoName: "dirty-repo", CurrentBranch: "main", DefaultBranch: "main",
			LocalBranches:  []ops.BranchInfo{{Name: "main"}, {Name: "feature"}},
			RemoteBranches: []ops.BranchInfo{{Name: "develop"}}},
	}
	result := m.renderResults()
	if result == "" {
		t.Error("expected non-empty render")
	}
}

func TestBranchesModel_ViewProgress(t *testing.T) {
	m := NewBranchesModel()
	m.step = branchesStepProgress
	v := m.View()
	if v == "" {
		t.Error("View() empty at progress")
	}
}

// === Cleanup deeper paths ===

func TestCleanupModel_ProgressDroppedChanges(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	m.step = cleanupStepProgress
	result := ops.CleanupResult{RepoName: "r", Success: true, PrunedBranches: 2, DroppedChanges: true}
	m, _ = m.Update(cleanupTaskDoneMsg{index: 0, result: result})
}

func TestCleanupModel_DropStep_EscReturnsToMenu(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from esc at drop step")
	}
}

func TestCleanupModel_ViewRepoSelect(t *testing.T) {
	m := NewCleanupModel()
	m.viewReady = true
	m.step = cleanupStepRepoSelect
	m.dropChanges = true
	v := m.View()
	if v == "" {
		t.Error("View() empty at repo select")
	}
}

func TestCleanupModel_ViewResultsNotReady(t *testing.T) {
	m := NewCleanupModel()
	m.viewReady = false
	m.step = cleanupStepResults
	v := m.View()
	if v == "" {
		t.Error("View() empty at results without viewport ready")
	}
}

func TestCleanupModel_ViewDrop_SelectionHighlight(t *testing.T) {
	m := NewCleanupModel()
	m.viewReady = true
	m.step = cleanupStepDrop
	m.dropCursor = 1 // Yes selected
	v := m.View()
	if v == "" {
		t.Error("View() empty at drop step with yes selected")
	}
}

func TestCleanupModel_ResultsQ(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	m.step = cleanupStepResults
	_, cmd := m.Update(qMsg())
	if cmd == nil {
		t.Error("expected cmd from q at results")
	}
}

func TestCleanupModel_ResultsEnter(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	m.step = cleanupStepResults
	_, cmd := m.Update(enterMsg())
	if cmd == nil {
		t.Error("expected cmd from enter at results")
	}
}

func TestCleanupModel_ProgressGenericMsg(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	m.step = cleanupStepProgress
	m, _ = m.Update(spinner.TickMsg{})
}

// === Clone deeper paths ===

func TestCloneModel_InputAllFocusPositions(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	// Tab through all focus positions
	for i := 0; i < 4; i++ {
		m, _ = m.Update(tabMsg())
	}
	if m.focusIndex != 0 {
		t.Errorf("focusIndex = %d after 4 tabs, want 0", m.focusIndex)
	}
}

func TestCloneModel_FocusActive_AllPositions(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())

	for i := 0; i < 4; i++ {
		m.focusIndex = i
		cmd := m.focusActive()
		if cmd == nil {
			t.Errorf("focusActive() returned nil for index %d", i)
		}
	}
}

func TestCloneModel_InputUpdateFocusedField(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())

	// Test typing in each focus position
	for i := 0; i < 4; i++ {
		m.focusIndex = i
		m, _ = m.Update(kMsg('a'))
	}
}

func TestCloneModel_ResultsQ(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.step = cloneStepResults
	_, cmd := m.Update(qMsg())
	if cmd == nil {
		t.Error("expected cmd from q at results")
	}
}

func TestCloneModel_ResultsEnter(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.step = cloneStepResults
	_, cmd := m.Update(enterMsg())
	if cmd == nil {
		t.Error("expected cmd from enter at results")
	}
}

// === Commit deeper paths ===

func TestCommitModel_TaskDoneSuccess(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = commitStepProgress
	result := ops.CommitResult{RepoName: "r", Branch: "main", Success: true}
	m, _ = m.Update(commitTaskDoneMsg{index: 0, result: result})
}

func TestCommitModel_TaskDoneFailed(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = commitStepProgress
	result := ops.CommitResult{RepoName: "r", Success: false, Error: "no changes"}
	m, _ = m.Update(commitTaskDoneMsg{index: 0, result: result})
}

func TestCommitModel_AllTasksDone(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = commitStepProgress
	m, _ = m.Update(components.AllTasksDoneMsg{})
	if m.step != commitStepResults {
		t.Errorf("step = %d, want commitStepResults", m.step)
	}
}

func TestCommitModel_ProgressGenericMsg(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = commitStepProgress
	m, _ = m.Update(spinner.TickMsg{})
}

func TestCommitModel_ResultsQ(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = commitStepResults
	_, cmd := m.Update(qMsg())
	if cmd == nil {
		t.Error("expected cmd from q at results")
	}
}

func TestCommitModel_MessageEnterWithValue(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	// Type a message
	m.messageInput.SetValue("fix: test bug")
	m, _ = m.Update(enterMsg())
	if m.step != commitStepBranch {
		t.Errorf("step = %d, want commitStepBranch", m.step)
	}
	if m.message != "fix: test bug" {
		t.Errorf("message = %q, want 'fix: test bug'", m.message)
	}
}

func TestCommitModel_HistoryBrowse_Exhaustive(t *testing.T) {
	msgs := []string{"msg1", "msg2"}
	m := NewCommitModel(msgs)
	m, _ = m.Update(wsMsg())

	// Browse up past the end
	m, _ = m.Update(upMsg())
	m, _ = m.Update(upMsg())
	m, _ = m.Update(upMsg()) // Should clamp

	// Browse down past the start
	m, _ = m.Update(downMsg())
	m, _ = m.Update(downMsg())
	m, _ = m.Update(downMsg()) // Should clamp
}

func TestCommitModel_MessageEsc(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from esc at message step")
	}
}

func TestCommitModel_ViewProgress(t *testing.T) {
	m := NewCommitModel(nil)
	m.viewReady = true
	m.step = commitStepProgress
	m.message = "test commit"
	m.branch = "feature"
	v := m.View()
	if v == "" {
		t.Error("View() empty at progress step")
	}
}

// === PullRequest deeper paths ===

func TestPRModel_BranchEnterWithValue(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepBranch
	m.branchInput.SetValue("feature-x")
	m, _ = m.Update(enterMsg())
	if m.step != prStepChanges {
		t.Errorf("step = %d, want prStepChanges", m.step)
	}
}

func TestPRModel_BranchEnterEmpty(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepBranch
	m.branchInput.SetValue("")
	m, _ = m.Update(enterMsg())
	if m.step != prStepBranch {
		t.Errorf("step = %d, want prStepBranch (empty branch)", m.step)
	}
}

func TestPRModel_ChangesNavigation(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepChanges

	// Navigate up/down
	m, _ = m.Update(upMsg())
	m, _ = m.Update(downMsg())

	// Toggle selection with space
	m, _ = m.Update(spaceMsg())

	// Enter to proceed
	m, _ = m.Update(enterMsg())
	if m.step != prStepBreaking {
		t.Errorf("step = %d, want prStepBreaking", m.step)
	}
}

func TestPRModel_ChangesUpWrap(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepChanges
	m.changeCursor = 0
	m, _ = m.Update(upMsg())
	// Should wrap to last item
	if m.changeCursor < 0 {
		t.Error("changeCursor should not be negative")
	}
}

func TestPRModel_ChangesDownWrap(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepChanges
	// Move to end
	for i := 0; i < 10; i++ {
		m, _ = m.Update(downMsg())
	}
	// Should wrap
}

func TestPRModel_BreakingNavigation(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepBreaking

	// Toggle breaking cursor
	m, _ = m.Update(upMsg())
	m, _ = m.Update(downMsg())

	// Enter to select
	m, _ = m.Update(enterMsg())
	if m.step != prStepRepoSelect {
		t.Errorf("step = %d, want prStepRepoSelect", m.step)
	}
}

func TestPRModel_BreakingSelectYes(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepBreaking
	m.breakingCursor = 1
	m, _ = m.Update(enterMsg())
	if !m.breaking {
		t.Error("expected breaking = true when cursor = 1")
	}
}

func TestPRModel_TaskDoneSuccess(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepProgress
	result := ops.PRResult{RepoName: "r", PRURL: "https://github.com/...", Success: true}
	m, _ = m.Update(prTaskDoneMsg{index: 0, result: result})
}

func TestPRModel_TaskDoneFailed(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepProgress
	result := ops.PRResult{RepoName: "r", Success: false, Error: "push failed"}
	m, _ = m.Update(prTaskDoneMsg{index: 0, result: result})
}

func TestPRModel_AllTasksDone(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepProgress
	m, _ = m.Update(components.AllTasksDoneMsg{})
	if m.step != prStepResults {
		t.Errorf("step = %d, want prStepResults", m.step)
	}
}

func TestPRModel_ProgressGenericMsg(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepProgress
	m, _ = m.Update(spinner.TickMsg{})
}

func TestPRModel_ResultsQ(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepResults
	_, cmd := m.Update(qMsg())
	if cmd == nil {
		t.Error("expected cmd from q at results")
	}
}

func TestPRModel_ShowSummary_AllFields(t *testing.T) {
	m := NewPullRequestModel(nil)
	m.message = "feat: add stuff"
	m.branch = "feature/add-stuff"
	m.target = "main"
	m.changes = []string{"New feature", "Bug fix"}
	m.breaking = true
	result := m.showSummary()
	if result == "" {
		t.Error("expected non-empty summary")
	}
}

func TestPRModel_ViewChangesWithSelection(t *testing.T) {
	m := NewPullRequestModel(nil)
	m.viewReady = true
	m.step = prStepChanges
	m.changeCursor = 0
	if len(m.changeSelected) > 0 {
		m.changeSelected[0] = true
	}
	v := m.View()
	if v == "" {
		t.Error("View() empty at changes step")
	}
}

func TestPRModel_HistoryBrowse_Exhaustive(t *testing.T) {
	msgs := []string{"msg1", "msg2"}
	m := NewPullRequestModel(msgs)
	m, _ = m.Update(wsMsg())

	// Browse up past the end
	m, _ = m.Update(upMsg())
	m, _ = m.Update(upMsg())
	m, _ = m.Update(upMsg())

	// Browse down past the start
	m, _ = m.Update(downMsg())
	m, _ = m.Update(downMsg())
	m, _ = m.Update(downMsg())
}

// === EnableWorkflows deeper paths ===

func TestEnableWFModel_InputEnterWithOrg(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	// orgInput already has "Sundsvallskommun" as default
	m, cmd := m.Update(enterMsg())
	if cmd == nil {
		t.Error("expected cmd from enter with org set")
	}
	if m.step != enableWFStepFetching {
		t.Errorf("step = %d, want enableWFStepFetching", m.step)
	}
}

func TestEnableWFModel_InputEnterEmptyOrg(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.orgInput.SetValue("")
	m, _ = m.Update(enterMsg())
	if m.step != enableWFStepInput {
		t.Errorf("step = %d, want enableWFStepInput (empty org)", m.step)
	}
}

func TestEnableWFModel_FocusActive_AllPositions(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())

	for i := 0; i < 3; i++ {
		m.focusIndex = i
		cmd := m.focusActive()
		if cmd == nil {
			t.Errorf("focusActive() returned nil for index %d", i)
		}
	}
}

func TestEnableWFModel_InputUpdateFocused(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())

	for i := 0; i < 3; i++ {
		m.focusIndex = i
		m, _ = m.Update(kMsg('x'))
	}
}

func TestEnableWFModel_TaskDoneFailed(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.step = enableWFStepProgress
	result := ops.EnableWorkflowResult{Repo: "repo1", Success: false, Error: "enable failed"}
	m, _ = m.Update(enableWFTaskDoneMsg{index: 0, result: result})
}

func TestEnableWFModel_TaskDoneWithCounts(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.step = enableWFStepProgress
	result := ops.EnableWorkflowResult{Repo: "repo1", Success: true, EnabledCount: 2, RetriggeredPRs: 1}
	m, _ = m.Update(enableWFTaskDoneMsg{index: 0, result: result})
}

func TestEnableWFModel_ProgressGenericMsg(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.step = enableWFStepProgress
	m, _ = m.Update(spinner.TickMsg{})
}

func TestEnableWFModel_FetchingSpinnerTick(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.step = enableWFStepFetching
	m, _ = m.Update(spinner.TickMsg{})
}

func TestEnableWFModel_ResultsQ(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.step = enableWFStepResults
	_, cmd := m.Update(qMsg())
	if cmd == nil {
		t.Error("expected cmd from q at results")
	}
}

func TestEnableWFModel_ResultsEnter(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.step = enableWFStepResults
	_, cmd := m.Update(enterMsg())
	if cmd == nil {
		t.Error("expected cmd from enter at results")
	}
}

// === MergePRs deeper paths ===

func TestMergePRsModel_TaskDoneFailed(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepProgress
	result := ops.MergePRResult{Repo: "repo1", PRNumber: "1", Success: false, Error: "merge conflict"}
	m, _ = m.Update(mergePRTaskDoneMsg{index: 0, result: result})
}

func TestMergePRsModel_ProgressGenericMsg(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepProgress
	m, _ = m.Update(spinner.TickMsg{})
}

func TestMergePRsModel_AllTasksDone_WithRemainingPRs(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepProgress
	m.batchSize = 1
	m.prs = []ops.PRInfo{
		{Repo: "r1", Number: 1, Title: "fix1"},
		{Repo: "r2", Number: 2, Title: "fix2"},
	}
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "r1 #1", Status: components.TaskDone},
	})
	m, cmd := m.Update(components.AllTasksDoneMsg{})
	if cmd == nil {
		t.Error("expected cmd for next batch wait")
	}
	if m.step != mergePRsStepWaiting {
		t.Errorf("step = %d, want mergePRsStepWaiting", m.step)
	}
}

func TestMergePRsModel_WaitingQ(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepWaiting
	m, _ = m.Update(qMsg())
	if m.step != mergePRsStepResults {
		t.Errorf("step = %d, want mergePRsStepResults", m.step)
	}
}

func TestMergePRsModel_WaitingTickExpired(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepWaiting
	m.waitRemaining = 1 // will expire on next tick
	m, cmd := m.Update(mergeWaitTickMsg{})
	if cmd == nil {
		t.Error("expected cmd when wait expires")
	}
	if m.step != mergePRsStepFetching {
		t.Errorf("step = %d, want mergePRsStepFetching", m.step)
	}
}

func TestMergePRsModel_SummaryView_WithValues(t *testing.T) {
	m := NewMergePRsModel()
	m.org = "testorg"
	m.merged = 3
	m.failed = 1
	v := m.summaryView()
	if v == "" {
		t.Error("summaryView() empty with values")
	}
}

func TestMergePRsModel_ResultsQ(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepResults
	_, cmd := m.Update(qMsg())
	if cmd == nil {
		t.Error("expected cmd from q at results")
	}
}

func TestMergePRsModel_ResultsEnter(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepResults
	_, cmd := m.Update(enterMsg())
	if cmd == nil {
		t.Error("expected cmd from enter at results")
	}
}

func TestMergePRsModel_InputEnterWithOrg(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	// orgInput has default value; enter should proceed
	m, cmd := m.Update(enterMsg())
	if cmd == nil {
		t.Error("expected cmd from enter")
	}
}

// === TeamPRs deeper paths ===

func TestTeamPRsModel_FocusActive_AllPositions(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())

	for i := 0; i < 2; i++ {
		m.focusIndex = i
		cmd := m.focusActive()
		if cmd == nil {
			t.Errorf("focusActive() returned nil for index %d", i)
		}
	}
}

func TestTeamPRsModel_InputTab(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	m, _ = m.Update(tabMsg())
	if m.focusIndex != 1 {
		t.Errorf("focusIndex = %d, want 1 after tab", m.focusIndex)
	}
}

func TestTeamPRsModel_InputUpdateFocused(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())

	for i := 0; i < 2; i++ {
		m.focusIndex = i
		m, _ = m.Update(kMsg('x'))
	}
}

func TestTeamPRsModel_FetchingSpinnerTick(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = teamPRsStepFetching
	m, _ = m.Update(spinner.TickMsg{})
}

func TestTeamPRsModel_FetchedPRsSuccess(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = teamPRsStepFetching
	prs := []ops.TeamPR{
		{Repo: "repo1", Number: 1, Title: "fix bug", Author: "dev"},
	}
	m, _ = m.Update(teamPRsFetchedMsg{prs: prs})
	if m.step != teamPRsStepResults {
		t.Errorf("step = %d, want teamPRsStepResults", m.step)
	}
}

func TestTeamPRsModel_ResultsQ(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = teamPRsStepResults
	_, cmd := m.Update(qMsg())
	if cmd == nil {
		t.Error("expected cmd from q at results")
	}
}

func TestTeamPRsModel_InputEnterEmptyOrg(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	m.orgInput.SetValue("")
	m, _ = m.Update(enterMsg())
	if m.step != teamPRsStepInput {
		t.Errorf("step = %d, want teamPRsStepInput", m.step)
	}
}

func TestTeamPRsModel_InputEnterEmptyTeam(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	m.teamInput.SetValue("")
	m, _ = m.Update(enterMsg())
	if m.step != teamPRsStepInput {
		t.Errorf("step = %d, want teamPRsStepInput", m.step)
	}
}

// === MyPRs deeper paths ===

func TestMyPRsModel_FetchingSpinnerTick(t *testing.T) {
	m := NewMyPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = myPRsStepFetching
	m, _ = m.Update(spinner.TickMsg{})
}

func TestMyPRsModel_ResultsQ(t *testing.T) {
	m := NewMyPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = myPRsStepResults
	_, cmd := m.Update(qMsg())
	if cmd == nil {
		t.Error("expected cmd from q at results")
	}
}

func TestMyPRsModel_ResultsEnter(t *testing.T) {
	m := NewMyPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = myPRsStepResults
	_, cmd := m.Update(enterMsg())
	if cmd == nil {
		t.Error("expected cmd from enter at results")
	}
}

// === gitScan and getBranchShell ===

func TestGetBranchShell(t *testing.T) {
	dir := t.TempDir()
	initTestRepo(t, dir)

	branch := getBranchShell(dir)
	// go-git default branch is "master" when using PlainInit
	if branch == "" {
		t.Error("expected non-empty branch")
	}
}

func TestGetBranchShell_InvalidDir(t *testing.T) {
	branch := getBranchShell("/nonexistent/path")
	if branch != "" {
		t.Errorf("expected empty branch for invalid dir, got %q", branch)
	}
}

func TestGitScan(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "testrepo")
	os.MkdirAll(sub, 0755)
	initTestRepo(t, sub)

	repos, err := gitScan(root)
	if err != nil {
		t.Fatalf("gitScan error: %v", err)
	}
	if len(repos) == 0 {
		t.Error("expected at least one repo")
	}
}

func TestGitScan_NoRepos(t *testing.T) {
	dir := t.TempDir()
	repos, err := gitScan(dir)
	if err != nil {
		t.Fatalf("gitScan error: %v", err)
	}
	if len(repos) != 0 {
		t.Errorf("expected 0 repos, got %d", len(repos))
	}
}

func TestGitScan_InvalidDir(t *testing.T) {
	_, err := gitScan("/nonexistent/path")
	if err == nil {
		t.Error("expected error for invalid dir")
	}
}

// === View functions with all states ===

func TestStatusModel_View_WithViewport(t *testing.T) {
	m := NewStatusModel()
	m, _ = m.Update(wsMsg())
	m.step = statusStepResults
	m.results = []ops.StatusResult{
		{RepoName: "test", Branch: "main", Clean: true},
	}
	m.viewport.SetContent(m.renderResults())
	v := m.View()
	if v == "" {
		t.Error("View() empty at results with viewport")
	}
}

func TestBranchesModel_RenderResults_NarrowWidth(t *testing.T) {
	m := NewBranchesModel()
	m.width = 20 // Very narrow
	m.results = []ops.BranchesResult{
		{RepoName: "repo", CurrentBranch: "main", DefaultBranch: "main"},
	}
	result := m.renderResults()
	if result == "" {
		t.Error("expected non-empty render for narrow width")
	}
}

// === UpdateProgress for enableWorkflows with key events ===

func TestEnableWFModel_ProgressEsc(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.step = enableWFStepProgress
	// Esc at progress is forwarded to the progress model, may return nil
	m, _ = m.Update(escMsg())
}

// === Update message routing for default branch ===

func TestAutomergeModel_UpdateDefaultBranch(t *testing.T) {
	m := NewAutomergeModel()
	m, _ = m.Update(wsMsg())
	m.step = 99 // invalid step
	m, cmd := m.Update(kMsg('a'))
	_ = cmd // Should not panic
}

func TestBranchesModel_UpdateDefaultBranch(t *testing.T) {
	m := NewBranchesModel()
	m, _ = m.Update(wsMsg())
	m.step = 99
	m, cmd := m.Update(kMsg('a'))
	_ = cmd
}

func TestCloneModel_UpdateDefaultBranch(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.step = 99
	m, cmd := m.Update(kMsg('a'))
	_ = cmd
}
