package git

import (
	"fmt"
	"io"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

// PushChanges pushes the current branch to the remote using SSH auth.
func PushChanges(repo *gogit.Repository) error {
	auth, err := sshAuthFunc()
	if err != nil {
		return fmt.Errorf("failed to set up SSH authentication: %w", err)
	}

	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get the current branch: %w", err)
	}

	err = repo.Push(&gogit.PushOptions{
		RemoteURL: RemoteURL(repo),
		Progress:  io.Discard,
		Auth:      auth,
		RefSpecs: []config.RefSpec{
			config.RefSpec(head.Name().String() + ":" + head.Name().String()),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to push to remote: %w", err)
	}
	return nil
}
