package screens

import (
	"strings"
	"testing"

	"github.com/CheeziCrew/swissgit/git"
)

func TestRenderSkippedWarning_Empty(t *testing.T) {
	if got := renderSkippedWarning(nil); got != "" {
		t.Errorf("expected empty string for no skipped repos, got %q", got)
	}
	if got := renderSkippedWarning([]git.SkippedRepo{}); got != "" {
		t.Errorf("expected empty string for empty slice, got %q", got)
	}
}

func TestRenderSkippedWarning_NamesRepoAndReason(t *testing.T) {
	out := renderSkippedWarning([]git.SkippedRepo{
		{Path: "/Users/x/Code/scit/api-service-broken", Reason: "dangling extensions.worktreeConfig"},
	})
	if !strings.Contains(out, "api-service-broken") {
		t.Errorf("expected the repo name in the warning, got: %q", out)
	}
	if !strings.Contains(out, "dangling extensions.worktreeConfig") {
		t.Errorf("expected the reason in the warning, got: %q", out)
	}
	if !strings.Contains(out, "1 repo(s) skipped") {
		t.Errorf("expected a count in the warning, got: %q", out)
	}
}

func TestRenderSkippedWarning_CountsAll(t *testing.T) {
	out := renderSkippedWarning([]git.SkippedRepo{
		{Path: "/a/one", Reason: "boom"},
		{Path: "/a/two", Reason: "boom"},
	})
	if !strings.Contains(out, "2 repo(s) skipped") {
		t.Errorf("expected a count of 2, got: %q", out)
	}
	for _, name := range []string{"one", "two"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q listed, got: %q", name, out)
		}
	}
}
