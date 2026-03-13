package screens

import (
	"testing"

	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/CheeziCrew/swissgit/ops"
	"github.com/CheeziCrew/swissgit/tui/components"
)

// Key helpers are defined in screens_update_test.go: escMsg, enterMsg, upMsg, downMsg, spaceMsg

// === PR updateMessage - cover all branches (73.9% -> higher) ===

func TestPRModel_UpdateMessage_EmptyEnter(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepMessage
	// Empty message, enter should be no-op
	m, cmd := m.Update(enterMsg())
	if cmd != nil {
		t.Error("expected nil cmd for empty message enter")
	}
	if m.step != prStepMessage {
		t.Errorf("step should stay at prStepMessage, got %d", m.step)
	}
}

func TestPRModel_UpdateMessage_WithValue(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepMessage
	m.messageInput.SetValue("fix: bug")
	m, cmd := m.Update(enterMsg())
	if cmd == nil {
		t.Error("expected non-nil cmd for valid message enter")
	}
	if m.step != prStepBranch {
		t.Errorf("step should move to prStepBranch, got %d", m.step)
	}
	if m.message != "fix: bug" {
		t.Errorf("message = %q, want %q", m.message, "fix: bug")
	}
}

func TestPRModel_UpdateMessage_Esc(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepMessage
	m, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected BackToMenuMsg cmd for esc")
	}
}

func TestPRModel_UpdateMessage_HistoryUpDown(t *testing.T) {
	recent := []string{"fix: first", "feat: second", "chore: third"}
	m := NewPullRequestModel(recent)
	m, _ = m.Update(wsMsg())
	m.step = prStepMessage
	m.messageInput.SetValue("current typing")

	// Press up to browse history
	m, _ = m.Update(upMsg())
	if m.history.cursor != 0 {
		t.Errorf("history.cursor after up = %d, want 0", m.history.cursor)
	}
	if m.messageInput.Value() != "fix: first" {
		t.Errorf("value = %q, want %q", m.messageInput.Value(), "fix: first")
	}

	// Press up again
	m, _ = m.Update(upMsg())
	if m.history.cursor != 1 {
		t.Errorf("history.cursor after 2nd up = %d, want 1", m.history.cursor)
	}

	// Press down to go back
	m, _ = m.Update(downMsg())
	if m.history.cursor != 0 {
		t.Errorf("history.cursor after down = %d, want 0", m.history.cursor)
	}

	// Press down past 0 to restore typed value
	m, _ = m.Update(downMsg())
	if m.history.cursor != -1 {
		t.Errorf("history.cursor after 2nd down = %d, want -1", m.history.cursor)
	}
	if m.messageInput.Value() != "current typing" {
		t.Errorf("value = %q, want %q", m.messageInput.Value(), "current typing")
	}
}

func TestPRModel_UpdateMessage_UpNoHistory(t *testing.T) {
	m := NewPullRequestModel(nil) // no history
	m, _ = m.Update(wsMsg())
	m.step = prStepMessage
	// Up with empty history should be no-op (fallthrough to text input update)
	m, _ = m.Update(upMsg())
	if m.history.IsActive() {
		t.Errorf("history should not be active, cursor = %d", m.history.cursor)
	}
}

func TestPRModel_UpdateMessage_DownNoHistory(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepMessage
	// Down with no cursor should be no-op
	m, _ = m.Update(downMsg())
}

func TestPRModel_UpdateMessage_RegularKey(t *testing.T) {
	m := NewPullRequestModel([]string{"old"})
	m, _ = m.Update(wsMsg())
	m.step = prStepMessage
	// First browse up
	m, _ = m.Update(upMsg())
	if !m.history.IsActive() {
		t.Error("expected history active after up")
	}
	// Now type a regular key - should reset cursor
	m, _ = m.Update(tea.KeyPressMsg{Code: 'a'})
	if m.history.IsActive() {
		t.Error("expected history inactive after regular key")
	}
}

// === PR updateBranch - cover empty enter and esc ===

func TestPRModel_UpdateBranch_EmptyEnter(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepBranch
	m, cmd := m.Update(enterMsg())
	if cmd != nil {
		t.Error("expected nil cmd for empty branch enter")
	}
}

func TestPRModel_UpdateBranch_WithValue(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepBranch
	m.branchInput.SetValue("feature-x")
	m, _ = m.Update(enterMsg())
	if m.step != prStepChanges {
		t.Errorf("step = %d, want prStepChanges", m.step)
	}
	if m.branch != "feature-x" {
		t.Errorf("branch = %q, want %q", m.branch, "feature-x")
	}
}

func TestPRModel_UpdateBranch_Esc(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepBranch
	m, _ = m.Update(escMsg())
	if m.step != prStepMessage {
		t.Errorf("step = %d, want prStepMessage", m.step)
	}
}

// === PR updateChanges - cover cursor navigation, space, enter, esc ===

func TestPRModel_UpdateChanges_Navigation(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepChanges
	m.changeCursor = 0

	// Down
	m, _ = m.Update(downMsg())
	if m.changeCursor != 1 {
		t.Errorf("changeCursor = %d, want 1", m.changeCursor)
	}

	// Up wraps around
	m.changeCursor = 0
	m, _ = m.Update(upMsg())
	if m.changeCursor != len(m.changeTypes)-1 {
		t.Errorf("changeCursor after up wrap = %d, want %d", m.changeCursor, len(m.changeTypes)-1)
	}

	// Down wraps around
	m.changeCursor = len(m.changeTypes) - 1
	m, _ = m.Update(downMsg())
	if m.changeCursor != 0 {
		t.Errorf("changeCursor after down wrap = %d, want 0", m.changeCursor)
	}

	// Space toggles
	m.changeCursor = 0
	m, _ = m.Update(spaceMsg())
	if !m.changeSelected[0] {
		t.Error("expected changeSelected[0] = true after space")
	}
	m, _ = m.Update(spaceMsg())
	if m.changeSelected[0] {
		t.Error("expected changeSelected[0] = false after second space")
	}
}

