package ops

import (
	"fmt"
	"strings"
	"testing"
)

func TestEnableAutomerge(t *testing.T) {
	origGhRunInDir := ghRunInDir
	t.Cleanup(func() { ghRunInDir = origGhRunInDir })

	// EnableAutomerge calls git.GetRepoName which needs a real git repo.
	// We test the logic by checking the result struct fields.

	t.Run("success path", func(t *testing.T) {
		callCount := 0
		ghRunInDir = func(dir string, args ...string) (string, error) {
			callCount++
			if callCount == 1 {
				// pr list call
				if args[0] != "pr" || args[1] != "list" {
					t.Errorf("expected pr list, got %v", args)
				}
				return "42\n", nil
			}
			if callCount == 2 {
				// pr merge call
				if args[0] != "pr" || args[1] != "merge" || args[2] != "42" {
					t.Errorf("expected pr merge 42, got %v", args)
				}
				return "", nil
			}
			return "", fmt.Errorf("unexpected call %d", callCount)
		}

		result := EnableAutomerge("feature/test", "/fake/repo")
		if !result.Success {
			t.Errorf("expected success, got error: %s", result.Error)
		}
		if result.PRNumber != "42" {
			t.Errorf("PRNumber = %q, want %q", result.PRNumber, "42")
		}
	})

	t.Run("no PR found", func(t *testing.T) {
		ghRunInDir = func(dir string, args ...string) (string, error) {
			return "\n", nil
		}

		result := EnableAutomerge("feature/test", "/fake/repo")
		if result.Success {
			t.Error("expected failure when no PR found")
		}
		if !strings.Contains(result.Error, "no matching PR found") {
			t.Errorf("expected 'no matching PR found' error, got: %s", result.Error)
		}
	})

	t.Run("pr list error", func(t *testing.T) {
		ghRunInDir = func(dir string, args ...string) (string, error) {
			return "", fmt.Errorf("gh auth required")
		}

		result := EnableAutomerge("feature/test", "/fake/repo")
		if result.Success {
			t.Error("expected failure on pr list error")
		}
		if !strings.Contains(result.Error, "failed to list pull requests") {
			t.Errorf("expected 'failed to list pull requests' error, got: %s", result.Error)
		}
	})

	t.Run("pr merge error", func(t *testing.T) {
		callCount := 0
		ghRunInDir = func(dir string, args ...string) (string, error) {
			callCount++
			if callCount == 1 {
				return "99\n", nil
			}
			return "", fmt.Errorf("merge conflict")
		}

		result := EnableAutomerge("feature/test", "/fake/repo")
		if result.Success {
			t.Error("expected failure on merge error")
		}
		if !strings.Contains(result.Error, "failed to enable auto-merge") {
			t.Errorf("expected auto-merge error, got: %s", result.Error)
		}
		if result.PRNumber != "99" {
			t.Errorf("PRNumber = %q, want %q", result.PRNumber, "99")
		}
	})
}
