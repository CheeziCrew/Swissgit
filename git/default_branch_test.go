package git

import (
	"fmt"
	"strings"
	"testing"
)

func TestDefaultBranch(t *testing.T) {
	orig := gitRun
	t.Cleanup(func() { gitRun = orig })

	t.Run("origin/HEAD succeeds", func(t *testing.T) {
		gitRun = func(repoPath string, args ...string) (string, error) {
			if args[0] == "symbolic-ref" {
				return "refs/remotes/origin/main\n", nil
			}
			return "", fmt.Errorf("unexpected call")
		}

		got := DefaultBranch("/fake/repo", "develop")
		if got != "main" {
			t.Errorf("DefaultBranch() = %q, want %q", got, "main")
		}
	})

	t.Run("origin/HEAD fails, show-ref finds fallback branch", func(t *testing.T) {
		gitRun = func(repoPath string, args ...string) (string, error) {
			if args[0] == "symbolic-ref" {
				return "", fmt.Errorf("not set")
			}
			if args[0] == "show-ref" {
				// The fallback param is "develop", so it checks "develop" first
				if strings.Contains(strings.Join(args, " "), "refs/heads/develop") {
					return "", nil // found
				}
				return "", fmt.Errorf("not found")
			}
			return "", fmt.Errorf("unexpected call: %v", args)
		}

		got := DefaultBranch("/fake/repo", "develop")
		if got != "develop" {
			t.Errorf("DefaultBranch() = %q, want %q", got, "develop")
		}
	})

	t.Run("origin/HEAD fails, show-ref finds main", func(t *testing.T) {
		gitRun = func(repoPath string, args ...string) (string, error) {
			if args[0] == "symbolic-ref" {
				return "", fmt.Errorf("not set")
			}
			if args[0] == "show-ref" {
				if strings.Contains(strings.Join(args, " "), "refs/heads/main") {
					return "", nil
				}
				return "", fmt.Errorf("not found")
			}
			return "", fmt.Errorf("unexpected call: %v", args)
		}

		// fallback is something that won't match show-ref
		got := DefaultBranch("/fake/repo", "nope")
		if got != "main" {
			t.Errorf("DefaultBranch() = %q, want %q", got, "main")
		}
	})

	t.Run("all fail except branch --show-current", func(t *testing.T) {
		gitRun = func(repoPath string, args ...string) (string, error) {
			if args[0] == "symbolic-ref" {
				return "", fmt.Errorf("not set")
			}
			if args[0] == "show-ref" {
				return "", fmt.Errorf("not found")
			}
			if args[0] == "branch" && args[1] == "--show-current" {
				return "feature-x\n", nil
			}
			return "", fmt.Errorf("unexpected call: %v", args)
		}

		got := DefaultBranch("/fake/repo", "develop")
		if got != "feature-x" {
			t.Errorf("DefaultBranch() = %q, want %q", got, "feature-x")
		}
	})

	t.Run("everything fails returns fallback", func(t *testing.T) {
		gitRun = func(repoPath string, args ...string) (string, error) {
			return "", fmt.Errorf("fail")
		}

		got := DefaultBranch("/fake/repo", "develop")
		if got != "develop" {
			t.Errorf("DefaultBranch() = %q, want %q", got, "develop")
		}
	})
}

func TestAheadBehind(t *testing.T) {
	orig := gitRun
	t.Cleanup(func() { gitRun = orig })

	t.Run("normal output", func(t *testing.T) {
		gitRun = func(repoPath string, args ...string) (string, error) {
			return "3\t5\n", nil
		}

		ahead, behind := AheadBehind("/fake/repo", "main")
		if ahead != 3 {
			t.Errorf("ahead = %d, want 3", ahead)
		}
		if behind != 5 {
			t.Errorf("behind = %d, want 5", behind)
		}
	})

	t.Run("zero counts", func(t *testing.T) {
		gitRun = func(repoPath string, args ...string) (string, error) {
			return "0\t0\n", nil
		}

		ahead, behind := AheadBehind("/fake/repo", "main")
		if ahead != 0 {
			t.Errorf("ahead = %d, want 0", ahead)
		}
		if behind != 0 {
			t.Errorf("behind = %d, want 0", behind)
		}
	})

	t.Run("error returns zeros", func(t *testing.T) {
		gitRun = func(repoPath string, args ...string) (string, error) {
			return "", fmt.Errorf("no upstream")
		}

		ahead, behind := AheadBehind("/fake/repo", "main")
		if ahead != 0 || behind != 0 {
			t.Errorf("expected (0, 0) on error, got (%d, %d)", ahead, behind)
		}
	})

	t.Run("malformed output returns zeros", func(t *testing.T) {
		gitRun = func(repoPath string, args ...string) (string, error) {
			return "garbage", nil
		}

		ahead, behind := AheadBehind("/fake/repo", "main")
		if ahead != 0 || behind != 0 {
			t.Errorf("expected (0, 0) on bad output, got (%d, %d)", ahead, behind)
		}
	})
}
