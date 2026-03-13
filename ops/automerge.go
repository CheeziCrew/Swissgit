package ops

import (
	"fmt"
	"strings"

	"github.com/CheeziCrew/swissgit/git"
)

// AutomergeResult holds the outcome of enabling automerge.
type AutomergeResult struct {
	RepoName string
	PRNumber string
	Success  bool
	Error    string
}

// EnableAutomerge finds a PR matching the target and enables auto-merge via gh CLI.
func EnableAutomerge(target, repoPath string) AutomergeResult {
	repoName, _ := git.GetRepoName(repoPath)
	result := AutomergeResult{RepoName: repoName}

	out, err := ghRunInDir(repoPath, "pr", "list", "--head", target, "--json", "number", "--jq", ".[0].number")
	if err != nil {
		result.Error = fmt.Sprintf("failed to list pull requests: %s", err)
		return result
	}

	prNumber := strings.TrimSpace(out)
	if prNumber == "" {
		result.Error = fmt.Sprintf("no matching PR found for target: %s", target)
		return result
	}
	result.PRNumber = prNumber

	_, err = ghRunInDir(repoPath, "pr", "merge", prNumber, "--auto", "--merge", "--delete-branch=true")
	if err != nil {
		result.Error = fmt.Sprintf("failed to enable auto-merge for PR #%s: %s", prNumber, err)
		return result
	}

	result.Success = true
	return result
}
