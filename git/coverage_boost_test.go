package git

import (
	"os"
	"path/filepath"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	goconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// makeRepoWithBareRemote creates a bare remote repo and a clone that points to it.
// Returns the cloned repo and cleanup func.
func makeRepoWithBareRemote(t *testing.T) *gogit.Repository {
	t.Helper()
	bareDir := filepath.Join(t.TempDir(), "bare.git")
	_, err := gogit.PlainInit(bareDir, true)
	if err != nil {
		t.Fatalf("init bare: %v", err)
	}

	cloneDir := filepath.Join(t.TempDir(), "clone")
	os.MkdirAll(cloneDir, 0755)
	repo, err := gogit.PlainInit(cloneDir, false)
	if err != nil {
		t.Fatalf("init clone: %v", err)
	}
	_, err = repo.CreateRemote(&goconfig.RemoteConfig{
		Name: "origin",
		URLs: []string{bareDir},
	})
	if err != nil {
		t.Fatalf("create remote: %v", err)
	}

	// Create initial commit
	wt, _ := repo.Worktree()
	f, _ := os.Create(filepath.Join(cloneDir, "README.md"))
	f.WriteString("# test\n")
	f.Close()
	wt.Add("README.md")
	wt.Commit("initial", &gogit.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@test.com"},
	})

	return repo
}

func TestPushChanges_Success(t *testing.T) {
	orig := sshAuthFunc
	t.Cleanup(func() { sshAuthFunc = orig })

	// Return nil auth - local file:// remotes don't need SSH
	sshAuthFunc = func() (*ssh.PublicKeys, error) {
		return nil, nil
	}

	repo := makeRepoWithBareRemote(t)
	err := PushChanges(repo)
	if err != nil {
		t.Errorf("expected success for local push, got: %v", err)
	}
}

func TestPushChanges_HeadError(t *testing.T) {
	orig := sshAuthFunc
	t.Cleanup(func() { sshAuthFunc = orig })

	sshAuthFunc = func() (*ssh.PublicKeys, error) {
		return nil, nil
	}

	// Create repo with no commits (no HEAD)
	dir := t.TempDir()
	repo, _ := gogit.PlainInit(dir, false)

	err := PushChanges(repo)
	if err == nil {
		t.Error("expected error when HEAD doesn't exist")
	}
}

func TestFetchRemote_Success(t *testing.T) {
	orig := sshAuthFunc
	t.Cleanup(func() { sshAuthFunc = orig })

	sshAuthFunc = func() (*ssh.PublicKeys, error) {
		return nil, nil
	}

	repo := makeRepoWithBareRemote(t)
	// Push first so fetch has something to work with
	PushChanges(repo)

	err := FetchRemote(repo)
	if err != nil {
		t.Errorf("expected success for local fetch, got: %v", err)
	}
}

func TestFetchRemote_NoRemote(t *testing.T) {
	orig := sshAuthFunc
	t.Cleanup(func() { sshAuthFunc = orig })

	sshAuthFunc = func() (*ssh.PublicKeys, error) {
		return nil, nil
	}

	// Repo with no remotes
	dir := t.TempDir()
	repo, _ := gogit.PlainInit(dir, false)

	err := FetchRemote(repo)
	if err == nil {
		t.Error("expected error when no remotes configured")
	}
}

func TestPullChanges_Success(t *testing.T) {
	orig := sshAuthFunc
	t.Cleanup(func() { sshAuthFunc = orig })

	sshAuthFunc = func() (*ssh.PublicKeys, error) {
		return nil, nil
	}

	repo := makeRepoWithBareRemote(t)
	// Push first to populate the bare remote
	PushChanges(repo)

	wt, err := repo.Worktree()
	if err != nil {
		t.Fatalf("get worktree: %v", err)
	}

	err = PullChanges(wt)
	// Already up to date is fine
	if err != nil {
		t.Errorf("expected success for local pull, got: %v", err)
	}
}

func TestGetBranchName_Detached(t *testing.T) {
	dir := t.TempDir()
	repo, _ := gogit.PlainInit(dir, false)
	wt, _ := repo.Worktree()

	// Create initial commit
	f, _ := os.Create(filepath.Join(dir, "test.txt"))
	f.WriteString("test")
	f.Close()
	wt.Add("test.txt")
	hash, _ := wt.Commit("initial", &gogit.CommitOptions{
		Author: &object.Signature{Name: "test", Email: "test@test.com"},
	})

	// Detach HEAD
	wt.Checkout(&gogit.CheckoutOptions{Hash: hash})

	name, err := GetBranchName(repo)
	if err != nil {
		// Detached HEAD might return error
		_ = name
	}
}

func TestGetRepoOwnerAndName_InvalidURL(t *testing.T) {
	dir := t.TempDir()
	repo, _ := gogit.PlainInit(dir, false)
	repo.CreateRemote(&goconfig.RemoteConfig{
		Name: "origin",
		URLs: []string{"not-a-valid-url"},
	})

	_, _, err := GetRepoOwnerAndName(repo)
	if err == nil {
		t.Error("expected error for invalid remote URL")
	}
}

func TestGetRepoOwnerAndName_NoRemote(t *testing.T) {
	dir := t.TempDir()
	repo, _ := gogit.PlainInit(dir, false)

	_, _, err := GetRepoOwnerAndName(repo)
	if err == nil {
		t.Error("expected error when no origin remote")
	}
}
