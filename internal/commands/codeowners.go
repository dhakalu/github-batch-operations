package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"go-repo-manager/internal/logger"
	"go-repo-manager/internal/repo"
)

const (
	// Formatting constants.
	shortSeparatorLength = 50
	longSeparatorLength  = 70
)

func newCodeownersCmd() *cobra.Command {
	var (
		repoName       string
		repoPrefix     string
		org            string
		username       string
		token          string
		concurrency    int
		codeownersFile string
	)

	cmd := &cobra.Command{
		Use:   "codeowners",
		Short: "Add or update CODEOWNERS file in repositories",
		Long:  "Add or update CODEOWNERS file in specified repositories, repositories with a given prefix, or all repositories in an organization or user account",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCodeownersCommand(repoName, repoPrefix, org, username, token, concurrency, codeownersFile)
		},
	}

	cmd.Flags().StringVar(&repoName, "repo", "", "Specific repository name")
	cmd.Flags().StringVar(&repoPrefix, "repo-prefix", "", "Repository name prefix to filter repositories")
	cmd.Flags().StringVar(&org, "org", "", "GitHub organization name")
	cmd.Flags().StringVar(&username, "username", "", "GitHub username")
	cmd.Flags().StringVar(&token, "token", "", "GitHub personal access token (optional, can also be set via GITHUB_TOKEN env var)")
	cmd.Flags().IntVar(&concurrency, "concurrency", 1, "Maximum number of concurrent workers for processing repositories (default: 1)")
	cmd.Flags().StringVar(&codeownersFile, "codeowner-file", "", "Path to the CODEOWNERS file to add to repositories (required)")

	// Mark the codeowner-file flag as required
	cmd.MarkFlagRequired("codeowner-file")

	return cmd
}

func runCodeownersCommand(repoName, repoPrefix, org, username, token string, concurrency int, codeownersFile string) error {
	log := logger.GetLogger()

	// Validate input parameters
	if err := validateCodeownersFlags(org, username, repoName, repoPrefix, codeownersFile); err != nil {
		return err
	}

	// Read the CODEOWNERS file content
	codeownersContent, err := readCodeownersFile(codeownersFile)
	if err != nil {
		return fmt.Errorf("failed to read CODEOWNERS file: %w", err)
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
		// Add CODEOWNERS to single repository
		return handleSingleRepoCodeowners(ctx, githubService, owner, repoName, codeownersContent)
	} else {
		if repoName == "" && repoPrefix == "" {
			if isUser {
				log.Info("No repository or prefix specified, adding CODEOWNERS to all repositories for user")
			} else {
				log.Info("No repository or prefix specified, adding CODEOWNERS to all repositories in organization")
			}
		}
		return handleMultipleReposCodeowners(ctx, githubService, owner, repoPrefix, isUser, codeownersContent)
	}
}

func validateCodeownersFlags(org, username, repoName, repoPrefix, codeownersFile string) error {
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

	if codeownersFile == "" {
		return fmt.Errorf("codeowners file path (--codeowner-file) is required")
	}

	return nil
}

func readCodeownersFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return string(content), nil
}

func handleSingleRepoCodeowners(ctx context.Context, githubService repo.GitHubClient, owner, repoName, codeownersContent string) error {
	log := logger.GetLogger()

	commitMessage := "Add/Update CODEOWNERS file"
	err := githubService.CreateOrUpdateFile(ctx, owner, repoName, ".github/CODEOWNERS", codeownersContent, commitMessage)
	if err != nil {
		log.Error("Failed to add CODEOWNERS to repository", "owner", owner, "repo", repoName, "error", err)
		return err
	}

	displaySingleRepoCodeownersResult(owner, repoName, true)
	return nil
}

func handleMultipleReposCodeowners(ctx context.Context, githubService repo.GitHubClient, owner, prefix string, isUser bool, codeownersContent string) error {
	log := logger.GetLogger()

	successRepos, failedRepos, err := githubService.AddCodeownersToReposWithPrefix(ctx, owner, prefix, isUser, codeownersContent)
	if err != nil {
		log.Error("Failed to add CODEOWNERS to repositories with prefix", "owner", owner, "prefix", prefix, "error", err)
		return err
	}

	if len(successRepos) == 0 && len(failedRepos) == 0 {
		log.Info("No repositories found matching the specified criteria", "owner", owner, "prefix", prefix)
		return nil
	}

	displayMultipleReposCodeownersResults(owner, prefix, successRepos, failedRepos, isUser)
	return nil
}

func displaySingleRepoCodeownersResult(owner, repoName string, success bool) {
	fmt.Println("\n📋 CODEOWNERS Update Result:")
	fmt.Println(strings.Repeat("-", shortSeparatorLength))

	if success {
		fmt.Printf("✅ Repository: %s/%s (SUCCESS)\n", owner, repoName)
		fmt.Printf("📁 CODEOWNERS file successfully added/updated\n")
		fmt.Printf("📍 Location: .github/CODEOWNERS\n")
	} else {
		fmt.Printf("❌ Repository: %s/%s (FAILED)\n", owner, repoName)
		fmt.Printf("❗ Failed to add/update CODEOWNERS file\n")
	}
	fmt.Println(strings.Repeat("-", shortSeparatorLength))
}

func displayMultipleReposCodeownersResults(owner, prefix string, successRepos, failedRepos []string, isUser bool) {
	// Sort the repositories for consistent output
	sort.Strings(successRepos)
	sort.Strings(failedRepos)

	fmt.Println("\n📋 CODEOWNERS Update Results:")
	fmt.Println(strings.Repeat("-", longSeparatorLength))

	// Display successful repositories
	if len(successRepos) > 0 {
		fmt.Printf("✅ SUCCESSFUL UPDATES (%d repositories):\n", len(successRepos))
		for _, repoName := range successRepos {
			fmt.Printf("  ✅ %s/%s\n", owner, repoName)
		}
		fmt.Println()
	}

	// Display failed repositories
	if len(failedRepos) > 0 {
		fmt.Printf("❌ FAILED UPDATES (%d repositories):\n", len(failedRepos))
		for _, repoName := range failedRepos {
			fmt.Printf("  ❌ %s/%s\n", owner, repoName)
		}
		fmt.Println()
	}

	// Display summary
	ownerType := "organization"
	if isUser {
		ownerType = "user"
	}

	fmt.Println("=" + strings.Repeat("=", longSeparatorLength))
	if prefix == "" {
		fmt.Printf("📊 SUMMARY for all repositories for %s '%s':\n", ownerType, owner)
	} else {
		fmt.Printf("📊 SUMMARY for repositories with prefix '%s' for %s '%s':\n", prefix, ownerType, owner)
	}
	fmt.Println(strings.Repeat("-", longSeparatorLength))
	fmt.Printf("📁 Total Repositories: %d\n", len(successRepos)+len(failedRepos))
	fmt.Printf("✅ Successful Updates: %d\n", len(successRepos))
	fmt.Printf("❌ Failed Updates: %d\n", len(failedRepos))

	if len(successRepos)+len(failedRepos) > 0 {
		successPercentage := float64(len(successRepos)) / float64(len(successRepos)+len(failedRepos)) * 100
		fmt.Printf("📈 Success Rate: %.1f%%\n", successPercentage)
	}

	if len(successRepos) > 0 {
		fmt.Printf("📍 CODEOWNERS files added/updated at: .github/CODEOWNERS\n")
	}

	if len(failedRepos) == 0 {
		fmt.Printf("🎉 All repositories successfully updated!\n")
	}
	fmt.Println("=" + strings.Repeat("=", longSeparatorLength))
}