func TestPRModel_UpdateChanges_Enter(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepChanges
	m.changeSelected[0] = true
	m.changeSelected[2] = true
	m, _ = m.Update(enterMsg())
	if m.step != prStepBreaking {
		t.Errorf("step = %d, want prStepBreaking", m.step)
	}
	if len(m.changes) == 0 {
		t.Error("expected changes to be collected")
	}
}

func TestPRModel_UpdateChanges_Esc(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepChanges
	m, _ = m.Update(escMsg())
	if m.step != prStepBranch {
		t.Errorf("step = %d, want prStepBranch", m.step)
	}
}

// === PR updateBreaking ===

func TestPRModel_UpdateBreaking_Toggle(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepBreaking
	m.breakingCursor = 0

	m, _ = m.Update(downMsg())
	if m.breakingCursor != 1 {
		t.Errorf("breakingCursor = %d, want 1", m.breakingCursor)
	}

	m, _ = m.Update(upMsg())
	if m.breakingCursor != 0 {
		t.Errorf("breakingCursor = %d, want 0", m.breakingCursor)
	}
}

func TestPRModel_UpdateBreaking_Enter(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepBreaking
	m.breakingCursor = 1
	m.message = "fix: test"
	m.branch = "feature"
	m, cmd := m.Update(enterMsg())
	if m.step != prStepRepoSelect {
		t.Errorf("step = %d, want prStepRepoSelect", m.step)
	}
	if !m.breaking {
		t.Error("expected breaking = true")
	}
	if cmd == nil {
		t.Error("expected repoSelect Init cmd")
	}
}

func TestPRModel_UpdateBreaking_Esc(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepBreaking
	m, _ = m.Update(escMsg())
	if m.step != prStepChanges {
		t.Errorf("step = %d, want prStepChanges", m.step)
	}
}

// === PR updateRepoSelect ===

func TestPRModel_UpdateRepoSelect_Done(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepRepoSelect
	m.message = "fix: test"
	m.branch = "feature"
	m, cmd := m.Update(RepoSelectDoneMsg{Paths: []string{"/tmp/repo"}})
	if m.step != prStepProgress {
		t.Errorf("step = %d, want prStepProgress", m.step)
	}
	if cmd == nil {
		t.Error("expected batch cmd")
	}
}

func TestPRModel_UpdateRepoSelect_Back(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepRepoSelect
	m, _ = m.Update(BackToMenuMsg{})
	if m.step != prStepBreaking {
		t.Errorf("step = %d, want prStepBreaking", m.step)
	}
}

// === PR updateProgress ===

func TestPRModel_UpdateProgress_TaskDone_Success(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepProgress
	m.repos = []string{"/tmp/repo"}
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "repo", Path: "/tmp/repo", Status: components.TaskRunning},
	})
	m, _ = m.Update(prTaskDoneMsg{index: 0, result: ops.PRResult{Success: true, PRURL: "https://github.com/pr/1"}})
}

func TestPRModel_UpdateProgress_TaskDone_Fail(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepProgress
	m.repos = []string{"/tmp/repo"}
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "repo", Path: "/tmp/repo", Status: components.TaskRunning},
	})
	m, _ = m.Update(prTaskDoneMsg{index: 0, result: ops.PRResult{Success: false, Error: "fail"}})
}

func TestPRModel_UpdateProgress_AllDone(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepProgress
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "repo", Path: "/tmp/repo", Status: components.TaskDone},
	})
	m, _ = m.Update(components.AllTasksDoneMsg{})
	if m.step != prStepResults {
		t.Errorf("step = %d, want prStepResults", m.step)
	}
}

// === PR View ===

func TestPRModel_View_AllSteps(t *testing.T) {
	for _, step := range []prStep{prStepMessage, prStepBranch, prStepChanges, prStepBreaking, prStepRepoSelect, prStepProgress, prStepResults} {
		m := NewPullRequestModel([]string{"recent msg"})
		m, _ = m.Update(wsMsg())
		m.step = step
		m.message = "fix: test"
		m.branch = "feature"
		if step == prStepProgress {
			m.progress = components.NewProgressModel(nil)
		}
		if step == prStepResults {
			m.results = components.NewResultModel("PR", nil)
			m.viewport.SetContent(m.results.View())
		}
		if step == prStepRepoSelect {
			m.repoSelect = NewRepoSelectModel("pr", ".", 5, 30)
		}
		v := m.View()
		if v == "" {
			t.Errorf("View() empty for step %d", step)
		}
	}
}

// === Automerge updateTarget (75%) ===

func TestAutomergeModel_UpdateTarget_EmptyEnter(t *testing.T) {
	m := NewAutomergeModel()
	m, _ = m.Update(wsMsg())
	m.step = automergeStepTarget
	m, cmd := m.Update(enterMsg())
	if cmd != nil {
		t.Error("expected nil cmd for empty target enter")
	}
}

func TestAutomergeModel_UpdateTarget_Esc(t *testing.T) {
	m := NewAutomergeModel()
	m, _ = m.Update(wsMsg())
	m.step = automergeStepTarget
	m, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected BackToMenuMsg cmd for esc")
	}
}

func TestAutomergeModel_UpdateTarget_RegularKey(t *testing.T) {
	m := NewAutomergeModel()
	m, _ = m.Update(wsMsg())
	m.step = automergeStepTarget
	// Regular key should update textinput
	m, _ = m.Update(tea.KeyPressMsg{Code: 'a'})
}

// === MergePRs startWait (75%) ===

func TestMergePRsModel_StartWait_Fields(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.waitMin = 1
	cmd := m.startWait()
	if cmd == nil {
		t.Error("expected tick cmd from startWait")
	}
	if m.step != mergePRsStepWaiting {
		t.Errorf("step = %d, want mergePRsStepWaiting", m.step)
	}
	if m.waitRemaining != 60 {
		t.Errorf("waitRemaining = %d, want 60", m.waitRemaining)
	}
}

func TestMergePRsModel_UpdateWaiting_TickDecrement(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepWaiting
	m.waitRemaining = 5
	m, cmd := m.Update(mergeWaitTickMsg{})
	if m.waitRemaining != 4 {
		t.Errorf("waitRemaining = %d, want 4", m.waitRemaining)
	}
	if cmd == nil {
		t.Error("expected tick cmd")
	}
}

