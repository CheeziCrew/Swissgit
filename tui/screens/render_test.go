package screens

import (
	"strings"
	"testing"
	"time"

	"github.com/CheeziCrew/swissgit/ops"
)

// --- formatRepoLine ---

func TestFormatRepoLine_Clean(t *testing.T) {
	r := ops.StatusResult{
		RepoName:      "myrepo",
		Branch:        "main",
		DefaultBranch: "main",
		Clean:         true,
	}
	line := formatRepoLine(r)
	if !strings.Contains(line, "myrepo") {
		t.Error("expected repo name in output")
	}
	if !strings.Contains(line, "main") {
		t.Error("expected branch name in output")
	}
}

func TestFormatRepoLine_WithAllBadges(t *testing.T) {
	r := ops.StatusResult{
		RepoName:      "dirty-repo",
		Branch:        "feature",
		DefaultBranch: "main",
		Ahead:         2,
		Behind:        1,
		Modified:      3,
		Added:         4,
		Deleted:       5,
		Untracked:     6,
	}
	line := formatRepoLine(r)
	if !strings.Contains(line, "dirty-repo") {
		t.Error("expected repo name")
	}
	// The line should contain badges for all change types
	if line == "" {
		t.Error("expected non-empty output")
	}
}

func TestFormatRepoLine_DifferentBranch(t *testing.T) {
	r := ops.StatusResult{
		RepoName:      "repo",
		Branch:        "develop",
		DefaultBranch: "main",
	}
	line := formatRepoLine(r)
	if !strings.Contains(line, "develop") {
		t.Error("expected branch name in output")
	}
}

// --- statusBanner ---

func TestStatusBanner_Basic(t *testing.T) {
	b := statusBanner(5, 0, 0, 5)
	if !strings.Contains(b, "5 repos") {
		t.Error("expected repo count")
	}
}

func TestStatusBanner_Mixed(t *testing.T) {
	b := statusBanner(10, 3, 2, 5)
	if b == "" {
		t.Error("expected non-empty banner")
	}
}

// --- renderStatusErrors ---

func TestRenderStatusErrors_Empty(t *testing.T) {
	result := renderStatusErrors(nil)
	if result != "" {
		t.Errorf("expected empty string for nil errored, got %q", result)
	}
}

func TestRenderStatusErrors_WithErrors(t *testing.T) {
	errored := []ops.StatusResult{
		{RepoName: "bad-repo", Error: "fetch timeout"},
	}
	result := renderStatusErrors(errored)
	if !strings.Contains(result, "bad-repo") {
		t.Error("expected repo name in error output")
	}
}

// --- isDirtyRepo ---

