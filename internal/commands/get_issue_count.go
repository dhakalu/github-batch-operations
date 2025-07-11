package commands

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"go-repo-manager/internal/logger"
	"go-repo-manager/internal/repo"

	"github.com/spf13/cobra"
)

func newGetIssueCountCmd() *cobra.Command {
	var (
		repoName    string
		repoPrefix  string
		org         string
		username    string
		token       string
		concurrency int
	)

	cmd := &cobra.Command{
		Use:   "get-issue-count",
		Short: "Get issue count from repositories",
		Long:  "Get the count of issues from specified repositories, repositories with a given prefix, or all repositories in an organization or user account",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.GetLogger()

			// Validate input - must have either org or username, but not both
			if org == "" && username == "" {
				return fmt.Errorf("either organization (--org) or username (--username) is required")
			}

			if org != "" && username != "" {
				return fmt.Errorf("cannot specify both --org and --username")
			}

			if repoName != "" && repoPrefix != "" {
				return fmt.Errorf("cannot specify both --repo and --repo-prefix")
			}

			// Get token from environment if not provided via flag
			if token == "" {
				token = os.Getenv("GITHUB_TOKEN")
			}

			if token == "" {
				return fmt.Errorf("GitHub token (--token) is required or must be set in GITHUB_TOKEN environment variable")
			}

			// Determine if we're working with a user or organization
			var owner string
			var isUser bool
			if username != "" {
				owner = username
				isUser = true
			} else {
				owner = org
				isUser = false
			}

			// Create GitHub client and service with dependency injection
			githubClient := repo.NewGitHubClient(token)
			githubService := repo.NewGitHubServiceWithConcurrency(githubClient, concurrency)
			ctx := context.Background()

			if repoName != "" {
				// Get issue count for single repository
				return handleSingleRepo(ctx, githubService, owner, repoName)
			} else {
				if repoName == "" && repoPrefix == "" {
					if isUser {
						log.Info("No repository or prefix specified, fetching all repositories for user")
					} else {
						log.Info("No repository or prefix specified, fetching all repositories in organization")
					}
				}
				return handleMultipleRepos(ctx, githubService, owner, repoPrefix, isUser)
			}
		},
	}

	cmd.Flags().StringVar(&repoName, "repo", "", "Specific repository name")
	cmd.Flags().StringVar(&repoPrefix, "repo-prefix", "", "Repository name prefix to filter repositories")
	cmd.Flags().StringVar(&org, "org", "", "GitHub organization name")
	cmd.Flags().StringVar(&username, "username", "", "GitHub username")
	cmd.Flags().StringVar(&token, "token", "", "GitHub personal access token (optional, can also be set via GITHUB_TOKEN env var)")
	cmd.Flags().IntVar(&concurrency, "concurrency", 1, "Maximum number of concurrent workers for processing repositories (default: 1)")

	return cmd
}

func handleSingleRepo(ctx context.Context, githubService repo.GitHubClient, owner, repoName string) error {
	stats, err := githubService.GetIssueStatsForRepo(ctx, owner, repoName)
	if err != nil {
		logger.GetLogger().Error("Failed to get issue stats for repository", "owner", owner, "repo", repoName, "error", err)
		return err
	}

	displaySingleRepoStats(owner, stats)
	return nil
}

func handleMultipleRepos(ctx context.Context, githubService repo.GitHubClient, owner, prefix string, isUser bool) error {
	log := logger.GetLogger()
	allStats, err := githubService.GetIssueStatsForReposWithPrefix(ctx, owner, prefix, isUser)
	if err != nil {
		log.Error("Failed to get issue stats for repositories with prefix", "owner", owner, "prefix", prefix, "error", err)
		return err
	}

	if len(allStats) == 0 {
		log.Info("No repositories found matching the specified criteria", "owner", owner, "prefix", prefix)
		return nil
	}

	displayMultipleReposStats(owner, prefix, allStats, isUser)
	return nil
}

