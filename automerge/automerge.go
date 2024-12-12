package automerge

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/CheeziCrew/swissgit/utils"
	"github.com/fatih/color"
)

func HandleAutoMerge(targetBranch, repoPath string, allRepoFlag bool) error {
	if allRepoFlag {
		return ProcessSubdirectories(targetBranch, repoPath)

	} else {
		return ProcessSingleRepository(targetBranch, repoPath)
	}
}

func ProcessSubdirectories(targetBranch, repoPath string) error {
	entries, err := os.ReadDir(repoPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subRepoPath := filepath.Join(repoPath, entry.Name())

			if !utils.IsGitRepository(subRepoPath) {
				continue
			}

			if err := ProcessSingleRepository(targetBranch, subRepoPath); err != nil {
				fmt.Printf("Error processing subdirectory %s: %v\n", subRepoPath, err)
			}
		}
	}
	return nil
}

func ProcessSingleRepository(targetBranch, repoPath string) error {
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	repoName, err := utils.GetRepoName(repoPath)
	if err != nil {
		statusMessage := fmt.Sprintf("%s: getting repoName", repoPath)
		fmt.Printf("\r%s failed [%s]: %s\n", statusMessage, red("x"), red(err.Error()))
		return err
	}

	statusMessage := fmt.Sprintf("%s: enabling auto merge", repoName)
	done := make(chan bool)
	go utils.ShowSpinner(statusMessage, done)

	prNumber, err := Automerge(targetBranch, repoPath)
	done <- true

	if err != nil {
		fmt.Printf("\r%s failed [%s]: %s \n", statusMessage, red("x"), red(err.Error()))
		return nil
	}

	fmt.Printf("\r%s [%s]\n", fmt.Sprintf("%s: enabled auto merge for PR #%s", repoName, prNumber), green("✔"))
	return nil
}

func Automerge(target, repoPath string) (string, error) {
	// Set the environment variable for the repository path
	os.Chdir(repoPath)

	// Step 1: Get the PR number by searching with the target
	cmdList := exec.Command("gh", "pr", "list", "--limit", "100", "--search", target, "--json", "number", "--jq", ".[0].number")
	var out bytes.Buffer
	cmdList.Stdout = &out
	if err := cmdList.Run(); err != nil {
		return "", fmt.Errorf("failed to list pull requests: %w", err)
	}

	prNumber := strings.TrimSpace(out.String())
	if prNumber == "" {
		return "", fmt.Errorf("no matching PR found in %s for target: %s", repoPath, target)
	}
	cmdMerge := exec.Command("gh", "pr", "merge", prNumber, "--auto", "--merge", "--delete-branch=true")
	if err := cmdMerge.Run(); err != nil {
		return "", fmt.Errorf("failed to enable auto-merge for PR #%s: %w", prNumber, err)
	}

	return prNumber, nil
}
