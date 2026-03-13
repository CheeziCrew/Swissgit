package ops

import (
	"testing"
)

func TestStatusResult_Fields(t *testing.T) {
	r := StatusResult{
		RepoName:      "my-repo",
		Branch:        "feature/x",
		DefaultBranch: "main",
		Modified:      2,
		Added:         1,
		Deleted:       3,
		Untracked:     4,
		Ahead:         1,
		Behind:        2,
		Clean:         false,
	}
	if r.RepoName != "my-repo" {
		t.Errorf("RepoName = %q, want %q", r.RepoName, "my-repo")
	}
	if r.Branch != "feature/x" {
		t.Errorf("Branch = %q, want %q", r.Branch, "feature/x")
	}
	if r.Modified != 2 {
		t.Errorf("Modified = %d, want 2", r.Modified)
	}
	if r.Clean {
		t.Error("expected Clean = false")
	}
}
