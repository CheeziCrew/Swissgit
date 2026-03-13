package git

import (
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// sshAuthFunc wraps SSHAuth for testability. Override in tests.
var sshAuthFunc = func() (*ssh.PublicKeys, error) {
	return SSHAuth()
}