func TestMergePRsModel_UpdateWaiting_TickExpires(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepWaiting
	m.waitRemaining = 1
	m, cmd := m.Update(mergeWaitTickMsg{})
	if m.step != mergePRsStepFetching {
		t.Errorf("step = %d, want mergePRsStepFetching", m.step)
	}
	if cmd == nil {
		t.Error("expected batch cmd")
	}
}

func TestMergePRsModel_UpdateWaiting_EnterSkips(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepWaiting
	m.waitRemaining = 30
	m, cmd := m.Update(enterMsg())
	if m.step != mergePRsStepFetching {
		t.Errorf("step = %d, want mergePRsStepFetching", m.step)
	}
	if cmd == nil {
		t.Error("expected batch cmd")
	}
}

func TestMergePRsModel_UpdateWaiting_Esc(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepWaiting
	m, _ = m.Update(escMsg())
	if m.step != mergePRsStepResults {
		t.Errorf("step = %d, want mergePRsStepResults", m.step)
	}
}

// === MergePRs summaryView (83.3%) ===

func TestMergePRsModel_SummaryView_WithBatch(t *testing.T) {
	m := NewMergePRsModel()
	m.org = "test-org"
	m.batchIndex = 2
	m.batchSize = 5
	m.waitMin = 10
	m.merged = 3
	m.failed = 1
	v := m.summaryView()
	if v == "" {
		t.Error("summaryView returned empty")
	}
}

func TestMergePRsModel_SummaryView_NoBatch(t *testing.T) {
	m := NewMergePRsModel()
	m.org = "test-org"
	m.batchIndex = 0
	v := m.summaryView()
	if v == "" {
		t.Error("summaryView returned empty")
	}
}

// === startTasks with repos that have GetRepoNameForPath error ===

func TestStartCleanupTasks_WithRepos(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	m.repos = []string{"/nonexistent/repo1", "/nonexistent/repo2"}
	m.dropChanges = true
	cmd := m.startCleanupTasks()
	if cmd == nil {
		t.Error("expected batch cmd")
	}
	if m.step != cleanupStepProgress {
		t.Errorf("step = %d, want cleanupStepProgress", m.step)
	}
}

func TestStartCommitTasks_WithRepos(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	m.repos = []string{"/nonexistent/repo"}
	m.message = "fix: test"
	m.branch = "feature"
	cmd := m.startCommitTasks()
	if cmd == nil {
		t.Error("expected batch cmd")
	}
	if m.step != commitStepProgress {
		t.Errorf("step = %d, want commitStepProgress", m.step)
	}
}

func TestStartPRTasks_WithRepos(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.repos = []string{"/nonexistent/repo1", "/nonexistent/repo2"}
	m.message = "fix: test"
	m.branch = "feature"
	m.target = "main"
	cmd := m.startPRTasks()
	if cmd == nil {
		t.Error("expected batch cmd")
	}
	if m.step != prStepProgress {
		t.Errorf("step = %d, want prStepProgress", m.step)
	}
}

func TestStartStatusTasks_WithRepos(t *testing.T) {
	m := NewStatusModel()
	m, _ = m.Update(wsMsg())
	cmd := m.startStatusTasks([]string{"/nonexistent/repo"})
	if cmd == nil {
		t.Error("expected batch cmd")
	}
}

func TestStartBranchesTasks_WithRepos(t *testing.T) {
	m := NewBranchesModel()
	m, _ = m.Update(wsMsg())
	cmd := m.startBranchesTasks([]string{"/nonexistent/repo"})
	if cmd == nil {
		t.Error("expected batch cmd")
	}
}

func TestStartEnableWFTasks_WithRepos(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.repos = []string{"repo1", "repo2"}
	m.org = "test-org"
	cmd := m.startTasks()
	if cmd == nil {
		t.Error("expected batch cmd")
	}
	if m.step != enableWFStepProgress {
		t.Errorf("step = %d, want enableWFStepProgress", m.step)
	}
}

// === MergePRs startBatch (85.7%) ===

func TestMergePRsModel_StartBatch_WithPRs(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.prs = []ops.PRInfo{
		{Repo: "repo-a", Number: 1, Title: "Fix"},
		{Repo: "repo-b", Number: 2, Title: "Add"},
	}
	m.batchSize = 5
	cmd := m.startBatch()
	if cmd == nil {
		t.Error("expected batch cmd from startBatch")
	}
	if m.step != mergePRsStepProgress {
		t.Errorf("step = %d, want mergePRsStepProgress", m.step)
	}
}

// === Cleanup updateDrop ===

func TestCleanupModel_UpdateDrop_Toggle(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	m.step = cleanupStepDrop
	m.dropCursor = 0

	m, _ = m.Update(upMsg())
	if m.dropCursor != 1 {
		t.Errorf("dropCursor = %d, want 1", m.dropCursor)
	}

	m, _ = m.Update(downMsg())
	if m.dropCursor != 0 {
		t.Errorf("dropCursor = %d, want 0", m.dropCursor)
	}
}

func TestCleanupModel_UpdateDrop_Enter(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	m.step = cleanupStepDrop
	m.dropCursor = 1
	m, cmd := m.Update(enterMsg())
	if m.step != cleanupStepRepoSelect {
		t.Errorf("step = %d, want cleanupStepRepoSelect", m.step)
	}
	if !m.dropChanges {
		t.Error("expected dropChanges = true")
	}
	if cmd == nil {
		t.Error("expected repoSelect init cmd")
	}
}

func TestCleanupModel_UpdateDrop_Esc(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	m.step = cleanupStepDrop
	m, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected BackToMenuMsg cmd for esc")
	}
}

// === Cleanup updateRepoSelect ===

func TestCleanupModel_UpdateRepoSelect_Done(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	m.step = cleanupStepRepoSelect
	m, cmd := m.Update(RepoSelectDoneMsg{Paths: []string{"/tmp/repo"}})
	if m.step != cleanupStepProgress {
		t.Errorf("step = %d, want cleanupStepProgress", m.step)
	}
	if cmd == nil {
		t.Error("expected batch cmd")
	}
}

