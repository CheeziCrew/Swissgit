package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	gogit "github.com/go-git/go-git/v5"
)

// initTestRepo creates a real git repo in a temp dir with an initial commit and origin remote.
func initTestRepo(t *testing.T) (string, *gogit.Repository) {
	t.Helper()
	dir := t.TempDir()

	// Use shell git to create a proper repo
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %s\n%s", args, err, out)
		}
	}

	run("init", "-b", "main")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "test")

	// Create a file and commit
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test"), 0644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-m", "initial commit")

	// Add a remote
	run("remote", "add", "origin", "git@github.com:testowner/testrepo.git")

	repo, err := gogit.PlainOpen(dir)
	if err != nil {
		t.Fatalf("failed to open repo: %v", err)
	}

	return dir, repo
}

func TestGetBranchName(t *testing.T) {
	_, repo := initTestRepo(t)

	branch, err := GetBranchName(repo)
	if err != nil {
		t.Fatalf("GetBranchName error: %v", err)
	}
	if branch != "main" {
		t.Errorf("branch = %q, want %q", branch, "main")
	}
}

func TestGetRepoOwnerAndName(t *testing.T) {
	_, repo := initTestRepo(t)

	owner, name, err := GetRepoOwnerAndName(repo)
	if err != nil {
		t.Fatalf("GetRepoOwnerAndName error: %v", err)
	}
	if owner != "testowner" {
		t.Errorf("owner = %q, want %q", owner, "testowner")
	}
	if name != "testrepo" {
		t.Errorf("name = %q, want %q", name, "testrepo")
	}
}

func TestGetRepoOwnerAndName_NoRemotes(t *testing.T) {
	dir := t.TempDir()

	// Init repo without remote
	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %s\n%s", err, out)
	}

	repo, err := gogit.PlainOpen(dir)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = GetRepoOwnerAndName(repo)
	if err == nil {
		t.Error("expected error for repo with no remotes")
	}
}

func TestGetRepoName(t *testing.T) {
	dir, _ := initTestRepo(t)

	name, err := GetRepoName(dir)
	if err != nil {
		t.Fatalf("GetRepoName error: %v", err)
	}
	if name != "testrepo" {
		t.Errorf("name = %q, want %q", name, "testrepo")
	}
}

func TestGetRepoName_InvalidPath(t *testing.T) {
	_, err := GetRepoName("/nonexistent/repo/path")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestGetRepoNameFromRepo(t *testing.T) {
	_, repo := initTestRepo(t)

	name, err := GetRepoNameFromRepo(repo)
	if err != nil {
		t.Fatalf("GetRepoNameFromRepo error: %v", err)
	}
	if name != "testrepo" {
		t.Errorf("name = %q, want %q", name, "testrepo")
	}
}

func TestGogitSshAuthFunc(t *testing.T) {
	// Test that sshAuthFunc is a function variable (mockable)
	if sshAuthFunc == nil {
		t.Error("sshAuthFunc is nil")
	}
}

func TestGitRunVar(t *testing.T) {
	// Test that gitRun is a function variable (mockable)
	if gitRun == nil {
		t.Error("gitRun is nil")
	}
}
