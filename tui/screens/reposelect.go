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
		// ScannerWithWarnings, not Scanner: a repo that cannot be opened must
		// be named, not silently dropped from the list.
		ScannerWithWarnings: gitScan,
	})
}

// gitScan discovers git repos and collects their status, reporting any it
// could not open rather than dropping them. Wording matches
// renderSkippedWarning so the same repo reads the same everywhere.
func gitScan(rootPath string) ([]curd.RepoInfo, []string, error) {
	paths, skipped, err := git.DiscoverReposWithSkipped(rootPath)
	if err != nil {
		return nil, nil, err
	}

	warnings := make([]string, 0, len(skipped))
	for _, s := range skipped {
		warnings = append(warnings, filepath.Base(s.Path)+" — "+s.Reason)
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

	return repos, warnings, nil
}

func getBranchShell(repoPath string) string {
	cmd := exec.Command("git", "-C", repoPath, "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
