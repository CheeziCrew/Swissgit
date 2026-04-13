package git

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/skeema/knownhosts"
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
			return
		}

		hostKeyCallback, err := knownhosts.New(filepath.Join(home, ".ssh", "known_hosts"))
		if err != nil {
			cachedErr = fmt.Errorf("could not parse known_hosts: %w", err)
			return
		}
		cachedAuth.HostKeyCallback = hostKeyCallback.HostKeyCallback()
	})
	return cachedAuth, cachedErr
}

// sshConfigEntry holds parsed Host block values from ~/.ssh/config.
type sshConfigEntry struct {
	HostName string
	Port     string
}

// parseSSHConfig reads ~/.ssh/config and returns entries keyed by Host pattern.
func parseSSHConfig() map[string]sshConfigEntry {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	f, err := os.Open(filepath.Join(home, ".ssh", "config"))
	if err != nil {
		return nil
	}
	defer f.Close()

	entries := make(map[string]sshConfigEntry)
	var currentHost string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])
		switch key {
		case "host":
			currentHost = val
		case "hostname":
			if currentHost != "" {
				e := entries[currentHost]
				e.HostName = val
				entries[currentHost] = e
			}
		case "port":
			if currentHost != "" {
				e := entries[currentHost]
				e.Port = val
				entries[currentHost] = e
			}
		}
	}
	return entries
}

// RewriteRemoteURL rewrites a git SSH URL based on ~/.ssh/config.
// e.g. git@github.com:org/repo.git -> ssh://git@ssh.github.com:443/org/repo.git
func RewriteRemoteURL(url string) string {
	// Only handle SCP-style URLs: git@host:path
	if !strings.Contains(url, "@") || !strings.Contains(url, ":") || strings.HasPrefix(url, "ssh://") || strings.HasPrefix(url, "https://") {
		return url
	}

	parts := strings.SplitN(url, "@", 2)
	user := parts[0]
	rest := parts[1]
	colonIdx := strings.Index(rest, ":")
	if colonIdx < 0 {
		return url
	}
	host := rest[:colonIdx]
	path := rest[colonIdx+1:]

	entries := parseSSHConfig()
	if entries == nil {
		return url
	}

	entry, ok := entries[host]
	if !ok {
		return url
	}

	targetHost := host
	if entry.HostName != "" {
		targetHost = entry.HostName
	}
	port := "22"
	if entry.Port != "" {
		port = entry.Port
	}

	// If nothing changed, return original
	if targetHost == host && port == "22" {
		return url
	}

	return fmt.Sprintf("ssh://%s@%s:%s/%s", user, targetHost, port, path)
}

// RemoteURL returns the rewritten origin URL for a repository.
func RemoteURL(repo *gogit.Repository) string {
	remotes, err := repo.Remotes()
	if err != nil || len(remotes) == 0 {
		return ""
	}
	urls := remotes[0].Config().URLs
	if len(urls) == 0 {
		return ""
	}
	return RewriteRemoteURL(urls[0])
}