func TestCleanupModel_UpdateRepoSelect_Back(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	m.step = cleanupStepRepoSelect
	m, _ = m.Update(BackToMenuMsg{})
	if m.step != cleanupStepDrop {
		t.Errorf("step = %d, want cleanupStepDrop", m.step)
	}
}

// === Cleanup updateProgress ===

func TestCleanupModel_UpdateProgress_TaskDone(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	m.step = cleanupStepProgress
	m.repos = []string{"/tmp/repo"}
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "repo", Path: "/tmp/repo", Status: components.TaskRunning},
	})
	m, _ = m.Update(cleanupTaskDoneMsg{index: 0, result: ops.CleanupResult{Success: true, PrunedBranches: 2}})
}

func TestCleanupModel_UpdateProgress_AllDone(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	m.step = cleanupStepProgress
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "repo", Path: "/tmp/repo", Status: components.TaskDone},
	})
	m, _ = m.Update(components.AllTasksDoneMsg{})
	if m.step != cleanupStepResults {
		t.Errorf("step = %d, want cleanupStepResults", m.step)
	}
}

// === Commit updateMessage (91.3%) ===

func TestCommitModel_UpdateMessage_Enter(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = commitStepMessage
	m.messageInput.SetValue("fix: something")
	m, _ = m.Update(enterMsg())
	if m.step != commitStepBranch {
		t.Errorf("step = %d, want commitStepBranch", m.step)
	}
}

func TestCommitModel_UpdateMessage_EmptyEnter(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = commitStepMessage
	m, cmd := m.Update(enterMsg())
	if cmd != nil {
		t.Error("expected nil cmd for empty message")
	}
}

func TestCommitModel_UpdateMessage_Esc(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = commitStepMessage
	m, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected BackToMenuMsg cmd")
	}
}

func TestCommitModel_UpdateMessage_HistoryUpDown(t *testing.T) {
	recent := []string{"msg1", "msg2"}
	m := NewCommitModel(recent)
	m, _ = m.Update(wsMsg())
	m.step = commitStepMessage
	m.messageInput.SetValue("current")

	m, _ = m.Update(upMsg())
	if m.history.cursor != 0 {
		t.Errorf("history.cursor = %d, want 0", m.history.cursor)
	}

	m, _ = m.Update(downMsg())
	if m.history.cursor != -1 {
		t.Errorf("history.cursor = %d, want -1", m.history.cursor)
	}
}

// === Commit updateBranch ===

func TestCommitModel_UpdateBranch_Enter(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = commitStepBranch
	m.branchInput.SetValue("feature")
	m, _ = m.Update(enterMsg())
	if m.step != commitStepRepoSelect {
		t.Errorf("step = %d, want commitStepRepoSelect", m.step)
	}
}

func TestCommitModel_UpdateBranch_Esc(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = commitStepBranch
	m, _ = m.Update(escMsg())
	if m.step != commitStepMessage {
		t.Errorf("step = %d, want commitStepMessage", m.step)
	}
}

// === Commit updateRepoSelect ===

func TestCommitModel_UpdateRepoSelect_Done(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = commitStepRepoSelect
	m.message = "fix: test"
	m.branch = "feature"
	m, cmd := m.Update(RepoSelectDoneMsg{Paths: []string{"/tmp/repo"}})
	if m.step != commitStepProgress {
		t.Errorf("step = %d, want commitStepProgress", m.step)
	}
	if cmd == nil {
		t.Error("expected batch cmd")
	}
}

// === Commit updateProgress ===

func TestCommitModel_UpdateProgress_TaskDone_Fail(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = commitStepProgress
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "repo", Path: "/tmp/repo", Status: components.TaskRunning},
	})
	m, _ = m.Update(commitTaskDoneMsg{index: 0, result: ops.CommitResult{Success: false, Error: "fail"}})
}

// === Clone updateInput ===

func TestCloneModel_UpdateInput_TabNavigation(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.step = cloneStepInput
	// Tab through inputs
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.focusIndex != 1 {
		t.Errorf("inputFocus after tab = %d, want 1", m.focusIndex)
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
	if m.focusIndex != 2 {
		t.Errorf("inputFocus after 2nd tab = %d, want 2", m.focusIndex)
	}
}

func TestCloneModel_UpdateInput_Enter(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.step = cloneStepInput
	m.repoInput.SetValue("git@github.com:org/repo.git")
	m, cmd := m.Update(enterMsg())
	if cmd == nil {
		t.Error("expected startClone cmd")
	}
}

func TestCloneModel_UpdateInput_Esc(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.step = cloneStepInput
	m, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected BackToMenuMsg cmd")
	}
}

// === Clone updateProgress ===

func TestCloneModel_UpdateProgress_OrgFetchError(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.step = cloneStepProgress
	m, _ = m.Update(cloneOrgFetchedMsg{err: fmt.Errorf("mock error")})
	if m.step != cloneStepResults {
		t.Errorf("step = %d, want cloneStepResults", m.step)
	}
}

func TestCloneModel_UpdateProgress_OrgFetchSuccess(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.step = cloneStepProgress
	m.pathInput.SetValue("/tmp/dest")
	repos := []ops.Repository{
		{Name: "repo1", SSHURL: "git@github.com:org/repo1.git"},
		{Name: "repo2", SSHURL: "git@github.com:org/repo2.git"},
	}
	m, cmd := m.Update(cloneOrgFetchedMsg{repos: repos})
	if cmd == nil {
		t.Error("expected batch cmd for org repos")
	}
}

func TestCloneModel_UpdateProgress_TaskDoneSkipped(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.step = cloneStepProgress
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "repo", Path: "/tmp/repo", Status: components.TaskRunning},
	})
	m, _ = m.Update(cloneTaskDoneMsg{index: 0, result: ops.CloneResult{RepoName: "repo", Success: true, Skipped: true}})
}

func TestCloneModel_UpdateProgress_TaskDoneFailed(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.step = cloneStepProgress
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "repo", Path: "/tmp/repo", Status: components.TaskRunning},
	})
	m, _ = m.Update(cloneTaskDoneMsg{index: 0, result: ops.CloneResult{RepoName: "repo", Success: false, Error: "clone failed"}})
}

