package ops

import (
	"bytes"
	"net/http"
	"os/exec"
	"strings"

	"github.com/CheeziCrew/swissgit/git"
	gogit "github.com/go-git/go-git/v5"
)

// httpClient is the HTTP client used for API calls. Override in tests.
var httpClient = &http.Client{}

// plainOpen opens a git repository. Override in tests.
var plainOpen = gogit.PlainOpen

// fetchRemote fetches from origin. Override in tests.
var fetchRemote = git.FetchRemote

// pushChanges pushes to origin. Override in tests.
var pushChanges = git.PushChanges

// Package-level function variables for testability.

// ghRun runs a gh CLI command and returns trimmed stdout.
var ghRun = func(args ...string) (string, error) {
	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd := exec.Command("gh", args...)
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", &cmdError{stderr: strings.TrimSpace(errBuf.String()), err: err}
	}
	return out.String(), nil
}

// ghRunInDir runs a gh CLI command in a specific directory and returns trimmed stdout.
var ghRunInDir = func(dir string, args ...string) (string, error) {
	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd := exec.Command("gh", args...)
	cmd.Dir = dir
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", &cmdError{stderr: strings.TrimSpace(errBuf.String()), err: err}
	}
	return out.String(), nil
}

// gitRunInDir runs a git command in a specific directory.
var gitRunInDir = func(dir string, args ...string) (string, error) {
	var out bytes.Buffer
	var errBuf bytes.Buffer
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", &cmdError{stderr: strings.TrimSpace(errBuf.String()), err: err}
	}
	return out.String(), nil
}

type cmdError struct {
	stderr string
	err    error
}

func (e *cmdError) Error() string {
	if e.stderr != "" {
		return e.stderr
	}
	return e.err.Error()
}
