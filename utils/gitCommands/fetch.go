package gitCommands

import (
	"fmt"
	"io"

	"github.com/CheeziCrew/swissgit/utils"
	"github.com/go-git/go-git/v5"
)

func FetchRemote(repo *git.Repository) error {
	// Set up SSH authentication
	auth, err := utils.SshAuth()
	if err != nil {
		return fmt.Errorf("failed to set up SSH authentication: %w", err)
	}

	// Fetch the remote references
	err = repo.Fetch(&git.FetchOptions{
		Prune:    true,
		Progress: io.Discard,
		Auth:     auth,
	})

	// Check if there's an error
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("could not fetch remote references: %w", err)
	}

	// No error or already up to date, return nil
	return nil
}
