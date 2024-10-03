package validation

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
)

func CheckForApiChange(repo *git.Repository) (bool, bool, bool, error) {
	// Get the working directory for the repository
	worktree, err := repo.Worktree()
	if err != nil {

		return false, false, false, fmt.Errorf("error getting worktree: %w", err)
	}

	// Get the status of the working directory
	status, err := worktree.Status()
	if err != nil {
		return false, false, false, fmt.Errorf("error getting status: %w", err)
	}

	resourceFileFound := false
	openApiFileFound := false
	versionTagChanged := false

	// Iterate over the status to find changed files
	for filePath := range status {
		if strings.HasSuffix(filePath, "Resource.java") {
			resourceFileFound = true
		}
		if filePath == "openapi.yaml" || filePath == "openapi.yml" {
			openApiFileFound = true
		}
		if filePath == "pom.xml" {
			versionTagChanged, err = checkPomVersion("pom.xml")
			if err != nil {
				return false, false, false, fmt.Errorf("error checking pom.xml: %w", err)
			}

		}

	}
	return resourceFileFound, openApiFileFound, versionTagChanged, nil
}

// checkPomVersion reads the pom.xml and looks for the <version> tag
func checkPomVersion(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inParentBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Detect start and end of <parent> block
		if strings.Contains(line, "<parent>") {
			inParentBlock = true
		}
		if strings.Contains(line, "</parent>") {
			inParentBlock = false
		}

		// Find <version> tag outside of <parent> block
		if !inParentBlock && strings.Contains(line, "<version>") {
			// Extract the version content
			version := extractTagValue(line, "version")
			if version != "" {
				return true, nil
			}
		}
	}

	return false, scanner.Err()
}

// extractTagValue extracts the content of a specific XML tag from a line
func extractTagValue(line, tag string) string {
	openTag := fmt.Sprintf("<%s>", tag)
	closeTag := fmt.Sprintf("</%s>", tag)

	startIdx := strings.Index(line, openTag)
	endIdx := strings.Index(line, closeTag)

	if startIdx != -1 && endIdx != -1 {
		return line[startIdx+len(openTag) : endIdx]
	}
	return ""
}
