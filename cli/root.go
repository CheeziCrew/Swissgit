package cli

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/CheeziCrew/swissgit/git"
	"github.com/CheeziCrew/swissgit/ops"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
)

var jsonOutput bool

const (
	descGitHubOrg    = "GitHub org"
	fmtStatusLine    = " %s %s %s\n"
	descProcessRepos = "Process all repos"
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

	root.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output JSON instead of styled text")

	root.AddCommand(prCmd())
	root.AddCommand(commitCLICmd())
	root.AddCommand(cleanupCLICmd())
	root.AddCommand(statusCLICmd())
	root.AddCommand(branchesCLICmd())
	root.AddCommand(cloneCLICmd())
	root.AddCommand(automergeCLICmd())
	root.AddCommand(mergePRsCLICmd())
	root.AddCommand(enableWorkflowsCLICmd())
	root.AddCommand(teamPRsCLICmd())

	return root
}

func prCmd() *cobra.Command {
	var message, branch, target string
	var allFlag bool
	cfg := ops.LoadConfig()

	cmd := &cobra.Command{
		Use:   "pr",
		Short: "Commit, push & create pull request",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := resolvePaths(".", allFlag)
			var results []ops.PRResult
			for _, p := range paths {
				results = append(results, ops.CommitAndCreatePR(p, branch, message, target, nil, false))
			}
			if jsonOutput {
				return printJSON(results)
			}
			for _, r := range results {
				printResult(r.RepoName, r.Success, r.PRURL, r.Error)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "PR title / commit message")
	cmd.Flags().StringVarP(&branch, "branch", "b", "", "Feature branch name")
	cmd.Flags().StringVarP(&target, "target", "t", cfg.TargetBranch, "Target branch")
	cmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Process all repos in subdirectories")
	cmd.MarkFlagRequired("message")
	cmd.MarkFlagRequired("branch")

	return cmd
}

func commitCLICmd() *cobra.Command {
	var message, branch string
	var allFlag bool

	cmd := &cobra.Command{
		Use:   "commit",
		Short: "Stage, commit & push changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := resolvePaths(".", allFlag)
			var results []ops.CommitResult
			for _, p := range paths {
				results = append(results, ops.CommitAndPush(p, branch, message))
			}
			if jsonOutput {
				return printJSON(results)
			}
			for _, r := range results {
				printResult(r.RepoName, r.Success, r.Branch, r.Error)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Commit message")
	cmd.Flags().StringVarP(&branch, "branch", "b", "", "Branch (optional)")
	cmd.Flags().BoolVarP(&allFlag, "all", "a", false, descProcessRepos)
	cmd.MarkFlagRequired("message")

	return cmd
}

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

			if jsonOutput {
				return printJSON(results)
			}
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
	cmd.Flags().BoolVarP(&allFlag, "all", "a", false, descProcessRepos)
	cmd.Flags().StringVar(&defaultBranch, "default-branch", "main", "Default branch name")

	return cmd
}

func printStatusResult(r ops.StatusResult) {
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

			if jsonOutput {
				return printJSON(results)
			}
			for _, r := range results {
				if !verbose && r.Clean && r.Branch == r.DefaultBranch {
					continue
				}
				if r.Error != "" {
					fmt.Printf(fmtStatusLine, fail, r.RepoName, r.Error)
					continue
				}
				printStatusResult(r)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Scan all repos")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show clean repos too")

	return cmd
}

func branchesCLICmd() *cobra.Command {
	var allFlag bool

	cmd := &cobra.Command{
		Use:   "branches",
		Short: "List branches",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := resolvePaths(".", allFlag)
			var results []ops.BranchesResult
			for _, p := range paths {
				results = append(results, ops.GetBranches(p))
			}
			if jsonOutput {
				return printJSON(results)
			}
			for _, r := range results {
				if r.Error != "" {
					fmt.Printf(fmtStatusLine, fail, r.RepoName, r.Error)
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
				var results []ops.CloneResult
				for _, r := range repos {
					dest := filepath.Join(destPath, r.Name)
					results = append(results, ops.CloneRepository(r, dest))
				}
				if jsonOutput {
					return printJSON(results)
				}
				for _, r := range results {
					printResult(r.RepoName, r.Success, "", r.Error)
				}
				return nil
			}
			if repoURL != "" {
				result := ops.CloneFromURL(repoURL, destPath)
				if jsonOutput {
					return printJSON(result)
				}
				printResult(result.RepoName, result.Success, "", result.Error)
				return nil
			}
			return fmt.Errorf("specify --repo or --org")
		},
	}

	cmd.Flags().StringVarP(&repoURL, "repo", "r", "", "Repo SSH URL")
	cmd.Flags().StringVarP(&orgName, "org", "o", "", descGitHubOrg)
	cmd.Flags().StringVarP(&teamName, "team", "t", "", "Team within org")
	cmd.Flags().StringVarP(&destPath, "path", "p", ".", "Destination")

	return cmd
}

func automergeCLICmd() *cobra.Command {
	var target string
	var allFlag bool

	cmd := &cobra.Command{
		Use:   "automerge",
		Short: "Enable auto-merge on PRs",
		RunE: func(cmd *cobra.Command, args []string) error {
			paths := resolvePaths(".", allFlag)
			var results []ops.AutomergeResult
			for _, p := range paths {
				results = append(results, ops.EnableAutomerge(target, p))
			}
			if jsonOutput {
				return printJSON(results)
			}
			for _, r := range results {
				info := ""
				if r.PRNumber != "" {
					info = "PR #" + r.PRNumber
				}
				printResult(r.RepoName, r.Success, info, r.Error)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "PR search target")
	cmd.Flags().BoolVarP(&allFlag, "all", "a", false, descProcessRepos)
	cmd.MarkFlagRequired("target")

	return cmd
}

type mergeConfig struct {
	org       string
	dryRun    bool
	batchSize int
	waitMin   int
	json      bool
}

type mergeSummary struct {
	Merged int          `json:"merged"`
	Failed int          `json:"failed"`
	PRs    []ops.PRInfo `json:"prs"`
}

func runMergePRs(cfg mergeConfig) error {
	merged, failed := 0, 0
	batch := 0
	var allPRs []ops.PRInfo

	for {
		prs, err := ops.FetchApprovedPRs(cfg.org)
		if err != nil {
			return err
		}
		if len(prs) == 0 {
			if !cfg.json {
				printNoPRsMessage(batch)
			}
			break
		}

		end := min(cfg.batchSize, len(prs))
		batch++
		if !cfg.json {
			fmt.Printf("\nBatch %d — %d approved PR(s), merging %d\n", batch, len(prs), end)
		}

		allPRs = append(allPRs, prs[:end]...)
		m, f := processMergeBatch(cfg, prs[:end])
		merged += m
		failed += f

		if cfg.dryRun {
			break
		}

		if !cfg.json {
			fmt.Printf("\nWaiting %d minutes before next batch…\n", cfg.waitMin)
		}
		time.Sleep(time.Duration(cfg.waitMin) * time.Minute)
	}

	if cfg.json {
		return printJSON(mergeSummary{Merged: merged, Failed: failed, PRs: allPRs})
	}
	if !cfg.dryRun && batch > 0 {
		fmt.Printf("\n=== Summary ===\nMerged: %d\nFailed: %d\n", merged, failed)
	}
	return nil
}

func printNoPRsMessage(batch int) {
	if batch == 0 {
		fmt.Println("No approved PRs found.")
	} else {
		fmt.Println("\nNo more approved PRs.")
	}
}

func processMergeBatch(cfg mergeConfig, prs []ops.PRInfo) (merged, failed int) {
	for _, pr := range prs {
		name := fmt.Sprintf("%s #%d", pr.Repo, pr.Number)
		if cfg.dryRun {
			fmt.Printf(fmtStatusLine, ok, name, dim.Render(pr.Title))
			continue
		}
		result := ops.MergePR(cfg.org, pr.Repo, pr.Number)
		printResult(name, result.Success, pr.Title, result.Error)
		if result.Success {
			merged++
		} else {
			failed++
		}
	}
	return merged, failed
}

func mergePRsCLICmd() *cobra.Command {
	var orgName string
	var dryRun bool
	var batchSize, waitMin int

	cmd := &cobra.Command{
		Use:   "merge-prs",
		Short: "Merge approved pull requests in batches",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMergePRs(mergeConfig{
				org:       orgName,
				dryRun:    dryRun,
				batchSize: batchSize,
				waitMin:   waitMin,
				json:      jsonOutput,
			})
		},
	}

	cmd.Flags().StringVarP(&orgName, "org", "o", "Sundsvallskommun", descGitHubOrg)
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be merged")
	cmd.Flags().IntVar(&batchSize, "batch-size", 5, "PRs per batch")
	cmd.Flags().IntVar(&waitMin, "wait", 10, "Minutes between batches")

	return cmd
}

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

			var results []ops.EnableWorkflowResult
			for _, repo := range repos {
				results = append(results, ops.FindAndEnableWorkflows(orgName, repo, workflowName, prBranch))
			}

			if jsonOutput {
				return printJSON(results)
			}

			fmt.Printf("Checking %d repos in %s\n", len(repos), orgName)
			for i, result := range results {
				var parts []string
				if result.EnabledCount > 0 {
					parts = append(parts, fmt.Sprintf("enabled %d", result.EnabledCount))
				}
				if result.RetriggeredPRs > 0 {
					parts = append(parts, fmt.Sprintf("retriggered %d PR(s)", result.RetriggeredPRs))
				}
				printResult(repos[i], result.Success, strings.Join(parts, ", "), result.Error)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&orgName, "org", "o", "Sundsvallskommun", descGitHubOrg)
	cmd.Flags().StringVarP(&workflowName, "workflow", "w", "Call Java CI with Maven", "Workflow name to enable (empty = all)")
	cmd.Flags().StringVar(&prBranch, "pr-branch", "", "Close/reopen PRs from this head branch to retrigger workflows")

	return cmd
}

func teamPRsCLICmd() *cobra.Command {
	var orgName, teamName string

	cmd := &cobra.Command{
		Use:   "team-prs",
		Short: "List open PRs across team repos",
		RunE: func(cmd *cobra.Command, args []string) error {
			excludePrefixes := []string{"web-app-", "Camunda-"}
			repos, err := ops.FetchTeamRepoNames(orgName, teamName, excludePrefixes)
			if err != nil {
				return err
			}

			prs, err := ops.FetchTeamPRs(orgName, repos)
			if err != nil {
				return err
			}

			if jsonOutput {
				return printJSON(prs)
			}

			fmt.Printf("Found %d repos for %s/%s\n", len(repos), orgName, teamName)
			if len(prs) == 0 {
				fmt.Println("No open PRs.")
				return nil
			}

			fmt.Printf("\n%d open PR(s):\n\n", len(prs))
			for _, pr := range prs {
				fmt.Printf(" %s %-30s #%-5d %-16s %s\n",
					ok, pr.Repo, pr.Number, dim.Render(pr.Author), pr.Title)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&orgName, "org", "o", "Sundsvallskommun", descGitHubOrg)
	cmd.Flags().StringVarP(&teamName, "team", "t", "team-unmasked", "Team slug")

	return cmd
}

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