func displaySingleRepoStats(owner string, stats *repo.IssueStats) {
	fmt.Println("\n📋 Repository Analysis:")
	fmt.Println(strings.Repeat("-", 50))

	// Determine status indicator based on whether repo has issues
	statusIcon := "❌" // Cross for repos with issues
	statusText := "HAS ISSUES"
	if stats.TotalIssues == 0 {
		statusIcon = "✅" // Checkmark for repos without issues
		statusText = "CLEAN"
	}

	fmt.Printf("%s Repository: %s/%s (%s)\n", statusIcon, owner, stats.RepoName, statusText)
	fmt.Printf("📊 Total Issues: %d\n", stats.TotalIssues)
	if stats.TotalIssues > 0 {
		fmt.Printf("🔓 Open Issues: %d\n", stats.OpenIssues)
		fmt.Printf("✔️  Closed Issues: %d\n", stats.ClosedIssues)
	} else {
		fmt.Printf("🎉 This repository has no issues!\n")
	}
	fmt.Println(strings.Repeat("-", 50))
}

func displayMultipleReposStats(owner, prefix string, allStats []*repo.IssueStats, isUser bool) {
	var totalIssuesAcrossRepos int
	var totalOpenIssues int
	var totalClosedIssues int
	var reposWithIssues int
	var reposWithoutIssues int

	// Sort repositories: repos without issues first, then repos with issues (sorted by total issues desc)
	sort.Slice(allStats, func(i, j int) bool {
		// If one has issues and the other doesn't, prioritize the one without issues
		if (allStats[i].TotalIssues == 0) != (allStats[j].TotalIssues == 0) {
			return allStats[i].TotalIssues == 0
		}
		// If both have issues or both don't have issues, sort by total issues (descending)
		return allStats[i].TotalIssues > allStats[j].TotalIssues
	})

	// Display individual repository stats
	fmt.Println("\n📋 Repository Analysis:")
	fmt.Println(strings.Repeat("-", 70))

	for _, stats := range allStats {
		// Determine status indicator based on whether repo has issues
		statusIcon := "❌" // Cross for repos with issues
		statusText := "HAS ISSUES"
		if stats.TotalIssues == 0 {
			statusIcon = "✅" // Checkmark for repos without issues
			statusText = "CLEAN"
			reposWithoutIssues++
		} else {
			reposWithIssues++
		}

		fmt.Printf("%s Repository: %s/%s (%s)\n", statusIcon, owner, stats.RepoName, statusText)
		fmt.Printf("  📊 Total Issues: %d\n", stats.TotalIssues)
		if stats.TotalIssues > 0 {
			fmt.Printf("  🔓 Open Issues: %d\n", stats.OpenIssues)
			fmt.Printf("  ✔️  Closed Issues: %d\n", stats.ClosedIssues)
		}
		fmt.Println()

		totalIssuesAcrossRepos += stats.TotalIssues
		totalOpenIssues += stats.OpenIssues
		totalClosedIssues += stats.ClosedIssues
	}

	// Display summary
	ownerType := "organization"
	if isUser {
		ownerType = "user"
	}

	fmt.Println("=" + strings.Repeat("=", 70))
	if prefix == "" {
		fmt.Printf("📊 SUMMARY for all repositories for %s '%s':\n", ownerType, owner)
	} else {
		fmt.Printf("📊 SUMMARY for repositories with prefix '%s' for %s '%s':\n", prefix, ownerType, owner)
	}
	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("📁 Total Repositories: %d\n", len(allStats))
	fmt.Printf("✅ Clean Repositories (no issues): %d\n", reposWithoutIssues)
	fmt.Printf("❌ Repositories with issues: %d\n", reposWithIssues)

	if totalIssuesAcrossRepos > 0 {
		fmt.Println(strings.Repeat("-", 70))
		fmt.Printf("🐛 Total Issues across all repos: %d\n", totalIssuesAcrossRepos)
		fmt.Printf("🔓 Total Open Issues: %d\n", totalOpenIssues)
		fmt.Printf("✔️  Total Closed Issues: %d\n", totalClosedIssues)

		// Calculate percentages
		cleanPercentage := float64(reposWithoutIssues) / float64(len(allStats)) * 100
		fmt.Printf("📈 Clean Repository Rate: %.1f%%\n", cleanPercentage)
	} else {
		fmt.Printf("🎉 Congratulations! All repositories are clean (no issues)!\n")
	}
	fmt.Println("=" + strings.Repeat("=", 70))
}
