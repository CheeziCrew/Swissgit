package ops

import (
	"fmt"
	"testing"

	gogit "github.com/go-git/go-git/v5"
)

func TestGetRepoStatus_OpenError(t *testing.T) {
	origOpen := plainOpen
	t.Cleanup(func() { plainOpen = origOpen })

	plainOpen = func(path string) (*gogit.Repository, error) {
		return nil, fmt.Errorf("not a git repo")
	}

	result := GetRepoStatus("/tmp/nonexistent")
	if result.Error == "" {
		t.Error("expected error for non-existent repo")
	}
	if result.RepoName != "nonexistent" {
		t.Errorf("RepoName = %q, want %q", result.RepoName, "nonexistent")
	}
}

func TestGetRepoStatus_FetchError(t *testing.T) {
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

	result := GetRepoStatus("/tmp/some-repo")
	if result.Error == "" {
		t.Error("expected error when fetch fails")
	}
	if result.RepoName != "some-repo" {
		t.Errorf("RepoName = %q, want %q", result.RepoName, "some-repo")
	}
}
