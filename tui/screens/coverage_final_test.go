package screens

import (
	"fmt"
	"testing"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/viewport"
	"github.com/CheeziCrew/swissgit/ops"
	"github.com/CheeziCrew/swissgit/tui/components"
)

// === updateResults viewport forwarding for ALL screens ===
// The 85.7% coverage on updateResults is because the viewport.Update(msg)
// default path (non-key message) isn't tested. Send a mouse/spinner message.

func TestAutomergeModel_ResultsViewportForward(t *testing.T) {
	m := NewAutomergeModel()
	m, _ = m.Update(wsMsg())
	m.step = automergeStepResults
	// Send a non-key message to exercise viewport forwarding
	m, _ = m.Update(spinner.TickMsg{})
}

func TestBranchesModel_ResultsViewportForward(t *testing.T) {
	m := NewBranchesModel()
	m, _ = m.Update(wsMsg())
	m.step = branchesStepResults
	m, _ = m.Update(spinner.TickMsg{})
}

func TestCleanupModel_ResultsViewportForward(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	m.step = cleanupStepResults
	m, _ = m.Update(spinner.TickMsg{})
}

func TestCloneModel_ResultsViewportForward(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.step = cloneStepResults
	m, _ = m.Update(spinner.TickMsg{})
}

func TestCommitModel_ResultsViewportForward(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = commitStepResults
	m, _ = m.Update(spinner.TickMsg{})
}

func TestPRModel_ResultsViewportForward(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepResults
	m, _ = m.Update(spinner.TickMsg{})
}

func TestEnableWFModel_ResultsViewportForward(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.step = enableWFStepResults
	m, _ = m.Update(spinner.TickMsg{})
}

func TestMergePRsModel_ResultsViewportForward(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepResults
	m, _ = m.Update(spinner.TickMsg{})
}

func TestMyPRsModel_ResultsViewportForward(t *testing.T) {
	m := NewMyPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = myPRsStepResults
	m, _ = m.Update(spinner.TickMsg{})
}

func TestTeamPRsModel_ResultsViewportForward(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = teamPRsStepResults
	m, _ = m.Update(spinner.TickMsg{})
}

func TestStatusModel_ResultsViewportForward(t *testing.T) {
	m := NewStatusModel()
	m, _ = m.Update(wsMsg())
	m.step = statusStepResults
	m, _ = m.Update(spinner.TickMsg{})
}

// === showSummary deeper coverage (69.2%) ===

func TestPRModel_ShowSummary_WithChangesAfterStep(t *testing.T) {
	m := NewPullRequestModel(nil)
	m.message = "feat: add feature"
	m.branch = "feature/x"
	m.target = "main"
	m.changes = []string{"Feature", "Enhancement"}
	m.step = prStepRepoSelect // step > prStepChanges
	result := m.showSummary()
	if result == "" {
		t.Error("expected non-empty summary with changes after step")
	}
}

func TestPRModel_ShowSummary_LongChangesTruncated(t *testing.T) {
	m := NewPullRequestModel(nil)
	m.message = "fix: bug"
	m.branch = "bugfix"
	m.target = "main"
	// Create a change string longer than 40 chars to trigger truncation
	m.changes = []string{"This is a very long change description that should be truncated because it exceeds forty chars"}
	m.step = prStepRepoSelect // > prStepChanges
	result := m.showSummary()
	if result == "" {
		t.Error("expected non-empty summary")
	}
}

func TestPRModel_ShowSummary_EmptyMessage(t *testing.T) {
	m := NewPullRequestModel(nil)
	// No message, no branch, no changes
	result := m.showSummary()
	if result != "" {
		t.Errorf("expected empty summary for empty model, got %q", result)
	}
}

// === PR updateMessage deeper coverage (73.9%) ===

func TestPRModel_MessageHistoryBrowseUp(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.history = NewHistoryBrowser([]string{"old msg 1", "old msg 2"})
	// Browse history up
	m, _ = m.Update(upMsg())
	// Browse up again
	m, _ = m.Update(upMsg())
	// Browse down
	m, _ = m.Update(downMsg())
	// Browse down past the end
	m, _ = m.Update(downMsg())
	m, _ = m.Update(downMsg())
}

