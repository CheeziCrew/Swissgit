package screens

import (
	"fmt"
	"testing"

	"github.com/CheeziCrew/swissgit/ops"
	"github.com/CheeziCrew/swissgit/tui/components"
)

// --- Clone: updateProgress and updateResults paths ---

func TestCloneModel_OrgFetchError(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(ws())
	m.step = cloneStepProgress
	m, _ = m.Update(cloneOrgFetchedMsg{repos: nil, err: fmt.Errorf("org not found")})
	if m.step != cloneStepResults {
		t.Errorf("step = %d, want cloneStepResults", m.step)
	}
}

func TestCloneModel_OrgFetchSuccess(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(ws())
	m.step = cloneStepProgress
	repos := []ops.Repository{
		{Name: "repo1", SSHURL: "git@github.com:org/repo1.git"},
	}
	m, cmd := m.Update(cloneOrgFetchedMsg{repos: repos})
	_ = cmd
	// Should still be at progress step but with tasks set up
}

func TestCloneModel_TaskDoneSuccess(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(ws())
	m.step = cloneStepProgress
	result := ops.CloneResult{RepoName: "repo1", Success: true}
	m, _ = m.Update(cloneTaskDoneMsg{index: 0, result: result})
}

func TestCloneModel_TaskDoneSkipped(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(ws())
	m.step = cloneStepProgress
	result := ops.CloneResult{RepoName: "repo1", Success: true, Skipped: true}
	m, _ = m.Update(cloneTaskDoneMsg{index: 0, result: result})
}

func TestCloneModel_TaskDoneFailed(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(ws())
	m.step = cloneStepProgress
	result := ops.CloneResult{RepoName: "repo1", Success: false, Error: "clone failed"}
	m, _ = m.Update(cloneTaskDoneMsg{index: 0, result: result})
}

func TestCloneModel_AllTasksDone(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(ws())
	m.step = cloneStepProgress
	m, _ = m.Update(components.AllTasksDoneMsg{})
	if m.step != cloneStepResults {
		t.Errorf("step = %d, want cloneStepResults", m.step)
	}
}

func TestCloneModel_ResultsEsc(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(ws())
	m.step = cloneStepResults
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from esc at results")
	}
}

// --- EnableWorkflows: updateFetching, updateProgress, updateResults ---

func TestEnableWFModel_FetchingReposError(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(ws())
	m.step = enableWFStepFetching
	m, _ = m.Update(enableWFReposFetchedMsg{repos: nil, err: fmt.Errorf("org not found")})
	// Should handle error (go to results or show error)
}

func TestEnableWFModel_FetchingReposEmpty(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(ws())
	m.step = enableWFStepFetching
	m, _ = m.Update(enableWFReposFetchedMsg{repos: nil})
	// Empty repos should be handled
}

func TestEnableWFModel_FetchingReposSuccess(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(ws())
	m.step = enableWFStepFetching
	m, _ = m.Update(enableWFReposFetchedMsg{repos: []string{"repo1", "repo2"}})
	if m.step != enableWFStepProgress {
		t.Errorf("step = %d, want enableWFStepProgress", m.step)
	}
}

func TestEnableWFModel_FetchingEsc(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(ws())
	m.step = enableWFStepFetching
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from esc at fetching")
	}
}

func TestEnableWFModel_TaskDone(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(ws())
	m.step = enableWFStepProgress
	result := ops.EnableWorkflowResult{Repo: "repo1", Success: true}
	m, _ = m.Update(enableWFTaskDoneMsg{index: 0, result: result})
}

func TestEnableWFModel_AllTasksDone(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(ws())
	m.step = enableWFStepProgress
	m, _ = m.Update(components.AllTasksDoneMsg{})
	if m.step != enableWFStepResults {
		t.Errorf("step = %d, want enableWFStepResults", m.step)
	}
}

func TestEnableWFModel_ResultsEsc(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(ws())
	m.step = enableWFStepResults
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from esc at results")
	}
}

// --- MergePRs: updateFetching, updateProgress, updateResults ---

func TestMergePRsModel_FetchingError(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(ws())
	m.step = mergePRsStepFetching
	m, _ = m.Update(mergePRsFetchedMsg{prs: nil, err: fmt.Errorf("gh error")})
	// Should handle error
}

func TestMergePRsModel_FetchingEmpty(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(ws())
	m.step = mergePRsStepFetching
	m, _ = m.Update(mergePRsFetchedMsg{prs: nil})
	// Empty PRs should be handled
}

func TestMergePRsModel_FetchingSuccess(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(ws())
	m.step = mergePRsStepFetching
	prs := []ops.PRInfo{
		{Repo: "repo1", Number: 1, Title: "fix"},
	}
	m, _ = m.Update(mergePRsFetchedMsg{prs: prs})
	// Should set up batch
}

func TestMergePRsModel_FetchingEsc(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(ws())
	m.step = mergePRsStepFetching
	m, _ = m.Update(escMsg())
	if m.step != mergePRsStepResults {
		t.Errorf("step = %d, want mergePRsStepResults", m.step)
	}
}

