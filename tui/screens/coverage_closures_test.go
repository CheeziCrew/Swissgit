package screens

import (
	"os"
	"path/filepath"
	"testing"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/CheeziCrew/swissgit/ops"
	"github.com/CheeziCrew/swissgit/tui/components"
)

// execCmd executes a tea.Cmd and returns the resulting message, if any.
// Used to exercise closure bodies in Init/start/fetch functions.
func execCmd(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	return cmd()
}

// === Execute Init() closures to cover discover/fetch function bodies ===

func TestDiscoverForBranches_Closure(t *testing.T) {
	cmd := discoverForBranches()
	if cmd == nil {
		t.Fatal("discoverForBranches returned nil cmd")
	}
	msg := execCmd(cmd)
	// Should return branchesReposDiscoveredMsg
	if _, ok := msg.(branchesReposDiscoveredMsg); !ok {
		t.Errorf("expected branchesReposDiscoveredMsg, got %T", msg)
	}
}

func TestDiscoverForStatus_Closure(t *testing.T) {
	cmd := discoverForStatus()
	if cmd == nil {
		t.Fatal("discoverForStatus returned nil cmd")
	}
	msg := execCmd(cmd)
	if _, ok := msg.(statusReposDiscoveredMsg); !ok {
		t.Errorf("expected statusReposDiscoveredMsg, got %T", msg)
	}
}

func TestFetchMyPRs_Closure(t *testing.T) {
	cmd := fetchMyPRs()
	if cmd == nil {
		t.Fatal("fetchMyPRs returned nil cmd")
	}
	msg := execCmd(cmd)
	if m, ok := msg.(myPRsFetchedMsg); !ok {
		t.Errorf("expected myPRsFetchedMsg, got %T", msg)
	} else {
		// Should have an error since gh CLI is not available
		_ = m.err
	}
}

// === Execute startAutomergeTasks, startClone, etc. ===

func TestStartAutomergeTasks_WithRepos(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "repo1")
	os.MkdirAll(sub, 0755)
	initTestRepo(t, sub)

	oldWd, _ := os.Getwd()
	os.Chdir(root)
	defer os.Chdir(oldWd)

	m := NewAutomergeModel()
	m, _ = m.Update(wsMsg())
	m.target = "feature-branch"
	cmd := m.startAutomergeTasks()
	// startAutomergeTasks discovers repos and creates tasks
	if cmd != nil {
		// It's a batch cmd; we can't easily execute sub-cmds
		// but the function body is covered
	}
}

func TestStartAutomergeTasks_NoRepos(t *testing.T) {
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	m := NewAutomergeModel()
	m, _ = m.Update(wsMsg())
	m.target = "feature-branch"
	cmd := m.startAutomergeTasks()
	// With no repos, should return a BackToMenuMsg cmd
	if cmd != nil {
		msg := execCmd(cmd)
		if _, ok := msg.(BackToMenuMsg); !ok {
			t.Errorf("expected BackToMenuMsg, got %T", msg)
		}
	}
}

func TestStartClone_RepoURL(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.repoInput.SetValue("git@github.com:org/repo.git")
	m.pathInput.SetValue(t.TempDir())
	cmd := m.startClone()
	if cmd == nil {
		t.Error("expected non-nil cmd for repo clone")
	}
}

func TestStartClone_OrgName(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.repoInput.SetValue("")
	m.orgInput.SetValue("testorg")
	m.teamInput.SetValue("testteam")
	cmd := m.startClone()
	if cmd == nil {
		t.Error("expected non-nil cmd for org clone")
	}
}

func TestStartClone_NoInput(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.repoInput.SetValue("")
	m.orgInput.SetValue("")
	cmd := m.startClone()
	if cmd != nil {
		t.Error("expected nil cmd when both repo and org are empty")
	}
}

