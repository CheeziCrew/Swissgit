package ops

import "github.com/CheeziCrew/swissgit/git"

// GetRepoNameForPath returns the repo name for a given path.
func GetRepoNameForPath(repoPath string) (string, error) {
	return git.GetRepoName(repoPath)
}
