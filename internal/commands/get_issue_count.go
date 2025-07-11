package commands

import (
	"context"
	"fmt"
	"os"

	"go-repo-manager/internal/logger"
	"go-repo-manager/internal/repo"

	"github.com/spf13/cobra"
)

func newGetIssueCountCmd() *cobra.Command {
	var (
		repoName    string
		repoPrefix  string
		org         string
		token       string
		concurrency int
	)

	cmd := &cobra.Command{
		Use:   "get-issue-count",
		Short: "Get issue count from repositories",
		Long:  "Get the count of issues from specified repositories, repositories with a given prefix, or all repositories in an organization",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.GetLogger()

			// Validate input
			if org == "" {
				return fmt.Errorf("organization (--org) is required")
			}

			if repoName != "" && repoPrefix != "" {
				return fmt.Errorf("cannot specify both --repo and --repo-prefix")
			}

			// Get token from environment if not provided via flag
			if token == "" {
				token = os.Getenv("GITHUB_TOKEN")
			}

			// Create GitHub client and service with dependency injection
			githubClient := repo.NewGitHubClient(token)
			githubService := repo.NewGitHubServiceWithConcurrency(githubClient, concurrency)
			ctx := context.Background()

			if repoName != "" {
				// Get issue count for single repository
				return handleSingleRepo(ctx, githubService, org, repoName)
			} else {
				// Get issue count for repositories with prefix
				// If no prefix provided, use empty string to get all repositories
				prefix := repoPrefix
				if repoName == "" && repoPrefix == "" {
					log.Info("No repository or prefix specified, fetching all repositories in organization")
				}
				return handleMultipleRepos(ctx, githubService, org, prefix)
			}
		},
	}

	cmd.Flags().StringVar(&repoName, "repo", "", "Specific repository name")
	cmd.Flags().StringVar(&repoPrefix, "repo-prefix", "", "Repository name prefix to filter repositories")
	cmd.Flags().StringVar(&org, "org", "", "GitHub organization name (required)")
	cmd.Flags().StringVar(&token, "token", "", "GitHub personal access token (optional, can also be set via GITHUB_TOKEN env var)")
	cmd.Flags().IntVar(&concurrency, "concurrency", 1, "Maximum number of concurrent workers for processing repositories (default: 1)")

	return cmd
}

func handleSingleRepo(ctx context.Context, githubService repo.GitHubClient, org, repoName string) error {
	stats, err := githubService.GetIssueStatsForRepo(ctx, org, repoName)
	if err != nil {
		logger.GetLogger().Error("Failed to get issue stats for repository", "org", org, "repo", repoName, "error", err)
		return err
	}

	displaySingleRepoStats(org, stats)
	return nil
}

func handleMultipleRepos(ctx context.Context, githubService repo.GitHubClient, org, prefix string) error {
	log := logger.GetLogger()
	allStats, err := githubService.GetIssueStatsForReposWithPrefix(ctx, org, prefix)
	if err != nil {
		log.Error("Failed to get issue stats for repositories with prefix", "org", org, "prefix", prefix, "error", err)
		return err
	}

	if len(allStats) == 0 {
		log.Info("No repositories found matching the specified criteria", "org", org, "prefix", prefix)
		return nil
	}

	displayMultipleReposStats(org, prefix, allStats)
	return nil
}

func displaySingleRepoStats(org string, stats *repo.IssueStats) {
	fmt.Printf("Repository: %s/%s\n", org, stats.RepoName)
	fmt.Printf("  Total Issues: %d\n", stats.TotalIssues)
	fmt.Printf("  Open Issues: %d\n", stats.OpenIssues)
	fmt.Printf("  Closed Issues: %d\n", stats.ClosedIssues)
	fmt.Println()
}

func displayMultipleReposStats(org, prefix string, allStats []*repo.IssueStats) {
	var totalIssuesAcrossRepos int
	var totalOpenIssues int
	var totalClosedIssues int

	// Display individual repository stats
	for _, stats := range allStats {
		fmt.Printf("Repository: %s/%s\n", org, stats.RepoName)
		fmt.Printf("  Total Issues: %d\n", stats.TotalIssues)
		fmt.Printf("  Open Issues: %d\n", stats.OpenIssues)
		fmt.Printf("  Closed Issues: %d\n", stats.ClosedIssues)
		fmt.Println()

		totalIssuesAcrossRepos += stats.TotalIssues
		totalOpenIssues += stats.OpenIssues
		totalClosedIssues += stats.ClosedIssues
	}

	// Display summary
	fmt.Printf("Summary for repositories with prefix '%s':\n", prefix)
	fmt.Printf("  Repositories found: %d\n", len(allStats))
	fmt.Printf("  Total Issues across all repos: %d\n", totalIssuesAcrossRepos)
	fmt.Printf("  Total Open Issues: %d\n", totalOpenIssues)
	fmt.Printf("  Total Closed Issues: %d\n", totalClosedIssues)
}
