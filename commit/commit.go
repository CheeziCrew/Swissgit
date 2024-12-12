package commit

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/CheeziCrew/swissgit/utils"
	gc "github.com/CheeziCrew/swissgit/utils/gitCommands"
	"github.com/CheeziCrew/swissgit/utils/validation"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
)

var green = color.New(color.FgGreen).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()

// CommitChanges adds all changes, checks out a new branch, commits with the provided message, and pushes to the remote repository
func CommitChanges(repoPath, branchName, commitMessage string) error {

	repoName, err := utils.GetRepoName(repoPath)
	if err != nil {
		return fmt.Errorf("could not get repository name: %w", err)
	}

	statusMessage := fmt.Sprintf("%s: committing and pushing", repoName)
	done := make(chan bool)
	go utils.ShowSpinner(statusMessage, done)

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	if err := commitAndPush(repo, repoPath, branchName, commitMessage); err != nil {
		done <- true
		fmt.Printf("\r%s failed [%s]: %s\n", statusMessage, red("x"), red(err.Error()))
		return err
	}

	done <- true
	statusMessage = fmt.Sprintf("%s: committed and pushed to remote", repoName)
	fmt.Printf("\r%s [%s] \n", statusMessage, green("✔"))
	return nil
}

func commitAndPush(repo *git.Repository, repoPath, branchName, commitMessage string) error {

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}
	err = worktree.AddGlob(".")
	if err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}

	//gc.PullChanges(worktree)

	status, err := worktree.Status()
	if err != nil {
		return fmt.Errorf("failed to get status for repository: %w", err)
	}

	modified, added, deleted, untracked := utils.CountChanges(status)
	if modified == 0 && added == 0 && deleted == 0 && untracked == 0 {
		return fmt.Errorf("no changes to commit")
	}

	// Create a new branch
	if branchName != "" {
		createNewBranch(branchName, worktree)
	} else if branchName == "" {
		branchName, _ = utils.GetBranchName(repo)
	}

	commit(branchName, commitMessage, repoPath)
	err = gc.PushChanges(repo)
	if err != nil {
		return fmt.Errorf("failed to push changes: %w", err)
	}
	return nil
}

func commit(branchName, commitMessage, repoPath string) error {
	fullCommitMessage := fmt.Sprintf("%s: %s", branchName, commitMessage)

	cmd := exec.Command("git", "-C", repoPath, "commit", "-m", fullCommitMessage)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}
	return nil
}

func verifyApi(repo *git.Repository, repoName string, statusMessage string, done chan bool) bool {

	resourceFileFound, openApiFileFound, versionTagChanged, err := validation.CheckForApiChange(repo)
	if err != nil {
		errMsg := fmt.Sprintf("%s: Could not check for API change:", filepath.Base(repoName))
		fmt.Printf("\r%s [%s]: %s\n", errMsg, red("x"), err)
		done <- true
		return false
	}

	yellow := color.New(color.FgYellow).SprintFunc()

	if resourceFileFound && !openApiFileFound {
		done <- true
		fmt.Print("\r")
		fmt.Println(yellow(repoName + ": [Warning!] API specification is not up to date"))
		if !getUserConfirmation() {
			return false
		}
		fmt.Print("\r\n")
		go utils.ShowSpinner(statusMessage, done)
	}
	if resourceFileFound && !versionTagChanged {
		done <- true
		fmt.Print("\r")
		fmt.Println(yellow(repoName + ": [Warning!] It seems like a new API was added, but the pom.xml file was not updated"))
		if !getUserConfirmation() {
			return false
		}
		fmt.Print("\r\n")
		go utils.ShowSpinner(statusMessage, done)
	}
	if resourceFileFound && !openApiFileFound && !versionTagChanged {
		done <- true
		fmt.Print("\r")
		fmt.Println(yellow(repoName + ": [Warning!] The pom.xml file was changed, but the <version> tag was not updated"))
		if !getUserConfirmation() {
			return false
		}
		fmt.Print("\r\n")
		go utils.ShowSpinner(statusMessage, done)
	}
	return true
}

func getUserConfirmation() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Continue with the push? [Y/N]: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToUpper(input))
	if input == "Y" || input == "YES" || input == "y" {
		fmt.Print("[" + green("✔") + "]" + " Continuing with the push")
		return true
	} else {
		fmt.Print("[" + red("x") + "]" + " Exiting the push")
		return false
	}
}
