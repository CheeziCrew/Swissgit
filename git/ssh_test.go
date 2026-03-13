package git

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestSSHAuth_NoSSHKey(t *testing.T) {
	// Reset sync.Once for this test.
	sshOnce = sync.Once{}
	cachedAuth = nil
	cachedErr = nil
	t.Cleanup(func() {
		sshOnce = sync.Once{}
		cachedAuth = nil
		cachedErr = nil
	})

	t.Setenv("SSH_KEY", "")

	_, err := SSHAuth()
	if err == nil {
		t.Fatal("expected error when SSH_KEY is not set")
	}
}

func TestSSHAuth_MissingKeyFile(t *testing.T) {
	sshOnce = sync.Once{}
	cachedAuth = nil
	cachedErr = nil
	t.Cleanup(func() {
		sshOnce = sync.Once{}
		cachedAuth = nil
		cachedErr = nil
	})

	t.Setenv("SSH_KEY", "nonexistent_key_file")
	t.Setenv("HOME", t.TempDir())

	_, err := SSHAuth()
	if err == nil {
		t.Fatal("expected error when SSH key file doesn't exist")
	}
}

func TestSSHAuth_InvalidKeyFile(t *testing.T) {
	sshOnce = sync.Once{}
	cachedAuth = nil
	cachedErr = nil
	t.Cleanup(func() {
		sshOnce = sync.Once{}
		cachedAuth = nil
		cachedErr = nil
	})

	home := t.TempDir()
	sshDir := filepath.Join(home, ".ssh")
	os.MkdirAll(sshDir, 0700)
	os.WriteFile(filepath.Join(sshDir, "bad_key"), []byte("not a valid ssh key"), 0600)

	t.Setenv("SSH_KEY", "bad_key")
	t.Setenv("HOME", home)

	_, err := SSHAuth()
	if err == nil {
		t.Fatal("expected error for invalid SSH key")
	}
}