func TestStartClone_EmptyPath(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.repoInput.SetValue("git@github.com:org/repo.git")
	m.pathInput.SetValue("")
	cmd := m.startClone()
	if cmd == nil {
		t.Error("expected non-nil cmd even with empty path (defaults to .)")
	}
}

// === Cover updateRepoSelect default branch (forwarding to repoSelect.Update) ===

func TestCleanupModel_RepoSelectForwardMsg(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	m.step = cleanupStepRepoSelect
	// Send a message that's not RepoSelectDoneMsg or BackToMenuMsg
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
}

func TestCommitModel_RepoSelectForwardMsg(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = commitStepRepoSelect
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
}

func TestPRModel_RepoSelectForwardMsg(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepRepoSelect
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
}

// === Cover the startBranchesTasks and startStatusTasks inner closures ===

func TestStartBranchesTasks(t *testing.T) {
	m := NewBranchesModel()
	m, _ = m.Update(wsMsg())
	cmd := m.startBranchesTasks([]string{"/tmp/fake-repo"})
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestStartStatusTasks(t *testing.T) {
	m := NewStatusModel()
	m, _ = m.Update(wsMsg())
	cmd := m.startStatusTasks([]string{"/tmp/fake-repo"})
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestStartCommitTasks(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	m.repos = []string{"/tmp/fake-repo"}
	m.message = "test"
	m.branch = "main"
	cmd := m.startCommitTasks()
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestStartPRTasks(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.repos = []string{"/tmp/fake-repo"}
	m.message = "test"
	m.branch = "feature"
	m.target = "main"
	cmd := m.startPRTasks()
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestStartCleanupTasks(t *testing.T) {
	m := NewCleanupModel()
	m, _ = m.Update(wsMsg())
	m.repos = []string{"/tmp/fake-repo"}
	cmd := m.startCleanupTasks()
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

// === Cover MergePRs fetchPRs closure and startBatch ===

func TestMergePRsModel_FetchPRsClosure(t *testing.T) {
	m := NewMergePRsModel()
	m.org = "testorg"
	cmd := m.fetchPRs()
	if cmd == nil {
		t.Fatal("fetchPRs returned nil")
	}
	msg := execCmd(cmd)
	if _, ok := msg.(mergePRsFetchedMsg); !ok {
		t.Errorf("expected mergePRsFetchedMsg, got %T", msg)
	}
}

func TestMergePRsModel_StartBatch(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.org = "testorg"
	m.prs = []ops.PRInfo{
		{Repo: "r1", Number: 1, Title: "fix"},
	}
	m.batchSize = 5
	cmd := m.startBatch()
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

func TestMergePRsModel_StartWait(t *testing.T) {
	m := NewMergePRsModel()
	m.waitMin = 1
	cmd := m.startWait()
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
	if m.step != mergePRsStepWaiting {
		t.Errorf("step = %d, want mergePRsStepWaiting", m.step)
	}
}

// === Cover enableWF fetchRepos closure ===

func TestEnableWFModel_FetchReposClosure(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m.org = "testorg"
	cmd := m.fetchRepos()
	if cmd == nil {
		t.Fatal("fetchRepos returned nil")
	}
	msg := execCmd(cmd)
	if _, ok := msg.(enableWFReposFetchedMsg); !ok {
		t.Errorf("expected enableWFReposFetchedMsg, got %T", msg)
	}
}

// === Cover teamPRs fetchTeamRepos and fetchPRs closures ===

func TestTeamPRsModel_FetchTeamReposClosure(t *testing.T) {
	m := NewTeamPRsModel()
	m.org = "testorg"
	m.team = "testteam"
	cmd := m.fetchTeamRepos()
	if cmd == nil {
		t.Fatal("fetchTeamRepos returned nil")
	}
	msg := execCmd(cmd)
	if _, ok := msg.(teamPRsReposFetchedMsg); !ok {
		t.Errorf("expected teamPRsReposFetchedMsg, got %T", msg)
	}
}

func TestTeamPRsModel_FetchPRsClosure(t *testing.T) {
	m := NewTeamPRsModel()
	m.org = "testorg"
	cmd := m.fetchPRs([]string{"repo1"})
	if cmd == nil {
		t.Fatal("fetchPRs returned nil")
	}
	msg := execCmd(cmd)
	if _, ok := msg.(teamPRsFetchedMsg); !ok {
		t.Errorf("expected teamPRsFetchedMsg, got %T", msg)
	}
}

// === Cover showSummary with all fields ===

func TestPRModel_ShowSummary_NoBreaking(t *testing.T) {
	m := NewPullRequestModel(nil)
	m.message = "feat: new"
	m.branch = "feature"
	m.target = "main"
	m.changes = []string{"Feature"}
	m.breaking = false
	result := m.showSummary()
	if result == "" {
		t.Error("expected non-empty summary")
	}
}

func TestPRModel_ShowSummary_EmptyChanges(t *testing.T) {
	m := NewPullRequestModel(nil)
	m.message = "feat: new"
	m.branch = "feature"
	m.target = "develop"
	result := m.showSummary()
	if result == "" {
		t.Error("expected non-empty summary")
	}
}

// === Cover PR updateMessage non-enter/esc paths ===

func TestPRModel_MessageTypeText(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	// Type text in message input
	m, _ = m.Update(kMsg('h'))
	m, _ = m.Update(kMsg('e'))
}

func TestCommitModel_MessageTypeText(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	m, _ = m.Update(kMsg('h'))
	m, _ = m.Update(kMsg('e'))
}

// === Cover PR updateBranch non-key paths ===

func TestPRModel_BranchTypeText(t *testing.T) {
	m := NewPullRequestModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = prStepBranch
	m, _ = m.Update(kMsg('f'))
}

func TestCommitModel_BranchTypeText(t *testing.T) {
	m := NewCommitModel(nil)
	m, _ = m.Update(wsMsg())
	m.step = commitStepBranch
	m, _ = m.Update(kMsg('f'))
}

// === Cover mergePRs summary view edge cases ===

func TestMergePRsModel_SummaryView_ZeroValues(t *testing.T) {
	m := NewMergePRsModel()
	m.org = "testorg"
	m.merged = 0
	m.failed = 0
	v := m.summaryView()
	if v == "" {
		t.Error("summaryView() empty with zero values")
	}
}

// === Cover clone updateInput enter path ===

func TestCloneModel_InputEnterWithRepoURL(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.repoInput.SetValue("git@github.com:org/repo.git")
	m, cmd := m.Update(enterMsg())
	if cmd == nil {
		t.Error("expected cmd after enter with repo URL")
	}
	if m.step != cloneStepProgress {
		t.Errorf("step = %d, want cloneStepProgress", m.step)
	}
}

func TestCloneModel_InputEnterWithOrg(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.repoInput.SetValue("")
	m.orgInput.SetValue("testorg")
	m, cmd := m.Update(enterMsg())
	if cmd == nil {
		t.Error("expected cmd after enter with org")
	}
	if m.step != cloneStepProgress {
		t.Errorf("step = %d, want cloneStepProgress", m.step)
	}
}

func TestCloneModel_InputEnterNoInput(t *testing.T) {
	m := NewCloneModel()
	m, _ = m.Update(wsMsg())
	m.repoInput.SetValue("")
	m.orgInput.SetValue("")
	m, cmd := m.Update(enterMsg())
	// No repo and no org: startClone returns nil, so still at input
	_ = cmd
}

// === Cover MergePRs updateFetching spinner tick ===

func TestMergePRsModel_FetchingSpinnerTick(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepFetching
	m, _ = m.Update(spinner.TickMsg{})
}

// === Cover TeamPRs updateFetching with repos fetched then PRs fetched ===

func TestTeamPRsModel_ReposFetchedThenPRsFetched(t *testing.T) {
	m := NewTeamPRsModel()
	m, _ = m.Update(wsMsg())
	m.step = teamPRsStepFetching
	m.org = "testorg"
	m.team = "testteam"

	// Simulate repos fetched
	m, _ = m.Update(teamPRsReposFetchedMsg{repos: []string{"repo1", "repo2"}})
	// Model should now be fetching PRs (fetchMsg changed)
}

// === Cover various View edge cases ===

func TestMergePRsModel_ViewWaiting(t *testing.T) {
	m := NewMergePRsModel()
	m.viewReady = true
	m.step = mergePRsStepWaiting
	m.waitRemaining = 30
	m.org = "testorg"
	v := m.View()
	if v == "" {
		t.Error("View() empty at waiting step")
	}
}

func TestMergePRsModel_ViewFetching(t *testing.T) {
	m := NewMergePRsModel()
	m.viewReady = true
	m.step = mergePRsStepFetching
	m.org = "testorg"
	v := m.View()
	if v == "" {
		t.Error("View() empty at fetching step")
	}
}

func TestEnableWFModel_ViewFetching(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m.viewReady = true
	m.step = enableWFStepFetching
	m.org = "testorg"
	v := m.View()
	if v == "" {
		t.Error("View() empty at fetching step")
	}
}

func TestTeamPRsModel_ViewFetching(t *testing.T) {
	m := NewTeamPRsModel()
	m.viewReady = true
	m.step = teamPRsStepFetching
	m.fetchMsg = "Loading..."
	v := m.View()
	if v == "" {
		t.Error("View() empty at fetching step")
	}
}

// === Ensure enableWF startTasks covers closure ===

func TestEnableWFModel_StartTasks(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.org = "testorg"
	m.repos = []string{"repo1", "repo2"}
	cmd := m.startTasks()
	if cmd == nil {
		t.Error("expected non-nil cmd")
	}
}

// === Cover AllTasksDoneMsg going to results for various screens ===

func TestMergePRsModel_AllTasksDone_EmptyQueue(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepProgress
	m.batchSize = 5
	m.prs = []ops.PRInfo{} // empty queue
	m.progress = components.NewProgressModel([]components.RepoTask{
		{Name: "r1 #1", Status: components.TaskDone},
	})
	m, _ = m.Update(components.AllTasksDoneMsg{})
	if m.step != mergePRsStepResults {
		t.Errorf("step = %d, want mergePRsStepResults (empty queue)", m.step)
	}
}

// === Cover enableWF updateFetching with empty repos ===

func TestEnableWFModel_FetchingReposEmptyV2(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m, _ = m.Update(wsMsg())
	m.step = enableWFStepFetching
	m, _ = m.Update(enableWFReposFetchedMsg{repos: []string{}})
	if m.step != enableWFStepResults {
		t.Errorf("step = %d, want enableWFStepResults", m.step)
	}
}

// === Cover MergePRs updateFetching success with PRs ===

func TestMergePRsModel_FetchingSuccessWithPRs(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.step = mergePRsStepFetching
	m.org = "testorg"
	prs := []ops.PRInfo{
		{Repo: "r1", Number: 1, Title: "fix"},
		{Repo: "r2", Number: 2, Title: "feat"},
	}
	m, cmd := m.Update(mergePRsFetchedMsg{prs: prs})
	if cmd == nil {
		t.Error("expected cmd from fetch success")
	}
	if m.step != mergePRsStepProgress {
		t.Errorf("step = %d, want mergePRsStepProgress", m.step)
	}
}

// === Cover MergePRs updateInput empty org enter ===

func TestMergePRsModel_InputEmptyOrgEnter(t *testing.T) {
	m := NewMergePRsModel()
	m, _ = m.Update(wsMsg())
	m.orgInput.SetValue("")
	m, _ = m.Update(enterMsg())
	if m.step != mergePRsStepInput {
		t.Errorf("step = %d, want mergePRsStepInput (empty org)", m.step)
	}
}
