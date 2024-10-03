package gitCommands

import (
	"fmt"
	"io"

	"github.com/CheeziCrew/swissgit/utils"
	"github.com/go-git/go-git/v5"
)

func PullChanges(worktree *git.Worktree) error {

	// Set up SSH authentication
	auth, err := utils.SshAuth()
	if err != nil {
		return fmt.Errorf("failed to set up SSH authentication: %w", err)
	}

	// Pull changes from the remote repository
	err = worktree.Pull(&git.PullOptions{
		Progress: io.Discard,
		Auth:     auth,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to pull changes: %w", err)
	}

	return nil
}
