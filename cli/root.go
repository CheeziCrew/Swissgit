package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/CheeziCrew/swissgit/git"
	"github.com/CheeziCrew/swissgit/ops"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	ok   = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("✔")
	fail = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("✗")
	dim  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// BuildCLI creates the Cobra command tree for fast CLI usage.
func BuildCLI() *cobra.Command {
	root := &cobra.Command{
		Use:           "swissgit",
		Short:         "SwissGit — multi-repo git workflows",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.AddCommand(prCmd())
	root.AddCommand(commitCLICmd())
	root.AddCommand(cleanupCLICmd())
	root.AddCommand(statusCLICmd())
	root.AddCommand(branchesCLICmd())
	root.AddCommand(cloneCLICmd())
	root.AddCommand(automergeCLICmd())
	root.AddCommand(mergePRsCLICmd())
	root.AddCommand(enableWorkflowsCLICmd())

	return root
}

// --- Pull Request ---

func prCmd() *cobra.Command {
	var message, branch, target string
	var allFlag bool

	cmd := &cobra.Command{
		Use:   "pr",
		Short: "Commit, push & create pull request",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := resolvePaths(".", allFlag)
			for _, p := range paths {
				result := ops.CommitAndCreatePR(p, branch, message, target, nil, false)
				printResult(result.RepoName, result.Success, result.PRURL, result.Error)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "PR title / commit message")
	cmd.Flags().StringVarP(&branch, "branch", "b", "", "Feature branch name")
	cmd.Flags().StringVarP(&target, "target", "t", "main", "Target branch")
	cmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Process all repos in subdirectories")
	cmd.MarkFlagRequired("message")
	cmd.MarkFlagRequired("branch")

	return cmd
}

// --- Commit ---

func commitCLICmd() *cobra.Command {
	var message, branch string
	var allFlag bool

	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Stage, commit & push changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := resolvePaths(".", allFlag)
			for _, p := range paths {
				result := ops.CommitAndPush(p, branch, message)
				printResult(result.RepoName, result.Success, result.Branch, result.Error)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Commit message")
	cmd.Flags().StringVarP(&branch, "branch", "b", "", "Branch (optional)")
	cmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Process all repos")
	cmd.MarkFlagRequired("message")

	return cmd
}

// --- Cleanup ---

func cleanupCLICmd() *cobra.Command {
	var dropChanges, allFlag bool
	var defaultBranch string

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Reset, update main, prune branches",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := resolvePaths(".", allFlag)

			results := make([]ops.CleanupResult, len(paths))
			sem := make(chan struct{}, 10)
			var wg sync.WaitGroup

			for i, p := range paths {
				wg.Add(1)
				sem <- struct{}{}
				go func(idx int, path string) {
					defer wg.Done()
					defer func() { <-sem }()
					results[idx] = ops.CleanupRepo(path, dropChanges, defaultBranch)
				}(i, p)
			}
			wg.Wait()

			for _, r := range results {
				info := ""
				if r.PrunedBranches > 0 {
					info = fmt.Sprintf("pruned %d branches", r.PrunedBranches)
				}
				printResult(r.RepoName, r.Success, info, r.Error)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&dropChanges, "drop", "d", false, "Drop uncommitted changes")
	cmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Process all repos")
	cmd.Flags().StringVar(&defaultBranch, "default-branch", "main", "Default branch name")

	return cmd
}

// --- Status ---

func statusCLICmd() *cobra.Command {
	var allFlag, verbose bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check repo status",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := resolvePaths(".", allFlag)

			results := make([]ops.StatusResult, len(paths))
			sem := make(chan struct{}, 10)
			var wg sync.WaitGroup

			for i, p := range paths {
				wg.Add(1)
				sem <- struct{}{}
				go func(idx int, path string) {
					defer wg.Done()
					defer func() { <-sem }()
					results[idx] = ops.GetRepoStatus(path)
				}(i, p)
			}
			wg.Wait()

			for _, r := range results {
				if !verbose && r.Clean && r.Branch == r.DefaultBranch {
					continue
				}
				if r.Error != "" {
					fmt.Printf(" %s %s %s\n", fail, r.RepoName, r.Error)
					continue
				}
				fmt.Printf(" %s %s [%s]", ok, r.RepoName, r.Branch)
				if r.Ahead > 0 || r.Behind > 0 {
					fmt.Printf(" %d↑/%d↓", r.Ahead, r.Behind)
				}
				if r.Modified > 0 {
					fmt.Printf(" ~%d", r.Modified)
				}
				if r.Added > 0 {
					fmt.Printf(" +%d", r.Added)
				}
				if r.Deleted > 0 {
					fmt.Printf(" -%d", r.Deleted)
				}
				if r.Untracked > 0 {
					fmt.Printf(" ?%d", r.Untracked)
				}
				fmt.Println()
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Scan all repos")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show clean repos too")

	return cmd
}

// --- Branches ---

func branchesCLICmd() *cobra.Command {
	var allFlag bool

	cmd := &cobra.Command{
		Use:   "branches",
		Short: "List branches",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := resolvePaths(".", allFlag)
			for _, p := range paths {
				r := ops.GetBranches(p)
				if r.Error != "" {
					fmt.Printf(" %s %s %s\n", fail, r.RepoName, r.Error)
					continue
				}
				var local, remote []string
				for _, b := range r.LocalBranches {
					local = append(local, b.Name)
				}
				for _, b := range r.RemoteBranches {
					remote = append(remote, b.Name)
				}
				fmt.Printf(" %s %s [%s] local: %s remote: %s\n",
					ok, r.RepoName, r.CurrentBranch,
					strings.Join(local, ", "),
					strings.Join(remote, ", "))
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Scan all repos")
	return cmd
}

// --- Clone ---

func cloneCLICmd() *cobra.Command {
	var repoURL, orgName, teamName, destPath string

	cmd := &cobra.Command{
		Use:   "clone",
		Short: "Clone repo or org",
		RunE: func(cmd *cobra.Command, args []string) error {
			if orgName != "" {
				repos, err := ops.GetOrgRepositories(orgName, teamName)
				if err != nil {
					return err
				}
				for _, r := range repos {
					dest := filepath.Join(destPath, r.Name)
					result := ops.CloneRepository(r, dest)
					printResult(result.RepoName, result.Success, "", result.Error)
				}
				return nil
			}
			if repoURL != "" {
				result := ops.CloneFromURL(repoURL, destPath)
				printResult(result.RepoName, result.Success, "", result.Error)
				return nil
			}
			return fmt.Errorf("specify --repo or --org")
		},
	}

	cmd.Flags().StringVarP(&repoURL, "repo", "r", "", "Repo SSH URL")
	cmd.Flags().StringVarP(&orgName, "org", "o", "", "GitHub org")
	cmd.Flags().StringVarP(&teamName, "team", "t", "", "Team within org")
	cmd.Flags().StringVarP(&destPath, "path", "p", ".", "Destination")

	return cmd
}

// --- Automerge ---

func automergeCLICmd() *cobra.Command {
	var target string
	var allFlag bool

	cmd := &cobra.Command{
		Use:   "automerge",
		Short: "Enable auto-merge on PRs",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := resolvePaths(".", allFlag)
			for _, p := range paths {
				result := ops.EnableAutomerge(target, p)
				info := ""
				if result.PRNumber != "" {
					info = "PR #" + result.PRNumber
				}
				printResult(result.RepoName, result.Success, info, result.Error)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "PR search target")
	cmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Process all repos")
	cmd.MarkFlagRequired("target")

	return cmd
}

// --- Merge PRs ---

func mergePRsCLICmd() *cobra.Command {
	var orgName string
	var dryRun bool
	var batchSize, waitMin int

	cmd := &cobra.Command{
		Use:   "merge-prs",
		Short: "Merge approved pull requests in batches",
		RunE: func(cmd *cobra.Command, args []string) error {
			merged, failed := 0, 0
			batch := 0

			for {
				prs, err := ops.FetchApprovedPRs(orgName)
				if err != nil {
					return err
				}
				if len(prs) == 0 {
					if batch == 0 {
						fmt.Println("No approved PRs found.")
					} else {
						fmt.Println("\nNo more approved PRs.")
					}
					break
				}

				end := batchSize
				if end > len(prs) {
					end = len(prs)
				}
				batch++
				fmt.Printf("\nBatch %d — %d approved PR(s), merging %d\n", batch, len(prs), end)

				for _, pr := range prs[:end] {
					name := fmt.Sprintf("%s #%d", pr.Repo, pr.Number)
					if dryRun {
						fmt.Printf(" %s %s %s\n", ok, name, dim.Render(pr.Title))
						continue
					}
					result := ops.MergePR(orgName, pr.Repo, pr.Number)
					printResult(name, result.Success, pr.Title, result.Error)
					if result.Success {
						merged++
					} else {
						failed++
					}
				}

				if dryRun {
					break
				}

				fmt.Printf("\nWaiting %d minutes before next batch…\n", waitMin)
				time.Sleep(time.Duration(waitMin) * time.Minute)
			}

			if !dryRun && batch > 0 {
				fmt.Printf("\n=== Summary ===\nMerged: %d\nFailed: %d\n", merged, failed)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&orgName, "org", "o", "Sundsvallskommun", "GitHub org")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be merged")
	cmd.Flags().IntVar(&batchSize, "batch-size", 5, "PRs per batch")
	cmd.Flags().IntVar(&waitMin, "wait", 10, "Minutes between batches")

	return cmd
}

// --- Enable Workflows ---

func enableWorkflowsCLICmd() *cobra.Command {
	var orgName, workflowName, prBranch string

	cmd := &cobra.Command{
		Use:   "enable-workflows",
		Short: "Re-enable GitHub Actions disabled by inactivity",
		RunE: func(cmd *cobra.Command, args []string) error {
			repos, err := ops.FetchOrgRepoNames(orgName)
			if err != nil {
				return err
			}
			fmt.Printf("Checking %d repos in %s\n", len(repos), orgName)
			for _, repo := range repos {
				result := ops.FindAndEnableWorkflows(orgName, repo, workflowName, prBranch)
				var parts []string
				if result.EnabledCount > 0 {
					parts = append(parts, fmt.Sprintf("enabled %d", result.EnabledCount))
				}
				if result.RetriggeredPRs > 0 {
					parts = append(parts, fmt.Sprintf("retriggered %d PR(s)", result.RetriggeredPRs))
				}
				printResult(repo, result.Success, strings.Join(parts, ", "), result.Error)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&orgName, "org", "o", "Sundsvallskommun", "GitHub org")
	cmd.Flags().StringVarP(&workflowName, "workflow", "w", "Call Java CI with Maven", "Workflow name to enable (empty = all)")
	cmd.Flags().StringVar(&prBranch, "pr-branch", "", "Close/reopen PRs from this head branch to retrigger workflows")

	return cmd
}

// --- Helpers ---

func resolvePaths(root string, all bool) []string {
	if all {
		paths, err := git.DiscoverRepos(root)
		if err != nil {
			return []string{root}
		}
		return paths
	}
	return []string{root}
}

func printResult(name string, success bool, info, errMsg string) {
	icon := ok
	if !success {
		icon = fail
	}
	line := fmt.Sprintf(" %s %s", icon, name)
	if info != "" {
		line += " " + dim.Render(info)
	}
	if errMsg != "" {
		line += " " + lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render(errMsg)
	}
	fmt.Println(line)
}
