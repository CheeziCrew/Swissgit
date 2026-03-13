package git

import "os/exec"

// Package-level function variables for testability.

// gitRun runs a git command in the given repo directory and returns trimmed output.
var gitRun = func(repoPath string, args ...string) (string, error) {
	fullArgs := append([]string{"-C", repoPath}, args...)
	out, err := exec.Command("git", fullArgs...).Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
