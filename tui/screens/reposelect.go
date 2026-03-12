package screens

import (
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/CheeziCrew/curd"
	"github.com/CheeziCrew/swissgit/git"
)

// Re-export for backward compat.
type RepoSelectModel = curd.RepoSelectModel

// NewRepoSelectModel creates a repo selector using swissgit's git scanner.
func NewRepoSelectModel(caller, rootPath string, parentOffset, termHeight int) curd.RepoSelectModel {
	return curd.NewRepoSelectModel(curd.RepoSelectConfig{
		Palette:      curd.SwissgitPalette,
		RootPath:     rootPath,
		Caller:       caller,
		ParentOffset: parentOffset,
		TermHeight:   termHeight,
		Scanner:      gitScan,
	})
}

// gitScan discovers git repos and collects their status.
func gitScan(rootPath string) ([]curd.RepoInfo, error) {
	paths, err := git.DiscoverRepos(rootPath)
	if err != nil {
		return nil, err
	}

	var repos []curd.RepoInfo
	for _, p := range paths {
		name := filepath.Base(p)
		repoName, err := git.GetRepoName(p)
		if err == nil {
			name = repoName
		}

		changes, _ := git.CountChangesShell(p)
		defaultBranch := git.DefaultBranch(p, "main")
		branch := getBranchShell(p)

		repos = append(repos, curd.RepoInfo{
			Path:          p,
			Name:          name,
			Branch:        branch,
			DefaultBranch: defaultBranch,
			Modified:      changes.Modified,
			Added:         changes.Added,
			Deleted:       changes.Deleted,
			Untracked:     changes.Untracked,
			IsDirty:       changes.HasChanges() || (branch != "" && branch != defaultBranch),
		})
	}

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Name < repos[j].Name
	})

	return repos, nil
}

func getBranchShell(repoPath string) string {
	cmd := exec.Command("git", "-C", repoPath, "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
