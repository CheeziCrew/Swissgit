package status

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// calculateAheadBehind calculates the number of commits the local branch is ahead or behind its remote tracking branch.
func calculateAheadBehind(repo *git.Repository, headRef *plumbing.Reference, branch string) (int, int) {
	remoteRef, err := repo.Reference(plumbing.NewRemoteReferenceName("origin", branch), true)
	if err != nil {
		return 0, 0 // No remote tracking branch found
	}

	var ahead, behind int

	// Calculate commits that are ahead
	cIter, _ := repo.Log(&git.LogOptions{From: headRef.Hash()})
	_ = cIter.ForEach(func(c *object.Commit) error {
		if isAncestor(c.Hash, remoteRef.Hash(), repo) {
			return storer.ErrStop
		}
		ahead++
		return nil
	})

	// Calculate commits that are behind
	cIter, _ = repo.Log(&git.LogOptions{From: remoteRef.Hash()})
	_ = cIter.ForEach(func(c *object.Commit) error {
		if isAncestor(c.Hash, headRef.Hash(), repo) {
			return storer.ErrStop
		}
		behind++
		return nil
	})

	return ahead, behind
}

// isAncestor checks if commitA is an ancestor of commitB
func isAncestor(commitA, commitB plumbing.Hash, repo *git.Repository) bool {
	cIter, _ := repo.Log(&git.LogOptions{From: commitB})
	found := false
	_ = cIter.ForEach(func(c *object.Commit) error {
		if c.Hash == commitA {
			found = true
			return storer.ErrStop
		}
		return nil
	})
	return found
}
