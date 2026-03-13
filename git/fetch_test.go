package git

import (
	"fmt"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
)

func newMemRepo(t *testing.T) *gogit.Repository {
	t.Helper()
	repo, err := gogit.Init(memory.NewStorage(), nil)
	if err != nil {
		t.Fatalf("failed to create in-memory repo: %v", err)
	}
	return repo
}

func TestFetchRemote_SSHAuthError(t *testing.T) {
	orig := sshAuthFunc
	t.Cleanup(func() { sshAuthFunc = orig })

	sshAuthFunc = func() (*ssh.PublicKeys, error) {
		return nil, fmt.Errorf("no ssh key")
	}

	repo := newMemRepo(t)
	err := FetchRemote(repo)
	if err == nil {
		t.Fatal("expected error when SSH auth fails")
	}
	if got := err.Error(); got == "" {
		t.Error("expected non-empty error message")
	}
}

func TestPushChanges_SSHAuthError(t *testing.T) {
	orig := sshAuthFunc
	t.Cleanup(func() { sshAuthFunc = orig })

	sshAuthFunc = func() (*ssh.PublicKeys, error) {
		return nil, fmt.Errorf("no ssh key")
	}

	repo := newMemRepo(t)
	err := PushChanges(repo)
	if err == nil {
		t.Fatal("expected error when SSH auth fails")
	}
}

func TestPullChanges_SSHAuthError(t *testing.T) {
	orig := sshAuthFunc
	t.Cleanup(func() { sshAuthFunc = orig })

	sshAuthFunc = func() (*ssh.PublicKeys, error) {
		return nil, fmt.Errorf("no ssh key")
	}

	repo := newMemRepo(t)
	wt, err := repo.Worktree()
	if err != nil {
		// in-memory repos without filesystem won't have worktrees
		t.Skip("in-memory repo has no worktree")
	}
	err = PullChanges(wt)
	if err == nil {
		t.Fatal("expected error when SSH auth fails")
	}
}
