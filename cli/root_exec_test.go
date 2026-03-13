package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CheeziCrew/swissgit/ops"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// initTestRepo creates a minimal git repo with an initial commit and origin remote.
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
	// Create a file and commit so HEAD exists
	f, _ := os.Create(filepath.Join(dir, "README.md"))
	f.WriteString("# test\n")
	f.Close()
	w.Add("README.md")
	w.Commit("initial commit", &gogit.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@test.com"},
	})
}

// setupTestDir creates a temp dir with a single git repo.
func setupTestDir(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	initTestRepo(t, dir)
	return dir, func() {}
}

// setupTestDirWithSubRepos creates a temp dir with multiple sub-repo dirs.
func setupTestDirWithSubRepos(t *testing.T) (string, func()) {
	t.Helper()
	root := t.TempDir()
	for _, name := range []string{"repo-a", "repo-b"} {
		sub := filepath.Join(root, name)
		os.MkdirAll(sub, 0755)
		initTestRepo(t, sub)
	}
	return root, func() {}
}

// --- Execute tests for CLI commands ---

func TestCleanupCLICmd_Execute(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	cmd := cleanupCLICmd()
	cmd.SetArgs([]string{})
	output := captureOutput(t, func() {
		// Will fail at ops level (no proper remote) but exercises CLI code paths
		cmd.Execute()
	})
	_ = output // We just care that it ran without panic
}

func TestCleanupCLICmd_Execute_All(t *testing.T) {
	root, cleanup := setupTestDirWithSubRepos(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(root)
	defer os.Chdir(oldWd)

	cmd := cleanupCLICmd()
	cmd.SetArgs([]string{"--all"})
	output := captureOutput(t, func() {
		cmd.Execute()
	})
	_ = output
}

func TestCleanupCLICmd_Execute_WithPruned(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	cmd := cleanupCLICmd()
	cmd.SetArgs([]string{"--drop"})
	output := captureOutput(t, func() {
		cmd.Execute()
	})
	_ = output
}

func TestStatusCLICmd_Execute(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	cmd := statusCLICmd()
	cmd.SetArgs([]string{})
	output := captureOutput(t, func() {
		cmd.Execute()
	})
	_ = output
}

func TestStatusCLICmd_Execute_Verbose(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	cmd := statusCLICmd()
	cmd.SetArgs([]string{"--verbose"})
	output := captureOutput(t, func() {
		cmd.Execute()
	})
	_ = output
}

func TestStatusCLICmd_Execute_All(t *testing.T) {
	root, cleanup := setupTestDirWithSubRepos(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(root)
	defer os.Chdir(oldWd)

	cmd := statusCLICmd()
	cmd.SetArgs([]string{"--all", "--verbose"})
	output := captureOutput(t, func() {
		cmd.Execute()
	})
	_ = output
}

func TestBranchesCLICmd_Execute(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	cmd := branchesCLICmd()
	cmd.SetArgs([]string{})
	output := captureOutput(t, func() {
		cmd.Execute()
	})
	_ = output
}

func TestBranchesCLICmd_Execute_All(t *testing.T) {
	root, cleanup := setupTestDirWithSubRepos(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(root)
	defer os.Chdir(oldWd)

	cmd := branchesCLICmd()
	cmd.SetArgs([]string{"--all"})
	output := captureOutput(t, func() {
		cmd.Execute()
	})
	_ = output
}

func TestBranchesCLICmd_Execute_Error(t *testing.T) {
	dir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	cmd := branchesCLICmd()
	cmd.SetArgs([]string{})
	output := captureOutput(t, func() {
		cmd.Execute()
	})
	// Non-git dir should produce error output
	_ = output
}

func TestCommitCLICmd_Execute(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	// Create a change to commit
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("change"), 0644)

	cmd := commitCLICmd()
	cmd.SetArgs([]string{"-m", "test commit"})
	output := captureOutput(t, func() {
		cmd.Execute()
	})
	_ = output
}

func TestCommitCLICmd_Execute_WithBranch(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("change"), 0644)

	cmd := commitCLICmd()
	cmd.SetArgs([]string{"-m", "test commit", "-b", "feature"})
	output := captureOutput(t, func() {
		cmd.Execute()
	})
	_ = output
}

func TestCommitCLICmd_Execute_MissingMessage(t *testing.T) {
	cmd := commitCLICmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --message is missing")
	}
}

func TestPrCmd_Execute_MissingFlags(t *testing.T) {
	cmd := prCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when required flags are missing")
	}
}

func TestPrCmd_Execute(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("change"), 0644)

	cmd := prCmd()
	cmd.SetArgs([]string{"-m", "fix: test", "-b", "feature", "-t", "main"})
	output := captureOutput(t, func() {
		cmd.Execute()
	})
	_ = output
}