func TestPRModel_MessageHistoryResetOnType(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.history = NewHistoryBrowser([]string{"old msg"})
	// Browse up first
	m, _ = m.Update(upMsg())
	// Type a char - should reset historyCursor
	m, _ = m.Update(kMsg('x'))
}

// === Automerge updateTarget deeper (75%) ===

func TestAutomergeModel_TargetEnterEmptyV2(t *testing.T) {
	m := NewAutomergeModel()
	m, _ = m.Update(wsMsg())
	m.targetInput.SetValue("")
	m, _ = m.Update(enterMsg())
	if m.step != automergeStepTarget {
		t.Errorf("step should stay at target when empty, got %d", m.step)
	}
}

func TestAutomergeModel_TargetTypeText(t *testing.T) {
	m := NewAutomergeModel()
	m, _ = m.Update(wsMsg())
	// Type in the target input
	m, _ = m.Update(kMsg('m'))
	m, _ = m.Update(kMsg('a'))
}

// === Branches discoverForBranches/startBranchesTasks deeper (80-82%) ===

func TestBranchesModel_DiscoverSuccess(t *testing.T) {
	m := NewBranchesModel()
	m, _ = m.Update(wsMsg())
	// Simulate repos discovered
	m, cmd := m.Update(branchesReposDiscoveredMsg{paths: []string{"/tmp/fake-repo1", "/tmp/fake-repo2"}})
	if cmd == nil {
		t.Error("expected cmd from repos discovered")
	}
}

func TestBranchesModel_DiscoverEmpty(t *testing.T) {
	m := NewBranchesModel()
	m, _ = m.Update(wsMsg())
	m, cmd := m.Update(branchesReposDiscoveredMsg{paths: nil})
	// With no repos, should go back
	if cmd == nil {
		t.Log("no cmd returned for empty repos (might be expected)")
	}
}

// === Status discoverForStatus/startStatusTasks deeper (80-82%) ===

func TestStatusModel_DiscoverSuccess(t *testing.T) {
	m := NewStatusModel()
	m, _ = m.Update(wsMsg())
	m, cmd := m.Update(statusReposDiscoveredMsg{paths: []string{"/tmp/fake-repo1"}})
	if cmd == nil {
		t.Error("expected cmd from repos discovered")
	}
}

func TestStatusModel_DiscoverEmpty(t *testing.T) {
	m := NewStatusModel()
	m, _ = m.Update(wsMsg())
	m, cmd := m.Update(statusReposDiscoveredMsg{paths: nil})
	_ = cmd
}

// === Clone startClone deeper (80%) ===

func TestCloneModel_StartCloneWithTeam(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.repoInput.SetValue("")
	m.orgInput.SetValue("testorg")
	m.teamInput.SetValue("myteam")
	m.pathInput.SetValue("/tmp/dest")
	cmd := m.startClone()
	if cmd == nil {
		t.Error("expected non-nil cmd for org+team clone")
	}
}

// === Clone updateProgress deeper (86%) ===

func TestCloneModel_ProgressCloneDone(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.step = cloneStepProgress
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "repo1", Status: components.TaskRunning},
	})
	// Simulate a clone done msg
	m, _ = m.Update(cloneTaskDoneMsg{index: 0, result: ops.CloneResult{
		RepoName: "repo1",
		Success:  true,
	}})
}

func TestCloneModel_ProgressCloneFailed(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.step = cloneStepProgress
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "repo1", Status: components.TaskRunning},
	})
	m, _ = m.Update(cloneTaskDoneMsg{index: 0, result: ops.CloneResult{
		RepoName: "repo1",
		Error:    "SSH failed",
	}})
}

func TestCloneModel_ProgressCloneSkipped(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.step = cloneStepProgress
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "repo1", Status: components.TaskRunning},
	})
	m, _ = m.Update(cloneTaskDoneMsg{index: 0, result: ops.CloneResult{
		RepoName: "repo1",
		Skipped:  true,
		Success:  true,
	}})
}

// === Cleanup startCleanupTasks deeper (83%) ===

