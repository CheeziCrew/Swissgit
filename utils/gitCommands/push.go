package gitCommands

import (
	"fmt"
	"io"

	"github.com/CheeziCrew/swissgit/utils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

func PushChanges(repo *git.Repository) error {
	// Set up SSH authentication
	auth, err := utils.SshAuth()
	if err != nil {
		return fmt.Errorf("failed to set up SSH authentication: %w", err)
	}

	// Get the current branch reference
	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get the current branch: %w", err)
	}

	// Push only the current branch to the remote repository
	err = repo.Push(&git.PushOptions{
		Progress: io.Discard,
		Auth:     auth,
		RefSpecs: []config.RefSpec{
			config.RefSpec(head.Name().String() + ":" + head.Name().String()),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to push to remote: %w", err)
	}
	return nil
}
