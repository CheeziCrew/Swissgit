package git

import (
	"fmt"
	"io"

	gogit "github.com/go-git/go-git/v5"
)

// PullChanges pulls changes from the remote using SSH auth.
func PullChanges(worktree *gogit.Worktree) error {
	auth, err := sshAuthFunc()
	if err != nil {
		return fmt.Errorf("failed to set up SSH authentication: %w", err)
	}

	err = worktree.Pull(&gogit.PullOptions{
		Progress: io.Discard,
		Auth:     auth,
	})
	if err != nil && err != gogit.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to pull changes: %w", err)
	}
	return nil
}