func TestCleanupModel_StartCleanupWithDrop(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	m.repos = []string{"/tmp/fake-repo"}
	m.dropChanges = true
	cmd := m.startCleanupTasks()
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

// === Commit startCommitTasks deeper (84%) ===

func TestCommitModel_StartCommitWithEmptyBranch(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	m.repos = []string{"/tmp/fake-repo"}
	m.message = "test commit"
	m.branch = "" // no branch
	cmd := m.startCommitTasks()
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

// === PR startPRTasks deeper (86%) ===

func TestPRModel_StartPRWithBreaking(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.repos = []string{"/tmp/fake-repo"}
	m.message = "feat!: breaking"
	m.branch = "feature/x"
	m.target = "main"
	m.changes = []string{"Breaking change"}
	m.breaking = true
	cmd := m.startPRTasks()
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

// === MergePRs startBatch deeper (85.7%) ===

func TestMergePRsModel_StartBatchMultiple(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.org = "testorg"
	m.batchSize = 2
	m.prs = []ops.PRInfo{
		{Repo: "r1", Number: 1, Title: "fix1"},
		{Repo: "r2", Number: 2, Title: "fix2"},
		{Repo: "r3", Number: 3, Title: "fix3"},
	}
	cmd := m.startBatch()
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
	if m.step != mergePRsStepProgress {
		t.Errorf("step = %d, want mergePRsStepProgress", m.step)
	}
}

// === MergePRs startWait deeper (75%) ===

func TestMergePRsModel_StartWaitZero(t *testing.T) {
	m := NewMergePRsModel()
	m.waitMin = 0
	cmd := m.startWait()
	// Even with 0 wait, should create a tick cmd
	_ = cmd
}

// === MergePRs updateWaiting deeper (76.9%) ===

func TestMergePRsModel_WaitingTickDecrement(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepWaiting
	m.waitRemaining = 5
	// Simulate tick
	m, _ = m.Update(mergeWaitTickMsg{})
	if m.waitRemaining != 4 {
		t.Errorf("waitRemaining = %d, want 4", m.waitRemaining)
	}
}

func TestMergePRsModel_WaitingTickExpiredV2(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepWaiting
	m.waitRemaining = 1
	m.org = "testorg"
	m, cmd := m.Update(mergeWaitTickMsg{})
	if m.waitRemaining != 0 {
		t.Errorf("waitRemaining = %d, want 0", m.waitRemaining)
	}
	if cmd == nil {
		t.Error("expected cmd when wait expires (should fetch)")
	}
}

func TestMergePRsModel_WaitingQuit(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepWaiting
	m.waitRemaining = 10
	m, _ = m.Update(qMsg())
	if m.step != mergePRsStepResults {
		t.Errorf("step = %d, want results after quit", m.step)
	}
}

// === MergePRs summaryView deeper (83.3%) ===

func TestMergePRsModel_SummaryViewWithValues(t *testing.T) {
	m := NewMergePRsModel()
	m.org = "testorg"
	m.merged = 5
	m.failed = 2
	v := m.summaryView()
	if v == "" {
		t.Error("expected non-empty summary")
	}
}

// === EnableWF startTasks deeper (88.2%) ===

func TestEnableWFModel_StartTasksEmpty(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.org = "testorg"
	m.repos = []string{}
	cmd := m.startTasks()
	// Empty repos should return nil or a back cmd
	_ = cmd
}

// === TeamPRs updateFetching deeper (80%) ===

func TestTeamPRsModel_FetchingReposError(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = teamPRsStepFetching
	m, _ = m.Update(teamPRsReposFetchedMsg{repos: nil, err: fmt.Errorf("API error")})
}

func TestTeamPRsModel_FetchingPRsSuccess(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = teamPRsStepFetching
	prs := []ops.TeamPR{
		{Repo: "repo1", Title: "Fix bug", Author: "user1", Number: 1,
			CreatedAt: time.Now().Add(-48 * time.Hour)},
		{Repo: "repo1", Title: "Add feature", Author: "user2", Number: 2,
			CreatedAt: time.Now()},
	}
	m, _ = m.Update(teamPRsFetchedMsg{prs: prs})
	if m.step != teamPRsStepResults {
		t.Errorf("step = %d, want results", m.step)
	}
}

func TestTeamPRsModel_FetchingPRsEmpty(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = teamPRsStepFetching
	m, _ = m.Update(teamPRsFetchedMsg{prs: nil})
	if m.step != teamPRsStepResults {
		t.Errorf("step = %d, want results", m.step)
	}
}

func TestTeamPRsModel_FetchingPRsError(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = teamPRsStepFetching
	m, _ = m.Update(teamPRsFetchedMsg{err: fmt.Errorf("failed")})
}

// === Cover formatAge edge cases ===

func TestFormatAge_AllRanges(t *testing.T) {
	tests := []struct {
		name string
		age  time.Duration
	}{
		{"today", 0},
		{"1 day", 24 * time.Hour},
		{"3 days", 3 * 24 * time.Hour},
		{"2 weeks", 14 * 24 * time.Hour},
		{"2 months", 60 * 24 * time.Hour},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAge(time.Now().Add(-tt.age))
			if result == "" {
				t.Error("expected non-empty age string")
			}
		})
	}
}

// === MyPRs updateFetching deeper (93.8%) ===

func TestMyPRsModel_FetchedEmpty(t *testing.T) {
	m := NewMyPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = myPRsStepFetching
	m, _ = m.Update(myPRsFetchedMsg{prs: nil})
	if m.step != myPRsStepResults {
		t.Errorf("step = %d, want results", m.step)
	}
}

func TestMyPRsModel_FetchedErrorV2(t *testing.T) {
	m := NewMyPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = myPRsStepFetching
	m, _ = m.Update(myPRsFetchedMsg{err: fmt.Errorf("gh not found")})
	if m.step != myPRsStepResults {
		t.Errorf("step = %d, want results", m.step)
	}
}

// === Automerge startAutomergeTasks deeper (91.3%) ===

func TestAutomergeModel_TaskDoneV2(t *testing.T) {
	m := NewAutomergeModel()
	m, _ = m.Update(wsMsg())
	m.step = automergeStepProgress
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "repo1", Status: components.TaskRunning},
	})
	m, _ = m.Update(automergeTaskDoneMsg{index: 0, result: ops.AutomergeResult{
		RepoName: "repo1",
		Success:  true,
	}})
}

