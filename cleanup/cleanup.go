package cleanup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/CheeziCrew/swissgit/utils"
	"github.com/CheeziCrew/swissgit/utils/gitCommands"
	"github.com/fatih/color"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// CleanupOptions holds the configuration for the cleanup process
type CleanupOptions struct {
	RepoPath    string
	AllFlag     bool
	DropChanges bool
}

// Cleanup performs the cleanup process on repositories in the specified directory
func Cleanup(opts CleanupOptions) {
	if opts.AllFlag {
		err := ProcessSubdirectories(opts)
		if err != nil {
			fmt.Printf("%s\n", err)
		}
	} else {
		err := ProcessSingleRepository(opts)
		if err != nil {
			fmt.Printf("%s: %s\n", opts.RepoPath, err)
		}
	}
}

func ProcessSubdirectories(opts CleanupOptions) error {
	entries, err := os.ReadDir(opts.RepoPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subRepoPath := filepath.Join(opts.RepoPath, entry.Name())

			// Check if the subdirectory is a Git repository
			if !utils.IsGitRepository(subRepoPath) {
				continue
			}

			err := ProcessSingleRepository(CleanupOptions{RepoPath: subRepoPath, AllFlag: false, DropChanges: opts.DropChanges})
			if err != nil {
				// Log the error but continue processing other repositories
				return fmt.Errorf(" %s: %w", subRepoPath, err)
			}
		}
	}
	return nil
}

// ProcessSingleRepository processes a single Git repository
func ProcessSingleRepository(opts CleanupOptions) error {
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	repoName, err := utils.GetRepoName(opts.RepoPath)
	if err != nil {
		statusMessage := fmt.Sprintf("%s: getting repoName", opts.RepoPath)
		fmt.Printf("\r%s failed [%s]: %s\n", statusMessage, red("x"), red(err))
		return nil
	}

	statusMessage := fmt.Sprintf("%s: Cleaning up repo", repoName)
	done := make(chan bool)
	go utils.ShowSpinner(statusMessage, done)

	changes, err := checkForChanges(opts.RepoPath, opts.DropChanges)
	if err != nil {
		done <- true
		fmt.Printf("\r%s failed [%s]: %s\n", statusMessage, red("x"), red(err))
		return nil
	}

	repo, err := git.PlainOpen(opts.RepoPath)
	if err != nil {
		done <- true
		fmt.Printf("\r%s failed [%s]: %s\n", statusMessage, red("x"), red(err))
		return nil
	}

	prunedBranches, branchesCount, err := updateBranches(repo)
	if err != nil {
		done <- true
		fmt.Printf("\r%s failed [%s]: %s\n", statusMessage, red("x"), red(err))
		return nil
	}

	currentBranch, err := utils.GetBranchName(repo)
	if err != nil {
		done <- true
		fmt.Printf("\r%s failed [%s]: %s\n", statusMessage, red("x"), red(err))
		return nil
	}

	statusLine := constructStatusLine(changes, currentBranch, prunedBranches, branchesCount)
	done <- true
	statusMessage = fmt.Sprintf("%s: Cleaned up repo", repoName)
	fmt.Printf("\r%s [%s] %s\n", statusMessage, green("✔"), statusLine)

	return nil
}

// checkForChanges handles resetting changes or listing uncommitted changes
func checkForChanges(repoPath string, dropChanges bool) (string, error) {
	var changes string

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to open repository: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := wt.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get status for repository: %w", err)
	}

	if !status.IsClean() {
		if dropChanges {
			_ = wt.Reset(&git.ResetOptions{Mode: git.HardReset})
			changes = color.RedString("Dropped all changes")
		} else {
			modified, added, deleted, untracked := CountChanges(status)
			changes = fmt.Sprintf("[%s] [%s] [%s] [%s]", color.YellowString("Modified: %d", modified), color.GreenString("Added: %d", added), color.RedString("Deleted: %d", deleted), color.BlueString("Untracked: %d", untracked))
		}
	}
	return changes, nil
}

func CountChanges(status git.Status) (int, int, int, int) {
	var modified, added, deleted, untracked int
	for _, state := range status {
		if state.Worktree == git.Untracked {
			untracked++
		}
		if state.Staging == git.Modified || state.Worktree == git.Modified {
			modified++
		}
		if state.Staging == git.Added || state.Worktree == git.Added {
			added++
		}
		if state.Staging == git.Deleted || state.Worktree == git.Deleted {
			deleted++
		}
	}
	return modified, added, deleted, untracked
}

func updateBranches(repo *git.Repository) (int, int, error) {
	wt, err := repo.Worktree()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get worktree: %w", err)
	}

	// Switch to main for cleanup
	err = wt.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("main")})
	if err != nil {
		return 0, 0, fmt.Errorf("failed to checkout main branch: %w", err)
	}

	gitCommands.FetchRemote(repo)
	gitCommands.PullChanges(wt)

	parentDir, err := os.Getwd()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get current directory: %w", err)
	}

	err = os.Chdir(wt.Filesystem.Root())
	if err != nil {
		return 0, 0, fmt.Errorf("failed to change directory to repo: %w", err)
	}

	cmd := exec.Command("git", "branch", "--merged")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get merged branches: %w, output: %s", err, output)
	}

	branches := strings.Split(string(output), "\n")
	protectedBranches := map[string]bool{
		"main": true,
	}

	var branchesToDelete []string
	for _, branch := range branches {
		branch = strings.TrimSpace(branch)
		if branch == "" {
			continue
		}
		// Remove leading '*' if present
		if strings.HasPrefix(branch, "*") {
			branch = strings.TrimSpace(branch[1:])
		}
		if protectedBranches[branch] {
			continue
		}
		branchesToDelete = append(branchesToDelete, branch)
	}

	if len(branchesToDelete) > 0 {
		cmd = exec.Command("git", append([]string{"branch", "-d"}, branchesToDelete...)...)
		_, err = cmd.CombinedOutput()
		if err != nil {
			return 0, 0, fmt.Errorf("failed to delete branches: %w", err)
		}
	}

	os.Chdir(parentDir)

	return len(branchesToDelete), len(branches), nil
}
