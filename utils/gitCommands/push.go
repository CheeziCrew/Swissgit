package gitCommands

import (
	"fmt"
	"io"

	"github.com/CheeziCrew/swissgo/utils"
	"github.com/go-git/go-git/v5"
)

func PushChanges(repo *git.Repository) error {

	// Set up SSH authentication
	auth, err := utils.SshAuth()
	if err != nil {
		return fmt.Errorf("failed to set up SSH authentication: %w", err)
	}

	// Push the commit to the remote repository
	err = repo.Push(&git.PushOptions{
		Progress: io.Discard,
		Auth:     auth,
	})
	if err != nil {
		return fmt.Errorf("failed to push to remote: %w", err)
	}
	return nil
}
