package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/CheeziCrew/swissgit/ops"
)

// captureOutput captures stdout during fn execution.
func captureOutput(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestBuildCLI(t *testing.T) {
	root := BuildCLI()
	if root == nil {
		t.Fatal("BuildCLI returned nil")
	}
	if root.Use != "swissgit" {
		t.Errorf("root.Use = %q, want %q", root.Use, "swissgit")
	}

	// Check all subcommands are registered
	names := make(map[string]bool)
	for _, cmd := range root.Commands() {
		names[cmd.Use] = true
	}

	expected := []string{"pr", "commit", "cleanup", "status", "branches", "clone", "automerge", "merge-prs", "enable-workflows", "team-prs"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("missing subcommand %q", name)
		}
	}
}

func TestPrCmd(t *testing.T) {
	cmd := prCmd()
	if cmd.Use != "pr" {
		t.Errorf("Use = %q, want %q", cmd.Use, "pr")
	}
	if cmd.Short == "" {
		t.Error("Short description is empty")
	}

	msg := cmd.Flag("message")
	if msg == nil {
		t.Fatal("missing --message flag")
	}
	branch := cmd.Flag("branch")
	if branch == nil {
		t.Fatal("missing --branch flag")
	}
	target := cmd.Flag("target")
	if target == nil {
		t.Fatal("missing --target flag")
	}
	if target.DefValue != "main" {
		t.Errorf("target default = %q, want %q", target.DefValue, "main")
	}
	allFlag := cmd.Flag("all")
	if allFlag == nil {
		t.Fatal("missing --all flag")
	}
}

func TestCommitCLICmd(t *testing.T) {
	cmd := commitCLICmd()
	if cmd.Use != "commit" {
		t.Errorf("Use = %q, want %q", cmd.Use, "commit")
	}

	msg := cmd.Flag("message")
	if msg == nil {
		t.Fatal("missing --message flag")
	}
	branch := cmd.Flag("branch")
	if branch == nil {
		t.Fatal("missing --branch flag")
	}
	all := cmd.Flag("all")
	if all == nil {
		t.Fatal("missing --all flag")
	}
}

func TestCleanupCLICmd(t *testing.T) {
	cmd := cleanupCLICmd()
	if cmd.Use != "cleanup" {
		t.Errorf("Use = %q, want %q", cmd.Use, "cleanup")
	}

	drop := cmd.Flag("drop")
	if drop == nil {
		t.Fatal("missing --drop flag")
	}
	defaultBranch := cmd.Flag("default-branch")
	if defaultBranch == nil {
		t.Fatal("missing --default-branch flag")
	}
	if defaultBranch.DefValue != "main" {
		t.Errorf("default-branch default = %q, want %q", defaultBranch.DefValue, "main")
	}
	all := cmd.Flag("all")
	if all == nil {
		t.Fatal("missing --all flag")
	}
}

func TestStatusCLICmd(t *testing.T) {
	cmd := statusCLICmd()
	if cmd.Use != "status" {
		t.Errorf("Use = %q, want %q", cmd.Use, "status")
	}

	verbose := cmd.Flag("verbose")
	if verbose == nil {
		t.Fatal("missing --verbose flag")
	}
	all := cmd.Flag("all")
	if all == nil {
		t.Fatal("missing --all flag")
	}
}

func TestBranchesCLICmd(t *testing.T) {
	cmd := branchesCLICmd()
	if cmd.Use != "branches" {
		t.Errorf("Use = %q, want %q", cmd.Use, "branches")
	}

	all := cmd.Flag("all")
	if all == nil {
		t.Fatal("missing --all flag")
	}
}

func TestCloneCLICmd(t *testing.T) {
	cmd := cloneCLICmd()
	if cmd.Use != "clone" {
		t.Errorf("Use = %q, want %q", cmd.Use, "clone")
	}

	for _, flag := range []string{"repo", "org", "team", "path"} {
		if cmd.Flag(flag) == nil {
			t.Errorf("missing --%s flag", flag)
		}
	}
}

func TestCloneCLICmd_NoRepoOrOrg(t *testing.T) {
	cmd := cloneCLICmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no --repo or --org provided")
	}
	if err != nil && !strings.Contains(err.Error(), "specify --repo or --org") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAutomergeCLICmd(t *testing.T) {
	cmd := automergeCLICmd()
	if cmd.Use != "automerge" {
		t.Errorf("Use = %q, want %q", cmd.Use, "automerge")
	}

	target := cmd.Flag("target")
	if target == nil {
		t.Fatal("missing --target flag")
	}
	all := cmd.Flag("all")
	if all == nil {
		t.Fatal("missing --all flag")
	}
}

func TestMergePRsCLICmd(t *testing.T) {
	cmd := mergePRsCLICmd()
	if cmd.Use != "merge-prs" {
		t.Errorf("Use = %q, want %q", cmd.Use, "merge-prs")
	}

	for _, tc := range []struct {
		flag string
		def  string
	}{
		{"org", "Sundsvallskommun"},
		{"batch-size", "5"},
		{"wait", "10"},
	} {
		f := cmd.Flag(tc.flag)
		if f == nil {
			t.Fatalf("missing --%s flag", tc.flag)
		}
		if f.DefValue != tc.def {
			t.Errorf("--%s default = %q, want %q", tc.flag, f.DefValue, tc.def)
		}
	}

	dryRun := cmd.Flag("dry-run")
	if dryRun == nil {
		t.Fatal("missing --dry-run flag")
	}
}

func TestEnableWorkflowsCLICmd(t *testing.T) {
	cmd := enableWorkflowsCLICmd()
	if cmd.Use != "enable-workflows" {
		t.Errorf("Use = %q, want %q", cmd.Use, "enable-workflows")
	}

	for _, flag := range []string{"org", "workflow", "pr-branch"} {
		if cmd.Flag(flag) == nil {
			t.Errorf("missing --%s flag", flag)
		}
	}
}

