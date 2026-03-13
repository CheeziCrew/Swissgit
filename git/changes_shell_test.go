package git

import (
	"fmt"
	"testing"
)

func TestCountChangesShell(t *testing.T) {
	orig := gitRun
	t.Cleanup(func() { gitRun = orig })

	t.Run("mixed changes", func(t *testing.T) {
		gitRun = func(repoPath string, args ...string) (string, error) {
			// Simulated porcelain output
			return " M modified.go\n?? untracked.txt\nA  added.go\n D deleted.go\n", nil
		}

		got, err := CountChangesShell("/fake/repo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Modified != 1 {
			t.Errorf("Modified = %d, want 1", got.Modified)
		}
		if got.Untracked != 1 {
			t.Errorf("Untracked = %d, want 1", got.Untracked)
		}
		if got.Added != 1 {
			t.Errorf("Added = %d, want 1", got.Added)
		}
		if got.Deleted != 1 {
			t.Errorf("Deleted = %d, want 1", got.Deleted)
		}
	})

	t.Run("clean repo", func(t *testing.T) {
		gitRun = func(repoPath string, args ...string) (string, error) {
			return "", nil
		}

		got, err := CountChangesShell("/fake/repo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.HasChanges() {
			t.Errorf("expected no changes, got %+v", got)
		}
	})

	t.Run("multiple modified files", func(t *testing.T) {
		gitRun = func(repoPath string, args ...string) (string, error) {
			return " M a.go\n M b.go\nMM c.go\n", nil
		}

		got, err := CountChangesShell("/fake/repo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Modified != 3 {
			t.Errorf("Modified = %d, want 3", got.Modified)
		}
	})

	t.Run("error from git", func(t *testing.T) {
		gitRun = func(repoPath string, args ...string) (string, error) {
			return "", fmt.Errorf("not a git repo")
		}

		_, err := CountChangesShell("/fake/repo")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("renamed and copied files count as added", func(t *testing.T) {
		gitRun = func(repoPath string, args ...string) (string, error) {
			return "R  old.go -> new.go\nC  orig.go -> copy.go\n", nil
		}

		got, err := CountChangesShell("/fake/repo")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Added != 2 {
			t.Errorf("Added = %d, want 2", got.Added)
		}
	})
}
