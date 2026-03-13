package ops

import (
	"fmt"
	"testing"

	gogit "github.com/go-git/go-git/v5"
)

func TestGetBranches_OpenError(t *testing.T) {
	origOpen := plainOpen
	t.Cleanup(func() { plainOpen = origOpen })

	plainOpen = func(path string) (*gogit.Repository, error) {
		return nil, fmt.Errorf("not a git repo")
	}

	result := GetBranches("/tmp/nonexistent")
	if result.Error == "" {
		t.Error("expected error for non-existent repo")
	}
	if result.RepoName != "nonexistent" {
		t.Errorf("RepoName = %q, want %q", result.RepoName, "nonexistent")
	}
}

func TestGetBranches_FetchError(t *testing.T) {
	origOpen := plainOpen
	origFetch := fetchRemote
	t.Cleanup(func() {
		plainOpen = origOpen
		fetchRemote = origFetch
	})

	plainOpen = func(path string) (*gogit.Repository, error) {
		return &gogit.Repository{}, nil
	}
	fetchRemote = func(repo *gogit.Repository) error {
		return fmt.Errorf("fetch failed: no network")
	}

	result := GetBranches("/tmp/test-repo")
	if result.Error == "" {
		t.Error("expected error when fetch fails")
	}
	if result.RepoName != "test-repo" {
		t.Errorf("RepoName = %q, want %q", result.RepoName, "test-repo")
	}
}

func TestBranchInfo_Fields(t *testing.T) {
	b := BranchInfo{Name: "feature/x", IsStale: true, IsRemote: false}
	if b.Name != "feature/x" {
		t.Errorf("Name = %q, want %q", b.Name, "feature/x")
	}
	if !b.IsStale {
		t.Error("expected IsStale to be true")
	}
	if b.IsRemote {
		t.Error("expected IsRemote to be false")
	}
}