func TestMergePRsModel_TaskDone(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(ws())
	m.step = mergePRsStepProgress
	result := ops.MergePRResult{Repo: "repo1", PRNumber: "1", Success: true}
	m, _ = m.Update(mergePRTaskDoneMsg{index: 0, result: result})
}

func TestMergePRsModel_AllTasksDone(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(ws())
	m.step = mergePRsStepProgress
	m, _ = m.Update(components.AllTasksDoneMsg{})
	// May go to waiting or results
}

func TestMergePRsModel_WaitingEsc(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(ws())
	m.step = mergePRsStepWaiting
	m, _ = m.Update(escMsg())
	if m.step != mergePRsStepResults {
		t.Errorf("step = %d, want mergePRsStepResults", m.step)
	}
}

func TestMergePRsModel_WaitingTick(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(ws())
	m.step = mergePRsStepWaiting
	m, _ = m.Update(mergeWaitTickMsg{})
	// Should handle tick
}

func TestMergePRsModel_WaitingEnter(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(ws())
	m.step = mergePRsStepWaiting
	m, _ = m.Update(enterMsg())
	// Should skip wait
}

func TestMergePRsModel_ResultsEsc(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(ws())
	m.step = mergePRsStepResults
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from esc at results")
	}
}

// --- TeamPRs: handleReposFetched ---

func TestTeamPRsModel_ReposFetchedSuccess(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(ws())
	m.step = teamPRsStepFetching
	m, _ = m.Update(teamPRsReposFetchedMsg{repos: []string{"repo1", "repo2"}})
	// Should start fetching PRs
}

func TestTeamPRsModel_ReposFetchedEmpty(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(ws())
	m.step = teamPRsStepFetching
	m, _ = m.Update(teamPRsReposFetchedMsg{repos: nil})
	// Empty repos should go to results
}

func TestTeamPRsModel_ReposFetchedError(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(ws())
	m.step = teamPRsStepFetching
	m, _ = m.Update(teamPRsReposFetchedMsg{repos: nil, err: fmt.Errorf("org error")})
	// Should show error
}

func TestTeamPRsModel_FetchingEsc(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(ws())
	m.step = teamPRsStepFetching
	_, cmd := m.Update(escMsg())
	if cmd == nil {
		t.Error("expected cmd from esc at fetching")
	}
}

// --- Cleanup: updateRepoSelect and updateResults ---

func TestCleanupModel_RepoSelectBackToMenu(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(ws())
	m.step = cleanupStepRepoSelect
	m, _ = m.Update(BackToMenuMsg{})
	if m.step != cleanupStepDrop {
		t.Errorf("step = %d, want cleanupStepDrop", m.step)
	}
}

func TestCleanupModel_RepoSelectDone(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(ws())
	m.step = cleanupStepRepoSelect
	m, cmd := m.Update(RepoSelectDoneMsg{Paths: []string{"/repo1"}})
	if m.step != cleanupStepProgress {
		t.Errorf("step = %d, want cleanupStepProgress", m.step)
	}
	_ = cmd
}

// --- Commit: updateRepoSelect ---

func TestCommitModel_RepoSelectBackToMenu(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(ws())
	m.step = commitStepRepoSelect
	m, _ = m.Update(BackToMenuMsg{})
	if m.step != commitStepBranch {
		t.Errorf("step = %d, want commitStepBranch", m.step)
	}
}

func TestCommitModel_RepoSelectDone(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(ws())
	m.step = commitStepRepoSelect
	m, cmd := m.Update(RepoSelectDoneMsg{Paths: []string{"/repo1"}})
	if m.step != commitStepProgress {
		t.Errorf("step = %d, want commitStepProgress", m.step)
	}
	_ = cmd
}

// --- PullRequest: updateChanges, updateBreaking, updateRepoSelect ---

func TestPRModel_ChangesEsc(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(ws())
	m.step = prStepChanges
	m, _ = m.Update(escMsg())
	if m.step != prStepBranch {
		t.Errorf("step = %d, want prStepBranch", m.step)
	}
}

func TestPRModel_BreakingEsc(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(ws())
	m.step = prStepBreaking
	m, _ = m.Update(escMsg())
	if m.step != prStepChanges {
		t.Errorf("step = %d, want prStepChanges", m.step)
	}
}

func TestPRModel_RepoSelectBackToMenu(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(ws())
	m.step = prStepRepoSelect
	m, _ = m.Update(BackToMenuMsg{})
	if m.step != prStepBreaking {
		t.Errorf("step = %d, want prStepBreaking", m.step)
	}
}

func TestPRModel_RepoSelectDone(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(ws())
	m.step = prStepRepoSelect
	m, cmd := m.Update(RepoSelectDoneMsg{Paths: []string{"/repo1"}})
	if m.step != prStepProgress {
		t.Errorf("step = %d, want prStepProgress", m.step)
	}
	_ = cmd
}
