package clone

import (
	"fmt"
	"io"
	"os"

	"github.com/CheeziCrew/swissgit/utils"
	"github.com/fatih/color"
	"github.com/go-git/go-git/v5"
)

// CloneRepository clones a single repository from the given URL into the specified directory using SSH.
func CloneRepository(repo Repository, destPath string) error {

	done := make(chan bool)
	go utils.ShowSpinner("Cloning "+repo.Name, done)
	// Create the directory if it doesn't exist
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		if err := os.MkdirAll(destPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", destPath, err)
		}
	}

	// Set up SSH authentication
	auth, err := utils.SshAuth()
	if err != nil {
		return fmt.Errorf("failed to set up SSH authentication: %w", err)
	}

	// Clone the repository using SSH
	_, err = git.PlainClone(destPath, false, &git.CloneOptions{
		URL:      repo.SSHURL,
		Progress: io.Discard, // Suppress default output
		Auth:     auth,
	})

	// Signal the spinner to stop
	done <- true

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	if err != nil {
		fmt.Printf("\rCloned failed for %s [%s] Error: %s\n", repo.Name, red("✗"), err)
		return nil
	}

	fmt.Printf("\rCloned %s [%s] \n", repo.Name, green("✔"))
	return nil
}
