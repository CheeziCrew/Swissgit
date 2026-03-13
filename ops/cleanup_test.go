package ops

import (
	"fmt"
	"testing"

	gogit "github.com/go-git/go-git/v5"
)

func TestParseBranchName(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"simple branch", "  feature-x", "feature-x"},
		{"current branch marker", "* main", "main"},
		{"current branch with spaces", "  * develop  ", "develop"},
		{"empty string", "", ""},
		{"whitespace only", "   ", ""},
		{"branch with slashes", "  feature/my-feature", "feature/my-feature"},
		{"star only", "*", ""},
		{"star with spaces", "*   ", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseBranchName(tt.raw); got != tt.want {
				t.Errorf("parseBranchName(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestCollectMergedBranches_Parsing(t *testing.T) {
	// We can't call collectMergedBranches directly since it shells out,
	// but we can test the parsing logic it uses (parseBranchName + set logic).
	// Simulate what collectMergedBranches does with fixture output.
	fixture := "* main\n  feature-done\n  bugfix-merged\n  develop\n"

	protected := map[string]bool{"main": true, "develop": true}
	deleteSet := map[string]bool{}

	for _, raw := range splitLines(fixture) {
		branch := parseBranchName(raw)
		if branch != "" && !protected[branch] {
			deleteSet[branch] = true
		}
	}

	if len(deleteSet) != 2 {
		t.Fatalf("expected 2 branches to delete, got %d: %v", len(deleteSet), deleteSet)
	}
	if !deleteSet["feature-done"] {
		t.Error("expected feature-done in deleteSet")
	}
	if !deleteSet["bugfix-merged"] {
		t.Error("expected bugfix-merged in deleteSet")
	}
	if deleteSet["main"] {
		t.Error("main should be protected")
	}
	if deleteSet["develop"] {
		t.Error("develop should be protected")
	}
}

func TestCollectGoneBranches_Parsing(t *testing.T) {
	// Simulate the parsing logic of collectGoneBranches with fixture output.
	fixture := `* main                 abc1234 [origin/main] latest commit
  feature-old          def5678 [origin/feature-old: gone] old commit
  feature-active       ghi9012 [origin/feature-active] active commit
  stale-branch         jkl3456 [origin/stale-branch: gone] stale
`
	protected := map[string]bool{"main": true}
	deleteSet := map[string]bool{}

	for _, raw := range splitLines(fixture) {
		line := parseBranchName(raw)
		if line == "" {
			continue
		}
		parts := splitFields(line)
		if len(parts) < 1 || protected[parts[0]] {
			continue
		}
		if containsGone(raw) {
			deleteSet[parts[0]] = true
		}
	}

	if len(deleteSet) != 2 {
		t.Fatalf("expected 2 gone branches, got %d: %v", len(deleteSet), deleteSet)
	}
	if !deleteSet["feature-old"] {
		t.Error("expected feature-old in deleteSet")
	}
	if !deleteSet["stale-branch"] {
		t.Error("expected stale-branch in deleteSet")
	}
}

// helpers to avoid importing strings in test (mirrors what the real code does)
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func splitFields(s string) []string {
	var fields []string
	inField := false
	start := 0
	for i, c := range s {
		if c == ' ' || c == '\t' {
			if inField {
				fields = append(fields, s[start:i])
				inField = false
			}
		} else {
			if !inField {
				start = i
				inField = true
			}
		}
	}
	if inField {
		fields = append(fields, s[start:])
	}
	return fields
}

func containsGone(s string) bool {
	for i := 0; i+6 <= len(s); i++ {
		if s[i:i+6] == ": gone" {
			return true
		}
	}
	return false
}

func TestCleanupRepo_OpenError(t *testing.T) {
	origOpen := plainOpen
	t.Cleanup(func() { plainOpen = origOpen })

	plainOpen = func(path string) (*gogit.Repository, error) {
		return nil, fmt.Errorf("not a git repository")
	}

	result := CleanupRepo("/tmp/nonexistent", false, "")
	if result.Success {
		t.Error("expected failure when repo cannot be opened")
	}
	if result.Error == "" {
		t.Error("expected non-empty error message")
	}
	if result.RepoName != "nonexistent" {
		t.Errorf("RepoName = %q, want %q", result.RepoName, "nonexistent")
	}
}

func TestCleanupRepo_FetchError(t *testing.T) {
	origOpen := plainOpen
	origFetch := fetchRemote
	origGitRun := gitRunInDir
	t.Cleanup(func() {
		plainOpen = origOpen
		fetchRemote = origFetch
		gitRunInDir = origGitRun
	})

	// Create a real repo so GetRepoNameFromRepo and DefaultBranch can work
	dir := t.TempDir()
	repo, err := gogit.PlainInit(dir, false)
	if err != nil {
		t.Fatalf("failed to init repo: %v", err)
	}

	plainOpen = func(path string) (*gogit.Repository, error) {
		return repo, nil
	}

	// gitRunInDir succeeds for checkout but fetchRemote fails
	gitRunInDir = func(d string, args ...string) (string, error) {
		// checkout default branch succeeds
		return "", nil
	}
	fetchRemote = func(r *gogit.Repository) error {
		return fmt.Errorf("could not connect to remote")
	}

	result := CleanupRepo(dir, false, "main")
	if result.Success {
		t.Error("expected failure when fetch fails")
	}
	if result.Error == "" {
		t.Error("expected non-empty error message")
	}
}

func TestCleanupResult_Fields(t *testing.T) {
	r := CleanupResult{
		RepoName:        "my-repo",
		PrunedBranches:  3,
		RemainingBranch: 1,
		CurrentBranch:   "main",
		DefaultBranch:   "main",
		Success:         true,
	}
	if r.RepoName != "my-repo" {
		t.Errorf("RepoName = %q, want %q", r.RepoName, "my-repo")
	}
	if r.PrunedBranches != 3 {
		t.Errorf("PrunedBranches = %d, want 3", r.PrunedBranches)
	}
	if r.RemainingBranch != 1 {
		t.Errorf("RemainingBranch = %d, want 1", r.RemainingBranch)
	}
	if !r.Success {
		t.Error("expected Success to be true")
	}
}

func TestCountLocalBranches(t *testing.T) {
	origGitRun := gitRunInDir
	t.Cleanup(func() { gitRunInDir = origGitRun })

	t.Run("counts branches correctly", func(t *testing.T) {
		gitRunInDir = func(dir string, args ...string) (string, error) {
			return "* main\n  feature-a\n  feature-b\n", nil
		}
		count := countLocalBranches("/fake")
		if count != 3 {
			t.Errorf("countLocalBranches = %d, want 3", count)
		}
	})

	t.Run("returns 0 on error", func(t *testing.T) {
		gitRunInDir = func(dir string, args ...string) (string, error) {
			return "", fmt.Errorf("not a repo")
		}
		count := countLocalBranches("/fake")
		if count != 0 {
			t.Errorf("countLocalBranches = %d, want 0", count)
		}
	})

	t.Run("handles empty output", func(t *testing.T) {
		gitRunInDir = func(dir string, args ...string) (string, error) {
			return "", nil
		}
		count := countLocalBranches("/fake")
		if count != 0 {
			t.Errorf("countLocalBranches = %d, want 0", count)
		}
	})
}

func TestDeleteBranchSet(t *testing.T) {
	origGitRun := gitRunInDir
	t.Cleanup(func() { gitRunInDir = origGitRun })

	t.Run("deletes branches", func(t *testing.T) {
		gitRunInDir = func(dir string, args ...string) (string, error) {
			return "", nil
		}
		count := deleteBranchSet("/fake", map[string]bool{"feature-a": true, "feature-b": true})
		if count != 2 {
			t.Errorf("deleteBranchSet = %d, want 2", count)
		}
	})

	t.Run("returns -1 on error", func(t *testing.T) {
		gitRunInDir = func(dir string, args ...string) (string, error) {
			return "", fmt.Errorf("cannot delete branch")
		}
		count := deleteBranchSet("/fake", map[string]bool{"feature-a": true})
		if count != -1 {
			t.Errorf("deleteBranchSet = %d, want -1", count)
		}
	})

	t.Run("empty set returns 0", func(t *testing.T) {
		gitRunInDir = func(dir string, args ...string) (string, error) {
			t.Error("should not be called for empty set")
			return "", nil
		}
		count := deleteBranchSet("/fake", map[string]bool{})
		if count != 0 {
			t.Errorf("deleteBranchSet = %d, want 0", count)
		}
	})
}

func TestCollectMergedBranches_Mocked(t *testing.T) {
	origGitRun := gitRunInDir
	t.Cleanup(func() { gitRunInDir = origGitRun })

	t.Run("collects merged branches", func(t *testing.T) {
		gitRunInDir = func(dir string, args ...string) (string, error) {
			return "* main\n  feature-done\n  bugfix-old\n", nil
		}
		protected := map[string]bool{"main": true}
		deleteSet := map[string]bool{}
		err := collectMergedBranches("/fake", protected, deleteSet)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(deleteSet) != 2 {
			t.Errorf("deleteSet size = %d, want 2", len(deleteSet))
		}
	})

	t.Run("returns error on git failure", func(t *testing.T) {
		gitRunInDir = func(dir string, args ...string) (string, error) {
			return "", fmt.Errorf("git error")
		}
		err := collectMergedBranches("/fake", map[string]bool{}, map[string]bool{})
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestCollectGoneBranches_Mocked(t *testing.T) {
	origGitRun := gitRunInDir
	t.Cleanup(func() { gitRunInDir = origGitRun })

	t.Run("collects gone branches", func(t *testing.T) {
		gitRunInDir = func(dir string, args ...string) (string, error) {
			return "* main abc123 [origin/main] commit\n  old-branch def456 [origin/old-branch: gone] old\n", nil
		}
		protected := map[string]bool{"main": true}
		deleteSet := map[string]bool{}
		collectGoneBranches("/fake", protected, deleteSet)
		if !deleteSet["old-branch"] {
			t.Error("expected old-branch in deleteSet")
		}
	})

	t.Run("handles git error gracefully", func(t *testing.T) {
		gitRunInDir = func(dir string, args ...string) (string, error) {
			return "", fmt.Errorf("git error")
		}
		deleteSet := map[string]bool{}
		collectGoneBranches("/fake", map[string]bool{}, deleteSet)
		if len(deleteSet) != 0 {
			t.Errorf("deleteSet should be empty on error, got %v", deleteSet)
		}
	})
}

func TestCollectOrphanedBranches_Mocked(t *testing.T) {
	origGitRun := gitRunInDir
	t.Cleanup(func() { gitRunInDir = origGitRun })

	t.Run("collects orphaned branches", func(t *testing.T) {
		callCount := 0
		gitRunInDir = func(dir string, args ...string) (string, error) {
			callCount++
			if callCount == 1 {
				// branch --format call
				return "main\norphan-branch\n", nil
			}
			// rev-parse --verify for orphan-branch fails (no remote)
			return "", fmt.Errorf("not found")
		}
		protected := map[string]bool{"main": true}
		deleteSet := map[string]bool{}
		collectOrphanedBranches("/fake", protected, deleteSet)
		if !deleteSet["orphan-branch"] {
			t.Error("expected orphan-branch in deleteSet")
		}
	})

	t.Run("skips already in deleteSet", func(t *testing.T) {
		gitRunInDir = func(dir string, args ...string) (string, error) {
			return "main\nalready-marked\n", nil
		}
		protected := map[string]bool{"main": true}
		deleteSet := map[string]bool{"already-marked": true}
		collectOrphanedBranches("/fake", protected, deleteSet)
		// Should still be true, not removed
		if !deleteSet["already-marked"] {
			t.Error("already-marked should remain in deleteSet")
		}
	})
}

func TestCmdError(t *testing.T) {
	// Test with stderr
	e := &cmdError{stderr: "permission denied", err: fmt.Errorf("exit 1")}
	if e.Error() != "permission denied" {
		t.Errorf("Error() = %q, want %q", e.Error(), "permission denied")
	}

	// Test without stderr
	e2 := &cmdError{stderr: "", err: fmt.Errorf("exit status 128")}
	if e2.Error() != "exit status 128" {
		t.Errorf("Error() = %q, want %q", e2.Error(), "exit status 128")
	}
}
