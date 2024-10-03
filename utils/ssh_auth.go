package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// sshAuth sets up SSH authentication using a private key file.
func SshAuth() (*ssh.PublicKeys, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not get user home directory: %w", err)
	}

	// Read the private key file
	sshKeyPath := filepath.Join(home, ".ssh", os.Getenv("SSH_KEY"))
	key, err := os.ReadFile(sshKeyPath)
	if err != nil {
		return nil, fmt.Errorf("could not read SSH key file: %w", err)
	}

	// Create the public keys object
	auth, err := ssh.NewPublicKeys("git", key, "")
	if err != nil {
		return nil, fmt.Errorf("could not create public keys: %w", err)
	}
	return auth, nil
}
