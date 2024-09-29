package status

import (
	"fmt"

	"github.com/go-git/go-git/v5"
)

// getBranchName retrieves the current branch name.
func getBranchName(repo *git.Repository) (string, error) {
	headRef, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("could not get head: %w", err)
	}
	return headRef.Name().Short(), nil
}
