package ops

import (
	"strings"
	"testing"
)

func TestBuildPullRequestBody(t *testing.T) {
	t.Run("no changes selected, no breaking change", func(t *testing.T) {
		body, err := BuildPullRequestBody(nil, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// All checkboxes should be unchecked for change types
		for _, ct := range ChangeTypes {
			checked := "- [x] " + ct
			if strings.Contains(body, checked) {
				t.Errorf("expected %q to be unchecked, but found checked", ct)
			}
		}
		// "No" should be checked for breaking change
		if !strings.Contains(body, "- [x] No") {
			t.Error("expected 'No' breaking change to be checked")
		}
		if strings.Contains(body, "- [x] Yes") {
			t.Error("expected 'Yes' breaking change to NOT be checked")
		}
	})

	t.Run("single change selected", func(t *testing.T) {
		body, err := BuildPullRequestBody([]string{"Bug fix"}, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(body, "- [x] Bug fix") {
			t.Error("expected 'Bug fix' to be checked")
		}
		if strings.Contains(body, "- [x] New feature") {
			t.Error("expected 'New feature' to be unchecked")
		}
	})

	t.Run("multiple changes selected", func(t *testing.T) {
		changes := []string{"Bug fix", "New feature", "Documentation content changes"}
		body, err := BuildPullRequestBody(changes, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, c := range changes {
			if !strings.Contains(body, "- [x] "+c) {
				t.Errorf("expected %q to be checked", c)
			}
		}
	})

	t.Run("breaking change", func(t *testing.T) {
		body, err := BuildPullRequestBody(nil, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(body, "- [x] Yes (I have stepped the version number accordingly)") {
			t.Error("expected 'Yes' breaking change to be checked")
		}
		// "No" should still be unchecked
		if strings.Contains(body, "- [x] No") {
			t.Error("expected 'No' to remain unchecked when breaking change is true")
		}
	})

	t.Run("body contains checklist section", func(t *testing.T) {
		body, err := BuildPullRequestBody(nil, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(body, "## Checklist:") {
			t.Error("expected body to contain Checklist section")
		}
		if !strings.Contains(body, "## Types of changes") {
			t.Error("expected body to contain Types of changes section")
		}
	})
}