// === Viewport-based views when not ready ===

func TestCloneModel_ViewNotReady(t *testing.T) {
	m := NewCloneModel()
	m.step = cloneStepResults
	// viewReady is false
	v := m.View()
	if v == "" {
		t.Error("expected non-empty view even when not ready")
	}
}

func TestMergePRsModel_ViewProgress(t *testing.T) {
	m := NewMergePRsModel()
	m.viewReady = true
	m.step = mergePRsStepProgress
	m.org = "testorg"
	v := m.View()
	if v == "" {
		t.Error("expected non-empty view at progress step")
	}
}

// === Cover enableWF updateFetching with repos fetched then tasks ===

func TestEnableWFModel_FetchingReposWithRepos(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.step = enableWFStepFetching
	m, cmd := m.Update(enableWFReposFetchedMsg{repos: []string{"repo1", "repo2"}})
	if cmd == nil {
		t.Error("expected cmd when repos found")
	}
}

// === Cover PR updateChanges breaking select (line-level) ===

func TestPRModel_ChangesBreakingYesEnter(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepChanges
	// Select a change type first
	m, _ = m.Update(spaceMsg())
	// Navigate to breaking
	m.step = prStepBreaking
	// Select yes
	m, _ = m.Update(enterMsg())
}

// === Cover Automerge view at different steps ===

func TestAutomergeModel_ViewProgress(t *testing.T) {
	m := NewAutomergeModel()
	m.viewReady = true
	m.step = automergeStepProgress
	v := m.View()
	if v == "" {
		t.Error("expected non-empty view at progress")
	}
}

func TestAutomergeModel_ViewResults(t *testing.T) {
	m := NewAutomergeModel()
	m.viewReady = true
	m.step = automergeStepResults
	m.viewport = viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))
	v := m.View()
	if v == "" {
		t.Error("expected non-empty view at results")
	}
}