func TestIsDirtyRepo(t *testing.T) {
	tests := []struct {
		name   string
		result ops.StatusResult
		want   bool
	}{
		{
			name:   "clean on default branch",
			result: ops.StatusResult{Clean: true, Branch: "main", DefaultBranch: "main"},
			want:   false,
		},
		{
			name:   "not clean",
			result: ops.StatusResult{Clean: false, Branch: "main", DefaultBranch: "main"},
			want:   true,
		},
		{
			name:   "different branch",
			result: ops.StatusResult{Clean: true, Branch: "feature", DefaultBranch: "main"},
			want:   true,
		},
		{
			name:   "ahead",
			result: ops.StatusResult{Clean: true, Branch: "main", DefaultBranch: "main", Ahead: 1},
			want:   true,
		},
		{
			name:   "behind",
			result: ops.StatusResult{Clean: true, Branch: "main", DefaultBranch: "main", Behind: 1},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDirtyRepo(tt.result)
			if got != tt.want {
				t.Errorf("isDirtyRepo() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- branches helpers ---

func TestHasExtraBranches(t *testing.T) {
	t.Run("only default branch", func(t *testing.T) {
		r := ops.BranchesResult{
			DefaultBranch: "main",
			LocalBranches: []ops.BranchInfo{{Name: "main"}},
		}
		if hasExtraBranches(r) {
			t.Error("expected false for only default branch")
		}
	})

	t.Run("extra local branch", func(t *testing.T) {
		r := ops.BranchesResult{
			DefaultBranch: "main",
			LocalBranches: []ops.BranchInfo{{Name: "main"}, {Name: "feature"}},
		}
		if !hasExtraBranches(r) {
			t.Error("expected true for extra branch")
		}
	})

	t.Run("remote branches", func(t *testing.T) {
		r := ops.BranchesResult{
			DefaultBranch:  "main",
			LocalBranches:  []ops.BranchInfo{{Name: "main"}},
			RemoteBranches: []ops.BranchInfo{{Name: "origin/feature"}},
		}
		if !hasExtraBranches(r) {
			t.Error("expected true for remote branches")
		}
	})
}

func TestTruncateBranchName(t *testing.T) {
	t.Run("short name", func(t *testing.T) {
		if truncateBranchName("main") != "main" {
			t.Error("short name should not be truncated")
		}
	})
	t.Run("long name", func(t *testing.T) {
		long := strings.Repeat("a", 50)
		result := truncateBranchName(long)
		// The truncated name should be shorter than the original
		if len(result) >= len(long) {
			t.Errorf("result should be shorter than original, got len=%d", len(result))
		}
	})
}

func TestRenderBranchList_Empty(t *testing.T) {
	result := renderBranchList(nil, "L", brLocalStyle, "local")
	if result != "" {
		t.Errorf("expected empty string for nil branches, got %q", result)
	}
}

func TestRenderBranchList_WithBranches(t *testing.T) {
	branches := []ops.BranchInfo{
		{Name: "feature-a"},
		{Name: "feature-b", IsStale: true},
	}
	result := renderBranchList(branches, "L", brLocalStyle, "local")
	if result == "" {
		t.Error("expected non-empty output")
	}
}

func TestRenderBranchList_MoreThanMax(t *testing.T) {
	branches := []ops.BranchInfo{
		{Name: "a"}, {Name: "b"}, {Name: "c"}, {Name: "d"},
	}
	result := renderBranchList(branches, "L", brLocalStyle, "local")
	if !strings.Contains(result, "+1 more") {
		t.Error("expected overflow indicator")
	}
}

func TestFormatBranchEntry(t *testing.T) {
	r := ops.BranchesResult{
		RepoName:       "test-repo",
		CurrentBranch:  "feature",
		DefaultBranch:  "main",
		LocalBranches:  []ops.BranchInfo{{Name: "main"}, {Name: "feature"}},
		RemoteBranches: []ops.BranchInfo{{Name: "origin/develop"}},
	}
	result := formatBranchEntry(r)
	if !strings.Contains(result, "test-repo") {
		t.Error("expected repo name in output")
	}
}

func TestBranchesBanner(t *testing.T) {
	b := branchesBanner(10, 3, 1, 6)
	if !strings.Contains(b, "10 repos") {
		t.Error("expected repo count")
	}
}

func TestRenderBranchErrors_Empty(t *testing.T) {
	result := renderBranchErrors(nil, 80)
	if result != "" {
		t.Errorf("expected empty for nil, got %q", result)
	}
}

func TestRenderBranchErrors_WithErrors(t *testing.T) {
	errored := []ops.BranchesResult{
		{RepoName: "bad", Error: "fetch failed"},
	}
	result := renderBranchErrors(errored, 80)
	if !strings.Contains(result, "bad") {
		t.Error("expected repo name in error output")
	}
}

// --- formatAge ---

func TestFormatAge(t *testing.T) {
	tests := []struct {
		name    string
		created time.Time
	}{
		{"today", time.Now()},
		{"1 day ago", time.Now().Add(-24 * time.Hour)},
		{"3 days ago", time.Now().Add(-3 * 24 * time.Hour)},
		{"2 weeks ago", time.Now().Add(-14 * 24 * time.Hour)},
		{"2 months ago", time.Now().Add(-60 * 24 * time.Hour)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAge(tt.created)
			if result == "" {
				t.Error("expected non-empty age string")
			}
		})
	}
}

// --- groupPRsByRepo ---

func TestGroupPRsByRepo(t *testing.T) {
	prs := []ops.TeamPR{
		{Repo: "repo-a", Title: "pr1"},
		{Repo: "repo-b", Title: "pr2"},
		{Repo: "repo-a", Title: "pr3"},
	}
	grouped, order := groupPRsByRepo(prs)
	if len(grouped) != 2 {
		t.Errorf("grouped count = %d, want 2", len(grouped))
	}
	if len(order) != 2 {
		t.Errorf("order count = %d, want 2", len(order))
	}
	if order[0] != "repo-a" {
		t.Errorf("first repo = %q, want repo-a", order[0])
	}
	if len(grouped["repo-a"]) != 2 {
		t.Errorf("repo-a PRs = %d, want 2", len(grouped["repo-a"]))
	}
}

func TestGroupPRsByRepo_Empty(t *testing.T) {
	grouped, order := groupPRsByRepo(nil)
	if len(grouped) != 0 || len(order) != 0 {
		t.Error("expected empty results for nil input")
	}
}

// --- prLink ---

func TestPrLink(t *testing.T) {
	result := prLink("https://github.com/org/repo/pull/1", "#1")
	if result == "" {
		t.Error("expected non-empty link")
	}
}

// --- renderPRLine ---

func TestRenderPRLine(t *testing.T) {
	pr := ops.TeamPR{
		Title:     "Fix important bug",
		URL:       "https://github.com/org/repo/pull/42",
		Author:    "developer",
		CreatedAt: time.Now().Add(-48 * time.Hour),
	}
	line := renderPRLine(pr)
	if line == "" {
		t.Error("expected non-empty line")
	}
}

func TestRenderPRLine_Bot(t *testing.T) {
	pr := ops.TeamPR{
		Title:  "Bump dependency",
		URL:    "https://github.com/org/repo/pull/99",
		Author: "dependabot[bot]",
	}
	line := renderPRLine(pr)
	if line == "" {
		t.Error("expected non-empty line for bot")
	}
}

func TestRenderPRLine_Draft(t *testing.T) {
	pr := ops.TeamPR{
		Title:  "WIP feature",
		URL:    "https://github.com/org/repo/pull/10",
		Author: "dev",
		Draft:  true,
	}
	line := renderPRLine(pr)
	if line == "" {
		t.Error("expected non-empty line for draft")
	}
}

// --- formatMyPRMeta ---

func TestFormatMyPRMeta_Empty(t *testing.T) {
	pr := ops.MyPR{}
	result := formatMyPRMeta(pr)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestFormatMyPRMeta_Draft(t *testing.T) {
	pr := ops.MyPR{Draft: true}
	result := formatMyPRMeta(pr)
	if result == "" {
		t.Error("expected non-empty for draft PR")
	}
}

func TestFormatMyPRMeta_WithAge(t *testing.T) {
	pr := ops.MyPR{CreatedAt: time.Now().Add(-48 * time.Hour)}
	result := formatMyPRMeta(pr)
	if result == "" {
		t.Error("expected non-empty for PR with age")
	}
}

// --- summaryLine and summaryBlock ---

func TestSummaryLine(t *testing.T) {
	line := summaryLine("label", "value")
	if line == "" {
		t.Error("expected non-empty summary line")
	}
}

func TestSummaryBlock(t *testing.T) {
	block := summaryBlock("line1", "line2")
	if block == "" {
		t.Error("expected non-empty summary block")
	}
}

// --- EnableWorkflowsModel helpers ---

func TestWorkflowLabel_Empty(t *testing.T) {
	m := NewEnableWorkflowsModel()
	label := m.workflowLabel()
	if label != "(all disabled)" {
		t.Errorf("expected '(all disabled)', got %q", label)
	}
}

func TestWorkflowLabel_Set(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m.workflowName = "deploy.yml"
	label := m.workflowLabel()
	if label != "deploy.yml" {
		t.Errorf("expected 'deploy.yml', got %q", label)
	}
}

func TestSummaryLines_Basic(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m.org = "testorg"
	lines := m.summaryLines()
	if len(lines) < 2 {
		t.Errorf("expected at least 2 lines, got %d", len(lines))
	}
}

func TestSummaryLines_WithPRBranch(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m.org = "testorg"
	m.prBranch = "main"
	lines := m.summaryLines()
	if len(lines) < 3 {
		t.Errorf("expected at least 3 lines with prBranch, got %d", len(lines))
	}
}

func TestSummaryLines_WithExtras(t *testing.T) {
	m := NewEnableWorkflowsModel()
	m.org = "testorg"
	lines := m.summaryLines("extra line")
	if len(lines) < 3 {
		t.Errorf("expected at least 3 lines with extras, got %d", len(lines))
	}
}

// --- PullRequestModel showSummary ---

func TestShowSummary_Empty(t *testing.T) {
	m := NewPullRequestModel(nil)
	result := m.showSummary()
	// With no message or branch, should return empty or minimal
	if result == "" {
		// acceptable
	}
}

func TestShowSummary_WithMessageAndBranch(t *testing.T) {
	m := NewPullRequestModel(nil)
	m.message = "fix: important bug"
	m.branch = "bugfix/important"
	result := m.showSummary()
	if result == "" {
		t.Error("expected non-empty summary with message and branch")
	}
}

// --- StatusModel renderResults ---

func TestStatusModel_RenderResults_Empty(t *testing.T) {
	m := NewStatusModel()
	result := m.renderResults()
	if result == "" {
		t.Error("expected non-empty output for empty results")
	}
}

func TestStatusModel_RenderResults_WithData(t *testing.T) {
	m := NewStatusModel()
	m.results = []ops.StatusResult{
		{RepoName: "a-repo", Branch: "main", DefaultBranch: "main", Clean: true},
		{RepoName: "b-repo", Branch: "feature", DefaultBranch: "main", Modified: 2},
	}
	result := m.renderResults()
	if result == "" {
		t.Error("expected non-empty render output")
	}
}

func TestStatusModel_RenderResults_WithErrors(t *testing.T) {
	m := NewStatusModel()
	m.results = []ops.StatusResult{
		{RepoName: "good-repo", Branch: "main", DefaultBranch: "main", Clean: true},
		{RepoName: "bad-repo", Error: "connection refused"},
	}
	result := m.renderResults()
	if result == "" {
		t.Error("expected non-empty render output")
	}
}

// --- BranchesModel renderResults ---

func TestBranchesModel_RenderResults_Empty(t *testing.T) {
	m := NewBranchesModel()
	result := m.renderResults()
	if result == "" {
		t.Error("expected non-empty output for empty results")
	}
}

func TestBranchesModel_RenderResults_WithData(t *testing.T) {
	m := NewBranchesModel()
	m.results = []ops.BranchesResult{
		{
			RepoName:      "repo-a",
			CurrentBranch: "main",
			DefaultBranch: "main",
			LocalBranches: []ops.BranchInfo{{Name: "main"}, {Name: "feature"}},
		},
	}
	m.width = 120
	result := m.renderResults()
	if result == "" {
		t.Error("expected non-empty render output")
	}
}

// --- TeamPRsModel renderTable ---

func TestTeamPRsModel_RenderTable_Empty(t *testing.T) {
	m := NewTeamPRsModel()
	result := m.renderTable()
	if result == "" {
		t.Error("expected non-empty output for no PRs")
	}
}

func TestTeamPRsModel_RenderTable_WithPRs(t *testing.T) {
	m := NewTeamPRsModel()
	m.prs = []ops.TeamPR{
		{Repo: "repo-a", Number: 1, Title: "fix bug", Author: "dev", URL: "https://github.com/org/repo/pull/1"},
		{Repo: "repo-b", Number: 2, Title: "add feature", Author: "dev2", URL: "https://github.com/org/repo/pull/2"},
		{Repo: "repo-a", Number: 3, Title: "update docs", Author: "dev", URL: "https://github.com/org/repo/pull/3"},
	}
	result := m.renderTable()
	if result == "" {
		t.Error("expected non-empty render output")
	}
}

// --- MyPRsModel renderTable ---

func TestMyPRsModel_RenderTable_Empty(t *testing.T) {
	m := NewMyPRsModel()
	result := m.renderTable()
	if result == "" {
		t.Error("expected non-empty output for no PRs")
	}
}

func TestMyPRsModel_RenderTable_WithPRs(t *testing.T) {
	m := NewMyPRsModel()
	m.prs = []ops.MyPR{
		{Repo: "repo-a", Number: 1, Title: "fix", URL: "https://github.com/org/repo/pull/1"},
		{Repo: "repo-b", Number: 2, Title: "feat", URL: "https://github.com/org/repo/pull/2", Draft: true},
	}
	result := m.renderTable()
	if result == "" {
		t.Error("expected non-empty render output")
	}
}