func TestCloneModel_UpdateProgress_AllDone(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.step = cloneStepProgress
	m.progress = components.NewProgressModel(nil)
	m, _ = m.Update(components.AllTasksDoneMsg{})
	if m.step != cloneStepResults {
		t.Errorf("step = %d, want cloneStepResults", m.step)
	}
}

func TestCloneModel_UpdateProgress_TaskDoneSuccess(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.step = cloneStepProgress
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "repo", Path: "/tmp/repo", Status: components.TaskRunning},
	})
	m, _ = m.Update(cloneTaskDoneMsg{index: 0, result: ops.CloneResult{RepoName: "repo", Success: true}})
}

// === EnableWorkflows updateInput ===

func TestEnableWorkflowsModel_UpdateInput_TabNav(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.step = enableWFStepInput
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
}

func TestEnableWorkflowsModel_UpdateInput_Enter(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.step = enableWFStepInput
	m.orgInput.SetValue("test-org")
	m, cmd := m.Update(enterMsg())
	if cmd == nil {
		t.Error("expected fetch cmd")
	}
}

// === EnableWorkflows updateFetching ===

func TestEnableWorkflowsModel_UpdateFetching_ReposFetched(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.step = enableWFStepFetching
	m, cmd := m.Update(enableWFReposFetchedMsg{repos: []string{"repo1"}, err: nil})
	if cmd == nil {
		t.Error("expected startTasks cmd")
	}
}

func TestEnableWorkflowsModel_UpdateFetching_Error(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.step = enableWFStepFetching
	m, _ = m.Update(enableWFReposFetchedMsg{err: fmt.Errorf("mock error")})
	if m.step != enableWFStepResults {
		t.Errorf("step = %d, want enableWFStepResults", m.step)
	}
}

func TestEnableWorkflowsModel_UpdateFetching_SpinnerTick(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.step = enableWFStepFetching
	m, _ = m.Update(spinner.TickMsg{})
}

func TestEnableWorkflowsModel_UpdateFetching_Esc(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.step = enableWFStepFetching
	m, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected BackToMenuMsg cmd")
	}
}

// === EnableWorkflows updateProgress ===

func TestEnableWorkflowsModel_UpdateProgress_TaskDone_Success(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.step = enableWFStepProgress
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "repo", Status: components.TaskRunning},
	})
	m, _ = m.Update(enableWFTaskDoneMsg{index: 0, result: ops.EnableWorkflowResult{Success: true, EnabledCount: 2, RetriggeredPRs: 1}})
}

func TestEnableWorkflowsModel_UpdateProgress_TaskDone_Fail(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.step = enableWFStepProgress
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "repo", Status: components.TaskRunning},
	})
	m, _ = m.Update(enableWFTaskDoneMsg{index: 0, result: ops.EnableWorkflowResult{Success: false, Error: "api error"}})
}

func TestEnableWorkflowsModel_UpdateProgress_TaskDone_NoAction(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.step = enableWFStepProgress
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "repo", Status: components.TaskRunning},
	})
	// Success but nothing was enabled or retriggered
	m, _ = m.Update(enableWFTaskDoneMsg{index: 0, result: ops.EnableWorkflowResult{Success: true, EnabledCount: 0, RetriggeredPRs: 0}})
}

func TestEnableWorkflowsModel_UpdateProgress_AllDone(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.step = enableWFStepProgress
	m.progress = components.NewProgressModel(nil)
	m, _ = m.Update(components.AllTasksDoneMsg{})
	if m.step != enableWFStepResults {
		t.Errorf("step = %d, want enableWFStepResults", m.step)
	}
}

// === TeamPRs updateInput ===

func TestTeamPRsModel_UpdateInput_Enter(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = teamPRsStepInput
	m.orgInput.SetValue("test-org")
	m.teamInput.SetValue("test-team")
	m, cmd := m.Update(enterMsg())
	if cmd == nil {
		t.Error("expected fetch cmd")
	}
}

func TestTeamPRsModel_UpdateInput_Esc(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = teamPRsStepInput
	m, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected BackToMenuMsg cmd")
	}
}

func TestTeamPRsModel_UpdateInput_Tab(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = teamPRsStepInput
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyTab})
}

// === TeamPRs updateFetching ===

func TestTeamPRsModel_UpdateFetching_ReposFetched(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = teamPRsStepFetching
	m, cmd := m.Update(teamPRsReposFetchedMsg{repos: []string{"repo1"}, err: nil})
	if cmd == nil {
		t.Error("expected fetch PRs cmd")
	}
}

func TestTeamPRsModel_UpdateFetching_ReposError(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = teamPRsStepFetching
	m, _ = m.Update(teamPRsReposFetchedMsg{err: fmt.Errorf("mock error")})
	if m.step != teamPRsStepResults {
		t.Errorf("step = %d, want teamPRsStepResults", m.step)
	}
}

func TestTeamPRsModel_UpdateFetching_PRsFetched(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = teamPRsStepFetching
	m, _ = m.Update(teamPRsFetchedMsg{prs: []ops.TeamPR{{Repo: "r", Number: 1, Title: "t", Author: "a"}}})
	if m.step != teamPRsStepResults {
		t.Errorf("step = %d, want teamPRsStepResults", m.step)
	}
}

func TestTeamPRsModel_UpdateFetching_PRsFetchedError(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = teamPRsStepFetching
	m, _ = m.Update(teamPRsFetchedMsg{err: fmt.Errorf("mock error")})
	if m.step != teamPRsStepResults {
		t.Errorf("step = %d, want teamPRsStepResults", m.step)
	}
}

func TestTeamPRsModel_UpdateFetching_SpinnerTick(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = teamPRsStepFetching
	m, _ = m.Update(spinner.TickMsg{})
}

func TestTeamPRsModel_UpdateFetching_Esc(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = teamPRsStepFetching
	m, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected BackToMenuMsg cmd")
	}
}

// === Status updateProgress ===