func TestTeamPRsCLICmd(t *testing.T) {
	cmd := teamPRsCLICmd()
	if cmd.Use != "team-prs" {
		t.Errorf("Use = %q, want %q", cmd.Use, "team-prs")
	}

	for _, flag := range []string{"org", "team"} {
		if cmd.Flag(flag) == nil {
			t.Errorf("missing --%s flag", flag)
		}
	}
}

func TestResolvePaths_SinglePath(t *testing.T) {
	paths := resolvePaths(".", false)
	if len(paths) != 1 {
		t.Fatalf("expected 1 path, got %d", len(paths))
	}
	if paths[0] != "." {
		t.Errorf("path = %q, want %q", paths[0], ".")
	}
}

func TestResolvePaths_AllFlag_InvalidDir(t *testing.T) {
	paths := resolvePaths("/nonexistent/path/does/not/exist", true)
	if len(paths) != 1 {
		t.Fatalf("expected 1 fallback path, got %d", len(paths))
	}
}

func TestPrintResult(t *testing.T) {
	tests := []struct {
		name    string
		success bool
		info    string
		errMsg  string
	}{
		{"test-repo", true, "done", ""},
		{"test-repo", false, "", "something failed"},
		{"test-repo", true, "", ""},
		{"test-repo", false, "info", "error"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("success=%v/info=%q/err=%q", tt.success, tt.info, tt.errMsg), func(t *testing.T) {
			output := captureOutput(t, func() {
				printResult(tt.name, tt.success, tt.info, tt.errMsg)
			})
			if !strings.Contains(output, tt.name) {
				t.Errorf("output missing name %q: %s", tt.name, output)
			}
		})
	}
}

func TestPrintStatusResult(t *testing.T) {
	tests := []struct {
		name   string
		result ops.StatusResult
		checks []string
	}{
		{
			name: "clean repo",
			result: ops.StatusResult{
				RepoName: "my-repo",
				Branch:   "main",
			},
			checks: []string{"my-repo", "[main]"},
		},
		{
			name: "dirty repo with all badges",
			result: ops.StatusResult{
				RepoName:  "dirty-repo",
				Branch:    "feature",
				Ahead:     2,
				Behind:    3,
				Modified:  1,
				Added:     4,
				Deleted:   2,
				Untracked: 5,
			},
			checks: []string{"dirty-repo", "[feature]"},
		},
		{
			name: "only ahead",
			result: ops.StatusResult{
				RepoName: "ahead-repo",
				Branch:   "main",
				Ahead:    1,
			},
			checks: []string{"ahead-repo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(t, func() {
				printStatusResult(tt.result)
			})
			for _, check := range tt.checks {
				if !strings.Contains(output, check) {
					t.Errorf("output missing %q: %s", check, output)
				}
			}
		})
	}
}

func TestPrintNoPRsMessage(t *testing.T) {
	t.Run("first batch", func(t *testing.T) {
		output := captureOutput(t, func() {
			printNoPRsMessage(0)
		})
		if !strings.Contains(output, "No approved PRs found") {
			t.Errorf("unexpected output: %s", output)
		}
	})

	t.Run("subsequent batch", func(t *testing.T) {
		output := captureOutput(t, func() {
			printNoPRsMessage(1)
		})
		if !strings.Contains(output, "No more approved PRs") {
			t.Errorf("unexpected output: %s", output)
		}
	})
}

func TestProcessMergeBatch_DryRun(t *testing.T) {
	prs := []ops.PRInfo{
		{Repo: "repo-a", Number: 1, Title: "Fix bug"},
		{Repo: "repo-b", Number: 2, Title: "Add feature"},
	}

	cfg := mergeConfig{
		org:       "testorg",
		dryRun:    true,
		batchSize: 5,
		waitMin:   0,
	}

	output := captureOutput(t, func() {
		merged, failed := processMergeBatch(cfg, prs)
		if merged != 0 {
			t.Errorf("dry run merged = %d, want 0", merged)
		}
		if failed != 0 {
			t.Errorf("dry run failed = %d, want 0", failed)
		}
	})

	if !strings.Contains(output, "repo-a #1") {
		t.Errorf("output missing repo-a #1: %s", output)
	}
	if !strings.Contains(output, "repo-b #2") {
		t.Errorf("output missing repo-b #2: %s", output)
	}
}

func TestMergeConfig_Fields(t *testing.T) {
	cfg := mergeConfig{
		org:       "my-org",
		dryRun:    true,
		batchSize: 10,
		waitMin:   5,
	}
	if cfg.org != "my-org" {
		t.Errorf("org = %q, want %q", cfg.org, "my-org")
	}
	if !cfg.dryRun {
		t.Error("expected dryRun = true")
	}
	if cfg.batchSize != 10 {
		t.Errorf("batchSize = %d, want 10", cfg.batchSize)
	}
	if cfg.waitMin != 5 {
		t.Errorf("waitMin = %d, want 5", cfg.waitMin)
	}
}

func TestConstants(t *testing.T) {
	if descGitHubOrg == "" {
		t.Error("descGitHubOrg is empty")
	}
	if fmtStatusLine == "" {
		t.Error("fmtStatusLine is empty")
	}
	if descProcessRepos == "" {
		t.Error("descProcessRepos is empty")
	}
}

func TestStyleVars(t *testing.T) {
	// Just verify the style variables are initialized (non-empty renders)
	if ok == "" {
		t.Error("ok style is empty")
	}
	if fail == "" {
		t.Error("fail style is empty")
	}
}
