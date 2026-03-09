package ops

import (
	"bytes"
	"fmt"
	"os/exec"
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

	cmdList := exec.Command("gh", "pr", "list", "--head", target, "--json", "number", "--jq", ".[0].number")
	cmdList.Dir = repoPath
	var out bytes.Buffer
	cmdList.Stdout = &out
	if err := cmdList.Run(); err != nil {
		result.Error = fmt.Sprintf("failed to list pull requests: %s", err)
		return result
	}

	prNumber := strings.TrimSpace(out.String())
	if prNumber == "" {
		result.Error = fmt.Sprintf("no matching PR found for target: %s", target)
		return result
	}
	result.PRNumber = prNumber

	cmdMerge := exec.Command("gh", "pr", "merge", prNumber, "--auto", "--merge", "--delete-branch=true")
	cmdMerge.Dir = repoPath
	if err := cmdMerge.Run(); err != nil {
		result.Error = fmt.Sprintf("failed to enable auto-merge for PR #%s: %s", prNumber, err)
		return result
	}

	result.Success = true
	return result
}
