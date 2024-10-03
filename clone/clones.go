package clone

import (
	"fmt"
	"path/filepath"
	"strings"
)

// CloneRepository clones a single repository from the given URL into the specified directory.
func CloneRepositories(repoURL, orgName, teamName, destPath string) error {
	if orgName != "" {
		// Clone all repositories for the organization
		if err := CloneOrgRepositories(orgName, teamName, destPath); err != nil {
			fmt.Println("Error:", err)
		}
	} else if repoURL != "" {
		// Clone a single repository
		repo := Repository{
			Name:   strings.TrimSuffix(filepath.Base(strings.Split(repoURL, "/")[1]), ".git"),
			SSHURL: repoURL,
		}
		if err := CloneRepository(repo, destPath); err != nil {
			fmt.Println("Error:", err)
		}
	} else {
		fmt.Println("Please specify either a repository URL or an organization name.")
	}

	return nil
}

// CloneOrgRepositories clones all repositories for the given organization.
func CloneOrgRepositories(orgName, teamName, baseDestPath string) error {
	// Fetch the list of repositories from the organization
	repos, err := GetOrgRepositories(orgName, teamName)
	if err != nil {
		return fmt.Errorf("failed to get repositories for organization %s: %w", orgName, err)
	}

	for _, repo := range repos {
		destPath := filepath.Join(baseDestPath, repo.Name)
		if err := CloneRepository(repo, destPath); err != nil {
			fmt.Printf("Error cloning repository %s: %v\n", repo.Name, err)
		}
	}

	return nil
}
