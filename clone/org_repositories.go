package clone

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// Repository represents a GitHub repository
type Repository struct {
	Name   string `json:"name"`
	SSHURL string `json:"ssh_url"`
}

// GetOrgRepositories fetches repositories for a given GitHub organization
func GetOrgRepositories(orgName, teamName string) ([]Repository, error) {
	accessToken := os.Getenv("GITHUB_TOKEN")
	if accessToken == "" {
		return nil, fmt.Errorf("GitHub access token not provided. Make sure to set the GITHUB_TOKEN environment variable")
	}

	var allRepositories []Repository
	url := ""

	if teamName != "" {
		url = fmt.Sprintf("https://api.github.com/orgs/%s/teams/%s/repos", orgName, teamName)
	} else {
		url = fmt.Sprintf("https://api.github.com/orgs/%s/repos", orgName)
	}

	client := &http.Client{}
	page := 1

	for {
		// Add pagination parameters to the URL
		paginatedURL := fmt.Sprintf("%s?page=%d&per_page=100", url, page)
		req, err := http.NewRequest("GET", paginatedURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Add authorization header
		req.Header.Set("Authorization", "token "+accessToken)

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to make request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to fetch repositories, status: %s", resp.Status)
		}

		// Decode the current page's repositories
		var repositories []Repository
		if err := json.NewDecoder(resp.Body).Decode(&repositories); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		// Append the repositories to the overall list
		allRepositories = append(allRepositories, repositories...)

		// Check for the "Link" header to see if there's a "next" page
		linkHeader := resp.Header.Get("Link")
		if !strings.Contains(linkHeader, `rel="next"`) {
			break
		}

		// Increment the page number for the next iteration
		page++
	}

	return allRepositories, nil
}
