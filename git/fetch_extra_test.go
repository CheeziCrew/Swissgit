package git

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

func TestFetchRemote_SSHAuthFails(t *testing.T) {
	orig := sshAuthFunc
	t.Cleanup(func() { sshAuthFunc = orig })

	sshAuthFunc = func() (*ssh.PublicKeys, error) {
		return nil, fmt.Errorf("SSH key not available")
	}

	err := FetchRemote(nil)
	if err == nil {
		t.Error("expected error when SSH auth fails")
	}
	if err != nil && !strings.Contains(err.Error(), "SSH authentication") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPullChanges_SSHAuthFails(t *testing.T) {
	orig := sshAuthFunc
	t.Cleanup(func() { sshAuthFunc = orig })

	sshAuthFunc = func() (*ssh.PublicKeys, error) {
		return nil, fmt.Errorf("SSH key not found")
	}

	err := PullChanges(nil)
	if err == nil {
		t.Error("expected error when SSH auth fails")
	}
	if err != nil && !strings.Contains(err.Error(), "SSH authentication") {
		t.Errorf("unexpected error: %v", err)
	}
}
