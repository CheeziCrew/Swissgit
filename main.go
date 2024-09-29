package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/CheeziCrew/swissgit/branches"
	"github.com/CheeziCrew/swissgit/cleanup"
	"github.com/CheeziCrew/swissgit/clone"
	"github.com/CheeziCrew/swissgit/commit"
	"github.com/CheeziCrew/swissgit/pull_request"
	"github.com/CheeziCrew/swissgit/status"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

func init() {
	// Get the directory where the executable is located
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("Error getting executable path: %v\n", err)
		return
	}
	exeDir := filepath.Dir(exePath)

	// Load the .env file from the executable's directory
	envPath := filepath.Join(exeDir, ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("No .env file found at %s. Using environment variables from the system.\n", envPath)
	}
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "swissgit",
		Short: "SwissGit - A versatile tool for handling Git repositories",
	}

	// Add the status and branches commands
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(branchesCmd())
	rootCmd.AddCommand(cloneCmd())
	rootCmd.AddCommand(commitCmd())
	rootCmd.AddCommand(pullRequestCmd())
	rootCmd.AddCommand(cleanupCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func statusCmd() *cobra.Command {
	var repoPath string
	var verbose bool
	var allFlag bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check the status of repositories",
		Run: func(cmd *cobra.Command, args []string) {
			status.ScanAndCheckStatus(repoPath, allFlag, verbose)
		},
	}

	cmd.Flags().StringVarP(&repoPath, "path", "p", ".", "Path to the directory containing repositories")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show all repositories, even if on 'main' and clean")
	cmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Recursively scan one level of subdirectories for repositories")
	return cmd
}

func branchesCmd() *cobra.Command {
	var repoPath string
	var allFlag bool
	var verbose bool

	cmd := &cobra.Command{
		Use:   "branches",
		Short: "List local, remote, and stale branches in the repository",
		Run: func(cmd *cobra.Command, args []string) {
			branches.ScanAndPrintBranches(repoPath, allFlag, verbose)
		},
	}

	cmd.Flags().StringVarP(&repoPath, "path", "p", ".", "Path to the directory containing the repository")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show all repositories, even if on 'main' and clean")
	cmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Recursively scan one level of subdirectories for repositories")
	return cmd
}

func cloneCmd() *cobra.Command {
	var repoURL string
	var orgName string
	var destPath string
	var teamName string

	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone a repository or all repositories from a GitHub organization",
		Run: func(cmd *cobra.Command, args []string) {
			clone.CloneRepositories(repoURL, orgName, teamName, destPath)
		},
	}

	cmd.Flags().StringVarP(&repoURL, "repo", "r", "", "URL of the repository to clone")
	cmd.Flags().StringVarP(&orgName, "org", "o", "", "Name of the GitHub organization")
	cmd.Flags().StringVarP(&destPath, "path", "p", ".", "Destination path to clone the repository")
	cmd.Flags().StringVarP(&teamName, "team", "t", "", "Name of the GitHub team within the organization")

	return cmd
}

func commitCmd() *cobra.Command {
	var repoPath, commitMessage, branch string

	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Add all files, commit changes, and push to the remote repository",
		Run: func(cmd *cobra.Command, args []string) {
			if repoPath == "" {
				repoPath, _ = os.Getwd()
			}

			err := commit.CommitChanges(repoPath, branch, commitMessage)
			if err != nil {
				return
			}
		},
	}

	cmd.Flags().StringVarP(&repoPath, "path", "p", ".", "Path to the repository (optional, defaults to current directory)")
	cmd.Flags().StringVarP(&commitMessage, "message", "m", "", "Commit message")
	cmd.Flags().StringVarP(&branch, "branch", "b", "", "Branch to create and push changes to before creating the pull request (optional, defaults to current branch)")
	cmd.MarkFlagRequired("message")
	return cmd
}

func pullRequestCmd() *cobra.Command {
	var repoPath, commitMessage, branch, target string
	var typeOfChanges []string
	var breakingChange, allFlag bool

	changeTypes := []string{
		"Bug fix",
		"New feature",
		"Removed feature",
		"Code style update (formatting etc.)",
		"Refactoring (no functional changes, no api changes)",
		"Build related changes",
		"Documentation content changes",
	}

	cmd := &cobra.Command{
		Use:   "pullrequest",
		Short: "Commit all changes and create a pull request on GitHub",
		Run: func(cmd *cobra.Command, args []string) {
			if repoPath == "" {
				repoPath, _ = os.Getwd()
			}

			// Interactive prompt for type of changes if not provided
			fmt.Println("What changes does this contain? Type each number that applies [e.g., 125]:")
			for i, change := range changeTypes {
				fmt.Printf("%d: %s\n", i+1, change)
			}

			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter numbers: ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			for _, numStr := range strings.Split(input, "") {
				num, err := strconv.Atoi(numStr)
				if err == nil && num >= 1 && num <= len(changeTypes) {
					typeOfChanges = append(typeOfChanges, changeTypes[num-1])
				}
			}

			reader = bufio.NewReader(os.Stdin)
			fmt.Print("Is this a breaking change? [Y/N]: ")
			input, _ = reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToUpper(input))
			breakingChange = (input == "Y" || input == "YES" || input == "y")

			// Run the commit and pull request process
			err := pull_request.CommitAndPull(repoPath, branch, commitMessage, target, typeOfChanges, breakingChange, allFlag)
			if err != nil {
				return
			}
		},
	}

	cmd.Flags().StringVarP(&repoPath, "path", "p", "", "Path to the repository (optional, defaults to current directory)")
	cmd.Flags().StringVarP(&commitMessage, "message", "m", "", "Commit message (will be used as the title of the pull request)")
	cmd.Flags().StringVarP(&branch, "branch", "b", "", "Branch to create and push changes to before creating the pull request")
	cmd.Flags().StringVarP(&target, "target", "t", "main", "Name of the branch you want the changes pulled into (optional, defaults to 'main')")
	cmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Recursively scan one level of subdirectories for repositories")

	cmd.MarkFlagRequired("message")
	cmd.MarkFlagRequired("branch")

	return cmd
}

func cleanupCmd() *cobra.Command {
	var repoPath string
	var dropChanges bool
	var allFlag bool

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Reset changes, update the main branch, and prune branches",
		Run: func(cmd *cobra.Command, args []string) {
			if repoPath == "" {
				repoPath, _ = os.Getwd()
			}

			err := cleanup.Cleanup(cleanup.CleanupOptions{
				RepoPath:    repoPath,
				AllFlag:     allFlag,
				DropChanges: dropChanges,
			})
			if err != nil {
				fmt.Println(err)
			}
		},
	}

	cmd.Flags().StringVarP(&repoPath, "path", "p", "", "Path to the repository (optional, defaults to current directory)")
	cmd.Flags().BoolVarP(&dropChanges, "drop", "d", false, "Drop all changes in the repository")
	cmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Recursively scan one level of subdirectories for repositories")

	return cmd
}
