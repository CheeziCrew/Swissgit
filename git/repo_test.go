package git

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	gogit "github.com/go-git/go-git/v5"
)

func TestIsGitRepository(t *testing.T) {
	t.Run("with .git dir", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
			t.Fatal(err)
		}
		if !IsGitRepository(dir) {
			t.Error("expected true for directory with .git")
		}
	})

	t.Run("without .git dir", func(t *testing.T) {
		dir := t.TempDir()
		if IsGitRepository(dir) {
			t.Error("expected false for directory without .git")
		}
	})

	t.Run("nonexistent path", func(t *testing.T) {
		if IsGitRepository("/nonexistent/path/that/does/not/exist") {
			t.Error("expected false for nonexistent path")
		}
	})
}

func TestDiscoverReposWithSkipped(t *testing.T) {
	root := t.TempDir()

	// Create a real repo dir — an empty .git directory is not openable, and
	// discovery only returns repos it can actually open.
	repoDir := filepath.Join(root, "my-repo")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := gogit.PlainInit(repoDir, false); err != nil {
		t.Fatal(err)
	}

	// Create a non-repo dir (no .git)
	plainDir := filepath.Join(root, "plain-dir")
	if err := os.MkdirAll(plainDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a file (not a directory)
	if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	repos, skipped, err := DiscoverReposWithSkipped(root)
	if err != nil {
		t.Fatalf("DiscoverReposWithSkipped() error: %v", err)
	}
	if len(skipped) != 0 {
		t.Errorf("nothing here should be skipped, got %v", skipped)
	}

	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d: %v", len(repos), repos)
	}
	if repos[0] != repoDir {
		t.Errorf("expected %s, got %s", repoDir, repos[0])
	}
}

func TestDiscoverReposWithSkipped_UnopenableRepoIsSkipped(t *testing.T) {
	root := t.TempDir()

	good := filepath.Join(root, "good-repo")
	if err := os.MkdirAll(good, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := gogit.PlainInit(good, false); err != nil {
		t.Fatal(err)
	}

	// Looks like a repo (has .git) but cannot be opened.
	broken := filepath.Join(root, "broken-repo")
	if err := os.MkdirAll(filepath.Join(broken, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	repos, skipped, err := DiscoverReposWithSkipped(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 1 || repos[0] != good {
		t.Fatalf("expected only %s, got %v", good, repos)
	}
	if len(skipped) != 1 || skipped[0].Path != broken {
		t.Fatalf("expected %s to be skipped, got %v", broken, skipped)
	}
	if skipped[0].Reason == "" {
		t.Error("expected a non-empty reason for the skipped repo")
	}
}

func TestDiscoverReposWithSkipped_DanglingWorktreeConfigExplained(t *testing.T) {
	root := t.TempDir()

	poisoned := filepath.Join(root, "poisoned-repo")
	if err := os.MkdirAll(poisoned, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := gogit.PlainInit(poisoned, false); err != nil {
		t.Fatal(err)
	}
	// Reproduce the real-world damage: the extension is set, but the
	// repository format version is left at 0.
	cfg := filepath.Join(poisoned, ".git", "config")
	data, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfg, append(data, []byte("[extensions]\n\tworktreeConfig = true\n")...), 0o644); err != nil {
		t.Fatal(err)
	}

	repos, skipped, err := DiscoverReposWithSkipped(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 0 {
		t.Fatalf("expected the poisoned repo to be skipped, got %v", repos)
	}
	if len(skipped) != 1 {
		t.Fatalf("expected 1 skipped repo, got %d", len(skipped))
	}
	if !strings.Contains(skipped[0].Reason, "extensions.worktreeConfig") {
		t.Errorf("reason should name the dangling extension, got: %s", skipped[0].Reason)
	}
	if !strings.Contains(skipped[0].Reason, "--unset") {
		t.Errorf("reason should include the repair command, got: %s", skipped[0].Reason)
	}
}

func TestDiscoverRepos_SkipsRatherThanFailing(t *testing.T) {
	root := t.TempDir()
	broken := filepath.Join(root, "broken-repo")
	if err := os.MkdirAll(filepath.Join(broken, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	repos, _, err := DiscoverReposWithSkipped(root)
	if err != nil {
		t.Fatalf("one unopenable repo must not fail the whole scan: %v", err)
	}
	if len(repos) != 0 {
		t.Errorf("expected 0 usable repos, got %v", repos)
	}
}

func TestDiscoverRepos_EmptyDir(t *testing.T) {
	root := t.TempDir()
	repos, _, err := DiscoverReposWithSkipped(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 0 {
		t.Errorf("expected 0 repos, got %d", len(repos))
	}
}

func TestDiscoverRepos_InvalidPath(t *testing.T) {
	_, _, err := DiscoverReposWithSkipped("/nonexistent/path/xyz")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestGetRepoOwnerAndName_Regex(t *testing.T) {
	// Test the regex used in GetRepoOwnerAndName directly,
	// since calling the full function requires a go-git Repository.
	re := regexp.MustCompile(`(?:[:/])([^/]+)/([^/]+?)(?:\.git)?$`)

	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
		wantMatch bool
	}{
		{
			"SSH URL",
			"git@github.com:CheeziCrew/swissgit.git",
			"CheeziCrew", "swissgit", true,
		},
		{
			"SSH URL without .git",
			"git@github.com:owner/repo",
			"owner", "repo", true,
		},
		{
			"HTTPS URL",
			"https://github.com/CheeziCrew/swissgit.git",
			"CheeziCrew", "swissgit", true,
		},
		{
			"HTTPS URL without .git",
			"https://github.com/myorg/myrepo",
			"myorg", "myrepo", true,
		},
		{
			"HTTPS with trailing slash stripped",
			"https://github.com/org/repo.git",
			"org", "repo", true,
		},
		{
			"no match",
			"not-a-url",
			"", "", false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := re.FindStringSubmatch(tt.url)
			if !tt.wantMatch {
				if len(matches) >= 3 {
					t.Errorf("expected no match for %q, got %v", tt.url, matches)
				}
				return
			}
			if len(matches) < 3 {
				t.Fatalf("expected match for %q, got none", tt.url)
			}
			if matches[1] != tt.wantOwner {
				t.Errorf("owner: got %q, want %q", matches[1], tt.wantOwner)
			}
			if matches[2] != tt.wantRepo {
				t.Errorf("repo: got %q, want %q", matches[2], tt.wantRepo)
			}
		})
	}
}
