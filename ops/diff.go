package ops

// GetDiffStat returns the git diff --stat output for a repo.
func GetDiffStat(repoPath string) (string, error) {
	return gitRunInDir(repoPath, "diff", "--stat")
}