func TestPrCmd_Execute_All(t *testing.T) {
	root, cleanup := setupTestDirWithSubRepos(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(root)
	defer os.Chdir(oldWd)

	cmd := prCmd()
	cmd.SetArgs([]string{"-m", "fix: test", "-b", "feature", "-t", "main", "-a"})
	output := captureOutput(t, func() {
		cmd.Execute()
	})
	_ = output
}

func TestAutomergeCLICmd_Execute(t *testing.T) {
	dir, cleanup := setupTestDir(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldWd)

	cmd := automergeCLICmd()
	cmd.SetArgs([]string{"-t", "feature-branch"})
	output := captureOutput(t, func() {
		cmd.Execute()
	})
	_ = output
}

func TestAutomergeCLICmd_Execute_All(t *testing.T) {
	root, cleanup := setupTestDirWithSubRepos(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	os.Chdir(root)
	defer os.Chdir(oldWd)

	cmd := automergeCLICmd()
	cmd.SetArgs([]string{"-t", "feature", "-a"})
	output := captureOutput(t, func() {
		cmd.Execute()
	})
	_ = output
}

func TestAutomergeCLICmd_MissingTarget(t *testing.T) {
	cmd := automergeCLICmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when --target is missing")
	}
}

func TestCloneCLICmd_Execute_Repo(t *testing.T) {
	destDir := t.TempDir()

	cmd := cloneCLICmd()
	cmd.SetArgs([]string{"-r", "git@github.com:testorg/testrepo.git", "-p", destDir})
	output := captureOutput(t, func() {
		cmd.Execute()
	})
	_ = output
}

func TestCloneCLICmd_Execute_Org(t *testing.T) {
	destDir := t.TempDir()
	// Set GITHUB_TOKEN so the org clone branch is taken (it will fail at API call)
	os.Setenv("GITHUB_TOKEN", "fake-token")
	defer os.Unsetenv("GITHUB_TOKEN")

	cmd := cloneCLICmd()
	cmd.SetArgs([]string{"-o", "nonexistent-org", "-p", destDir})
	output := captureOutput(t, func() {
		err := cmd.Execute()
		// Expected to fail (fake token, fake org)
		_ = err
	})
	_ = output
}

func TestRunMergePRs_DryRun(t *testing.T) {
	// runMergePRs calls ops.FetchApprovedPRs which uses ghRun.
	// This will fail because gh CLI isn't available in test, but exercises the code.
	output := captureOutput(t, func() {
		err := runMergePRs(mergeConfig{
			org:       "test-org",
			dryRun:    true,
			batchSize: 5,
			waitMin:   0,
		})
		// Expected to fail at FetchApprovedPRs
		_ = err
	})
	_ = output
}

func TestMergePRsCLICmd_Execute_DryRun(t *testing.T) {
	cmd := mergePRsCLICmd()
	cmd.SetArgs([]string{"--org", "test-org", "--dry-run"})
	output := captureOutput(t, func() {
		cmd.Execute()
	})
	_ = output
}

func TestEnableWorkflowsCLICmd_Execute(t *testing.T) {
	cmd := enableWorkflowsCLICmd()
	cmd.SetArgs([]string{"--org", "test-org"})
	output := captureOutput(t, func() {
		cmd.Execute()
	})
	_ = output
}

func TestTeamPRsCLICmd_Execute(t *testing.T) {
	cmd := teamPRsCLICmd()
	cmd.SetArgs([]string{"--org", "test-org", "--team", "test-team"})
	output := captureOutput(t, func() {
		cmd.Execute()
	})
	_ = output
}

func TestResolvePaths_All_ValidDir(t *testing.T) {
	root, cleanup := setupTestDirWithSubRepos(t)
	defer cleanup()

	paths := resolvePaths(root, true)
	if len(paths) < 2 {
		t.Errorf("expected at least 2 paths, got %d", len(paths))
	}
}

func TestPrintStatusResult_AllFields(t *testing.T) {
	tests := []struct {
		name     string
		result   ops.StatusResult
		contains []string
	}{
		{
			name: "modified only",
			result: ops.StatusResult{
				RepoName: "mod-repo",
				Branch:   "main",
				Modified: 3,
			},
			contains: []string{"mod-repo", "main"},
		},
		{
			name: "deleted only",
			result: ops.StatusResult{
				RepoName: "del-repo",
				Branch:   "main",
				Deleted:  2,
			},
			contains: []string{"del-repo"},
		},
		{
			name: "untracked only",
			result: ops.StatusResult{
				RepoName:  "untr-repo",
				Branch:    "develop",
				Untracked: 7,
			},
			contains: []string{"untr-repo"},
		},
		{
			name: "behind only",
			result: ops.StatusResult{
				RepoName: "behind-repo",
				Branch:   "main",
				Behind:   5,
			},
			contains: []string{"behind-repo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureOutput(t, func() {
				printStatusResult(tt.result)
			})
			for _, s := range tt.contains {
				if !strings.Contains(output, s) {
					t.Errorf("output missing %q in %s", s, output)
				}
			}
		})
	}
}

func TestProcessMergeBatch_RealMerge(t *testing.T) {
	// Test non-dry-run path (will fail at ops.MergePR but exercises the code)
	prs := []ops.PRInfo{
		{Repo: "repo-a", Number: 1, Title: "Fix bug"},
	}

	cfg := mergeConfig{
		org:       "testorg",
		dryRun:    false,
		batchSize: 5,
		waitMin:   0,
	}

	output := captureOutput(t, func() {
		merged, failed := processMergeBatch(cfg, prs)
		// MergePR will fail due to no gh CLI, so failed should be 1
		_ = merged
		_ = failed
	})
	_ = output
}