func TestStatusModel_UpdateProgress_TaskDone(t *testing.T) {
	m := NewStatusModel()
	m, _ = m.Update(wsMsg())
	m.step = statusStepProgress
	m.results = make([]ops.StatusResult, 1)
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "repo", Path: "/tmp/repo", Status: components.TaskRunning},
	})
	m, _ = m.Update(statusTaskDoneMsg{index: 0, result: ops.StatusResult{RepoName: "repo", Branch: "main"}})
}

func TestStatusModel_UpdateProgress_AllDone(t *testing.T) {
	m := NewStatusModel()
	m, _ = m.Update(wsMsg())
	m.step = statusStepProgress
	m.results = make([]ops.StatusResult, 0)
	m.progress = components.NewProgressModel(nil)
	m, _ = m.Update(components.AllTasksDoneMsg{})
	if m.step != statusStepResults {
		t.Errorf("step = %d, want statusStepResults", m.step)
	}
}

// === Branches updateProgress ===

func TestBranchesModel_UpdateProgress_TaskDone(t *testing.T) {
	m := NewBranchesModel()
	m, _ = m.Update(wsMsg())
	m.step = branchesStepProgress
	m.results = make([]ops.BranchesResult, 1)
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "repo", Path: "/tmp/repo", Status: components.TaskRunning},
	})
	m, _ = m.Update(branchesTaskDoneMsg{index: 0, result: ops.BranchesResult{RepoName: "repo", CurrentBranch: "main"}})
}

func TestBranchesModel_UpdateProgress_AllDone(t *testing.T) {
	m := NewBranchesModel()
	m, _ = m.Update(wsMsg())
	m.step = branchesStepProgress
	m.results = make([]ops.BranchesResult, 0)
	m.progress = components.NewProgressModel(nil)
	m, _ = m.Update(components.AllTasksDoneMsg{})
	if m.step != branchesStepResults {
		t.Errorf("step = %d, want branchesStepResults", m.step)
	}
}

// === MyPRs updateFetching ===

func TestMyPRsModel_UpdateFetching_Fetched(t *testing.T) {
	m := NewMyPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = myPRsStepFetching
	m, _ = m.Update(myPRsFetchedMsg{prs: []ops.MyPR{{Repo: "r", Number: 1, Title: "t", State: "open"}}})
	if m.step != myPRsStepResults {
		t.Errorf("step = %d, want myPRsStepResults", m.step)
	}
}

func TestMyPRsModel_UpdateFetching_Error(t *testing.T) {
	m := NewMyPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = myPRsStepFetching
	m, _ = m.Update(myPRsFetchedMsg{err: fmt.Errorf("mock error")})
	if m.step != myPRsStepResults {
		t.Errorf("step = %d, want myPRsStepResults", m.step)
	}
}

func TestMyPRsModel_UpdateFetching_SpinnerTick(t *testing.T) {
	m := NewMyPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = myPRsStepFetching
	m, _ = m.Update(spinner.TickMsg{})
}

func TestMyPRsModel_UpdateFetching_Esc(t *testing.T) {
	m := NewMyPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = myPRsStepFetching
	m, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected BackToMenuMsg cmd")
	}
}

// === View calls for various screens ===

func TestCleanupModel_View_AllSteps(t *testing.T) {
	for _, step := range []cleanupStep{cleanupStepDrop, cleanupStepRepoSelect, cleanupStepProgress, cleanupStepResults} {
		m := NewCleanupModel()
		m, _ = m.Update(wsMsg())
		m.step = step
		if step == cleanupStepProgress {
			m.progress = components.NewProgressModel(nil)
		}
		if step == cleanupStepRepoSelect {
			m.repoSelect = NewRepoSelectModel("cleanup", ".", 5, 30)
		}
		if step == cleanupStepResults {
			m.results = components.NewResultModel("Cleanup", nil)
			m.viewport.SetContent(m.results.View())
		}
		v := m.View()
		if v == "" {
			t.Errorf("View() empty for step %d", step)
		}
	}
}

func TestCommitModel_View_AllSteps(t *testing.T) {
	for _, step := range []commitStep{commitStepMessage, commitStepBranch, commitStepRepoSelect, commitStepProgress, commitStepResults} {
		m := NewCommitModel([]string{"recent"})
		m, _ = m.Update(wsMsg())
		m.step = step
		m.message = "fix: x"
		m.branch = "feat"
		if step == commitStepProgress {
			m.progress = components.NewProgressModel(nil)
		}
		if step == commitStepRepoSelect {
			m.repoSelect = NewRepoSelectModel("commit", ".", 5, 30)
		}
		if step == commitStepResults {
			m.results = components.NewResultModel("Commit", nil)
			m.viewport.SetContent(m.results.View())
		}
		v := m.View()
		if v == "" {
			t.Errorf("View() empty for step %d", step)
		}
	}
}

// === updateResults with q key ===

func TestUpdateResults_QKey_AllScreens(t *testing.T) {
	qMsg := tea.KeyPressMsg{Code: 'q'}

	// Test q key for screens that accept it in updateResults
	{
		m := NewMergePRsModel()
		m, _ = m.Update(wsMsg())
		m.step = mergePRsStepResults
		m, cmd := m.Update(qMsg)
		if cmd == nil {
			t.Error("mergePRs updateResults: expected BackToMenuMsg cmd for q")
		}
	}
	{
		m := NewTeamPRsModel()
		m, _ = m.Update(wsMsg())
		m.step = teamPRsStepResults
		m, cmd := m.Update(qMsg)
		if cmd == nil {
			t.Error("teamPRs updateResults: expected BackToMenuMsg cmd for q")
		}
	}
	{
		m := NewMyPRsModel()
		m, _ = m.Update(wsMsg())
		m.step = myPRsStepResults
		m, cmd := m.Update(qMsg)
		if cmd == nil {
			t.Error("myPRs updateResults: expected BackToMenuMsg cmd for q")
		}
	}
}

// === updateResults with non-matching key (covers the "no match" path in switch) ===

