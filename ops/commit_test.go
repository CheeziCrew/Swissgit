package ops

import (
	"fmt"
	"testing"

	gogit "github.com/go-git/go-git/v5"
)

func TestCommitAndPush_OpenError(t *testing.T) {
	origOpen := plainOpen
	t.Cleanup(func() { plainOpen = origOpen })

	plainOpen = func(path string) (*gogit.Repository, error) {
		return nil, fmt.Errorf("not a git repo")
	}

	result := CommitAndPush("/tmp/nonexistent", "feature", "test commit")
	if result.Error == "" {
		t.Error("expected error for non-existent repo")
	}
	if result.Success {
		t.Error("expected Success to be false")
	}
}

func TestCommitResult_Fields(t *testing.T) {
	r := CommitResult{
		RepoName: "my-repo",
		Branch:   "main",
		Success:  true,
	}
	if r.RepoName != "my-repo" {
		t.Errorf("RepoName = %q, want %q", r.RepoName, "my-repo")
	}
	if r.Branch != "main" {
		t.Errorf("Branch = %q, want %q", r.Branch, "main")
	}
	if !r.Success {
		t.Error("expected Success to be true")
	}
}
