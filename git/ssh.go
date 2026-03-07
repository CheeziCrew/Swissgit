package git

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

var (
	sshOnce    sync.Once
	cachedAuth *ssh.PublicKeys
	cachedErr  error
)

// SSHAuth sets up SSH authentication using the private key file
// specified by the SSH_KEY env var under ~/.ssh/.
// Result is cached after the first call.
func SSHAuth() (*ssh.PublicKeys, error) {
	sshOnce.Do(func() {
		var home string
		home, cachedErr = os.UserHomeDir()
		if cachedErr != nil {
			cachedErr = fmt.Errorf("could not get user home directory: %w", cachedErr)
			return
		}

		sshKey := os.Getenv("SSH_KEY")
		if sshKey == "" {
			cachedErr = fmt.Errorf("SSH_KEY environment variable not set")
			return
		}

		var key []byte
		key, cachedErr = os.ReadFile(filepath.Join(home, ".ssh", sshKey))
		if cachedErr != nil {
			cachedErr = fmt.Errorf("could not read SSH key file: %w", cachedErr)
			return
		}

		cachedAuth, cachedErr = ssh.NewPublicKeys("git", key, "")
		if cachedErr != nil {
			cachedErr = fmt.Errorf("could not create public keys: %w", cachedErr)
		}
	})
	return cachedAuth, cachedErr
}