func TestUpdateResults_UnmatchedKey_AllScreens(t *testing.T) {
	// Send a key that doesn't match esc/q to exercise the KeyPressMsg case with no match
	xMsg := tea.KeyPressMsg{Code: 'x'}

	screens := []struct {
		name string
		fn   func()
	}{
		{"pr", func() {
			m := NewPullRequestModel(nil)
			m, _ = m.Update(wsMsg())
			m.step = prStepResults
			m.Update(xMsg)
		}},
		{"commit", func() {
			m := NewCommitModel(nil)
			m, _ = m.Update(wsMsg())
			m.step = commitStepResults
			m.Update(xMsg)
		}},
		{"cleanup", func() {
			m := NewCleanupModel()
			m, _ = m.Update(wsMsg())
			m.step = cleanupStepResults
			m.Update(xMsg)
		}},
		{"automerge", func() {
			m := NewAutomergeModel()
			m, _ = m.Update(wsMsg())
			m.step = automergeStepResults
			m.Update(xMsg)
		}},
		{"branches", func() {
			m := NewBranchesModel()
			m, _ = m.Update(wsMsg())
			m.step = branchesStepResults
			m.Update(xMsg)
		}},
		{"status", func() {
			m := NewStatusModel()
			m, _ = m.Update(wsMsg())
			m.step = statusStepResults
			m.Update(xMsg)
		}},
		{"clone", func() {
			m := NewCloneModel()
			m, _ = m.Update(wsMsg())
			m.step = cloneStepResults
			m.Update(xMsg)
		}},
		{"enableWF", func() {
			m := NewEnableWorkflowsModel()
			m, _ = m.Update(wsMsg())
			m.step = enableWFStepResults
			m.Update(xMsg)
		}},
		{"mergePRs", func() {
			m := NewMergePRsModel()
			m, _ = m.Update(wsMsg())
			m.step = mergePRsStepResults
			m.Update(xMsg)
		}},
		{"teamPRs", func() {
			m := NewTeamPRsModel()
			m, _ = m.Update(wsMsg())
			m.step = teamPRsStepResults
			m.Update(xMsg)
		}},
		{"myPRs", func() {
			m := NewMyPRsModel()
			m, _ = m.Update(wsMsg())
			m.step = myPRsStepResults
			m.Update(xMsg)
		}},
	}

	for _, s := range screens {
		t.Run(s.name, func(t *testing.T) {
			s.fn()
		})
	}
}

// === HistoryBrowser edge cases (via PullRequestModel) ===

func TestBrowseHistoryUp_AtMaxStaysAtMax(t *testing.T) {
	recent := []string{"a", "b"}
	m := NewPullRequestModel(recent)
	m.messageInput.SetValue("typed")
	m.history.BrowseUp(&m.messageInput)
	if m.history.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m.history.cursor)
	}
	m.history.BrowseUp(&m.messageInput)
	if m.history.cursor != 1 {
		t.Errorf("cursor = %d, want 1", m.history.cursor)
	}
	// At max, should stay at max
	m.history.BrowseUp(&m.messageInput)
	if m.history.cursor != 1 {
		t.Errorf("cursor = %d, want 1 (clamped)", m.history.cursor)
	}
}

// === viewChanges ===

func TestPRModel_ViewChanges_WithSelection(t *testing.T) {
	m := NewPullRequestModel(nil)
	m.message = "fix: test"
	m.branch = "feature"
	m.changeCursor = 1
	m.changeSelected[0] = true
	v := m.viewChanges()
	if v == "" {
		t.Error("viewChanges returned empty")
	}
}

// === collectSelectedChanges ===

func TestPRModel_CollectSelectedChanges(t *testing.T) {
	m := NewPullRequestModel(nil)
	m.changeSelected[0] = true
	m.changeSelected[2] = true
	m.collectSelectedChanges()
	if len(m.changes) != 2 {
		t.Errorf("len(changes) = %d, want 2", len(m.changes))
	}
}

// === Additional viewport-forwarding for esc/q in updateResults ===

func TestPRModel_UpdateResults_Esc(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepResults
	m, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected BackToMenuMsg cmd for esc in PR results")
	}
}

func TestCommitModel_UpdateResults_Esc(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = commitStepResults
	m, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected BackToMenuMsg cmd for esc in commit results")
	}
}

func TestCleanupModel_UpdateResults_Esc(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	m.step = cleanupStepResults
	m, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected BackToMenuMsg cmd for esc in cleanup results")
	}
}

func TestAutomergeModel_UpdateResults_Esc(t *testing.T) {
	m := NewAutomergeModel()
	m, _ = m.Update(wsMsg())
	m.step = automergeStepResults
	m, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected BackToMenuMsg cmd for esc in automerge results")
	}
}

func TestBranchesModel_UpdateResults_EscKey(t *testing.T) {
	m := NewBranchesModel()
	m, _ = m.Update(wsMsg())
	m.step = branchesStepResults
	m, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected BackToMenuMsg cmd for esc in branches results")
	}
}

func TestStatusModel_UpdateResults_EscKey(t *testing.T) {
	m := NewStatusModel()
	m, _ = m.Update(wsMsg())
	m.step = statusStepResults
	m, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected BackToMenuMsg cmd for esc in status results")
	}
}

func TestCloneModel_UpdateResults_Esc(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.step = cloneStepResults
	m, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected BackToMenuMsg cmd for esc in clone results")
	}
}

func TestEnableWorkflowsModel_UpdateResults_Esc(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.step = enableWFStepResults
	m, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected BackToMenuMsg cmd for esc in enableWF results")
	}
}

// === MergePRs updateFetching ===

func TestMergePRsModel_UpdateFetching_Fetched(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepFetching
	m, _ = m.Update(mergePRsFetchedMsg{prs: []ops.PRInfo{{Repo: "r", Number: 1, Title: "t"}}})
	if m.step != mergePRsStepProgress {
		t.Errorf("step = %d, want mergePRsStepProgress", m.step)
	}
}

func TestMergePRsModel_UpdateFetching_NoPRs(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepFetching
	m, _ = m.Update(mergePRsFetchedMsg{prs: nil})
	if m.step != mergePRsStepResults {
		t.Errorf("step = %d, want mergePRsStepResults", m.step)
	}
}

