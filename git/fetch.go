package git

import (
	"fmt"
	"io"

	gogit "github.com/go-git/go-git/v5"
)

// FetchRemote fetches from the origin remote with SSH auth.
func FetchRemote(repo *gogit.Repository) error {
	auth, err := sshAuthFunc()
	if err != nil {
		return fmt.Errorf("failed to set up SSH authentication: %w", err)
	}

	err = repo.Fetch(&gogit.FetchOptions{
		RemoteURL: RemoteURL(repo),
		Prune:     true,
		Progress:  io.Discard,
		Auth:      auth,
	})

	if err != nil && err != gogit.NoErrAlreadyUpToDate {
		return fmt.Errorf("could not fetch remote references: %w", err)
	}
	return nil
}
