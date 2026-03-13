package git

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
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

func TestDiscoverRepos(t *testing.T) {
	root := t.TempDir()

	// Create a repo dir (has .git)
	repoDir := filepath.Join(root, "my-repo")
	if err := os.MkdirAll(filepath.Join(repoDir, ".git"), 0o755); err != nil {
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

	repos, err := DiscoverRepos(root)
	if err != nil {
		t.Fatalf("DiscoverRepos() error: %v", err)
	}

	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d: %v", len(repos), repos)
	}
	if repos[0] != repoDir {
		t.Errorf("expected %s, got %s", repoDir, repos[0])
	}
}

func TestDiscoverRepos_EmptyDir(t *testing.T) {
	root := t.TempDir()
	repos, err := DiscoverRepos(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repos) != 0 {
		t.Errorf("expected 0 repos, got %d", len(repos))
	}
}

func TestDiscoverRepos_InvalidPath(t *testing.T) {
	_, err := DiscoverRepos("/nonexistent/path/xyz")
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