func TestMergePRsModel_UpdateFetching_Error(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepFetching
	m, _ = m.Update(mergePRsFetchedMsg{err: fmt.Errorf("mock error")})
	if m.step != mergePRsStepResults {
		t.Errorf("step = %d, want mergePRsStepResults", m.step)
	}
}

// === MergePRs updateProgress ===

func TestMergePRsModel_UpdateProgress_TaskDone(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepProgress
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "repo #1", Status: components.TaskRunning},
	})
	m, _ = m.Update(mergePRTaskDoneMsg{index: 0, result: ops.MergePRResult{Success: true}})
}

func TestMergePRsModel_UpdateProgress_TaskDoneFail(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepProgress
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "repo #1", Status: components.TaskRunning},
	})
	m, _ = m.Update(mergePRTaskDoneMsg{index: 0, result: ops.MergePRResult{Success: false, Error: "fail"}})
}

func TestMergePRsModel_UpdateProgress_AllDone_HasMore(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepProgress
	m.prs = []ops.PRInfo{{Repo: "r", Number: 1}, {Repo: "r2", Number: 2}}
	m.batchSize = 1
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "r #1", Status: components.TaskDone},
	})
	m, _ = m.Update(components.AllTasksDoneMsg{})
	// Should start waiting for next batch
}

func TestMergePRsModel_UpdateProgress_AllDone_NoMore(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepProgress
	m.prs = nil
	m.batchSize = 5
	m.progress = components.NewProgressModel(nil)
	m, _ = m.Update(components.AllTasksDoneMsg{})
	// Should go to results (no more PRs)
}

// === View with viewReady=false (covers the !viewReady path) ===

func TestBranchesModel_View_NotViewReady(t *testing.T) {
	m := NewBranchesModel()
	// Don't send WSM so viewReady stays false
	m.step = branchesStepResults
	m.results = make([]ops.BranchesResult, 0)
	v := m.View()
	if v == "" {
		t.Error("View returned empty")
	}
}

func TestStatusModel_View_NotViewReady(t *testing.T) {
	m := NewStatusModel()
	m.step = statusStepResults
	m.results = make([]ops.StatusResult, 0)
	v := m.View()
	if v == "" {
		t.Error("View returned empty")
	}
}

func TestBranchesModel_View_Progress(t *testing.T) {
	m := NewBranchesModel()
	m, _ = m.Update(wsMsg())
	m.step = branchesStepProgress
	m.progress = components.NewProgressModel(nil)
	v := m.View()
	if v == "" {
		t.Error("View returned empty")
	}
}

func TestStatusModel_View_Progress(t *testing.T) {
	m := NewStatusModel()
	m, _ = m.Update(wsMsg())
	m.step = statusStepProgress
	m.progress = components.NewProgressModel(nil)
	v := m.View()
	if v == "" {
		t.Error("View returned empty")
	}
}

// === key helper using key.Matches (to avoid relying only on Code) ===

func TestKeyMatches(t *testing.T) {
	// Verify our key messages work with key.Matches
	msg := enterMsg()
	if !key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
		t.Error("enterMsg doesn't match enter binding")
	}
	msg2 := escMsg()
	if !key.Matches(msg2, key.NewBinding(key.WithKeys("esc"))) {
		t.Error("escMsg doesn't match esc binding")
	}
}

// === WindowSizeMsg handling ===

func TestPRModel_WindowSizeMsg_Resize(t *testing.T) {
	m := NewPullRequestModel(nil)
	// First WSM sets viewport
	m, _ = m.Update(wsMsg())
	if !m.viewReady {
		t.Error("viewReady should be true after first WSM")
	}
	// Second WSM resizes
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
}

func TestCommitModel_WindowSizeMsg(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	if !m.viewReady {
		t.Error("viewReady should be true")
	}
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
}

func TestCleanupModel_WindowSizeMsg(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	if !m.viewReady {
		t.Error("viewReady should be true")
	}
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
}

// === Non-key messages falling through updateResults (the 85.7% gap) ===

func TestUpdateResults_NonKeyMsg_AllScreens(t *testing.T) {
	// Send a non-key message to exercise the viewport.Update fallthrough in updateResults
	// Using a custom struct that will be forwarded to viewport.Update
	type dummyMsg struct{}
	vMsg := dummyMsg{}

	{
		m := NewPullRequestModel(nil)
		m, _ = m.Update(wsMsg())
		m.step = prStepResults
		m, _ = m.Update(vMsg)
	}
	{
		m := NewCommitModel(nil)
		m, _ = m.Update(wsMsg())
		m.step = commitStepResults
		m, _ = m.Update(vMsg)
	}
	{
		m := NewCleanupModel()
		m, _ = m.Update(wsMsg())
		m.step = cleanupStepResults
		m, _ = m.Update(vMsg)
	}
	{
		m := NewAutomergeModel()
		m, _ = m.Update(wsMsg())
		m.step = automergeStepResults
		m, _ = m.Update(vMsg)
	}
	{
		m := NewBranchesModel()
		m, _ = m.Update(wsMsg())
		m.step = branchesStepResults
		m, _ = m.Update(vMsg)
	}
	{
		m := NewStatusModel()
		m, _ = m.Update(wsMsg())
		m.step = statusStepResults
		m, _ = m.Update(vMsg)
	}
	{
		m := NewCloneModel()
		m, _ = m.Update(wsMsg())
		m.step = cloneStepResults
		m, _ = m.Update(vMsg)
	}
	{
		m := NewEnableWorkflowsModel()
		m, _ = m.Update(wsMsg())
		m.step = enableWFStepResults
		m, _ = m.Update(vMsg)
	}
	{
		m := NewMergePRsModel()
		m, _ = m.Update(wsMsg())
		m.step = mergePRsStepResults
		m, _ = m.Update(vMsg)
	}
	{
		m := NewTeamPRsModel()
		m, _ = m.Update(wsMsg())
		m.step = teamPRsStepResults
		m, _ = m.Update(vMsg)
	}
	{
		m := NewMyPRsModel()
		m, _ = m.Update(wsMsg())
		m.step = myPRsStepResults
		m, _ = m.Update(vMsg)
	}
}
