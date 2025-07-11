package repo

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"go-repo-manager/internal/logger"

	"github.com/google/go-github/v62/github"
)

// IssueStats represents issue statistics for a repository
type IssueStats struct {
	RepoName     string
	TotalIssues  int
	OpenIssues   int
	ClosedIssues int
}

// GitHubClient defines the interface for GitHub API operations
type GitHubClient interface {
	// GetIssueStatsForRepo retrieves issue statistics for a single repository.
	// It returns the total count of issues, open issues, and closed issues.
	// Pull requests are excluded from the count as they are separate from issues in GitHub's API.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - org: GitHub organization name
	//   - repoName: Name of the repository within the organization
	//
	// Returns:
	//   - *IssueStats: Statistics containing issue counts for the repository
	//   - error: Any error encountered during the API calls
	GetIssueStatsForRepo(ctx context.Context, org, repoName string) (*IssueStats, error)

	// GetRepositoriesWithPrefix retrieves all repositories in an organization that have names
	// starting with the specified prefix. If prefix is empty, it returns all repositories.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - org: GitHub organization name
	//   - prefix: Repository name prefix to filter by (empty string matches all)
	//
	// Returns:
	//   - []*github.Repository: Slice of repositories matching the prefix
	//   - error: Any error encountered during the API calls
	GetRepositoriesWithPrefix(ctx context.Context, org, prefix string) ([]*github.Repository, error)

	// GetIssueStatsForReposWithPrefix retrieves issue statistics for all repositories
	// in an organization that match the specified prefix. This is a convenience method
	// that combines GetRepositoriesWithPrefix and GetIssueStatsForRepo.
	//
	// If any individual repository fails to return stats, it logs the error and continues
	// processing other repositories rather than failing the entire operation.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - org: GitHub organization name
	//   - prefix: Repository name prefix to filter by (empty string matches all)
	//
	// Returns:
	//   - []*IssueStats: Slice of issue statistics for each matching repository
	//   - error: Any error encountered during repository discovery (individual repo errors are logged)
	GetIssueStatsForReposWithPrefix(ctx context.Context, org, prefix string) ([]*IssueStats, error)
}

// gitHubService is the concrete implementation of GitHubClient
type gitHubService struct {
	client         *github.Client
	log            *slog.Logger
	maxConcurrency int
}

// NewGitHubService creates a new GitHub service instance that implements GitHubClient
// The GitHub client is injected as a dependency for better testability and flexibility
func NewGitHubService(client *github.Client) GitHubClient {
	return NewGitHubServiceWithConcurrency(client, 1) // Default concurrency of 1
}

// NewGitHubServiceWithConcurrency creates a new GitHub service instance with configurable concurrency
// The maxConcurrency parameter controls how many repositories can be processed concurrently
func NewGitHubServiceWithConcurrency(client *github.Client, maxConcurrency int) GitHubClient {
	log := logger.GetLogger()
	return NewGitHubServiceWithLogger(client, maxConcurrency, log)
}

// NewGitHubServiceWithLogger creates a new GitHub service instance with configurable concurrency and logger
// This constructor is primarily used for testing to inject a custom logger
func NewGitHubServiceWithLogger(client *github.Client, maxConcurrency int, log *slog.Logger) GitHubClient {
	if maxConcurrency <= 0 {
		log.Warn("Invalid maxConcurrency value, using default", "provided", maxConcurrency, "default", 10)
		maxConcurrency = 10
	}

	return &gitHubService{
		client:         client,
		log:            log,
		maxConcurrency: maxConcurrency,
	}
}

// NewGitHubClient creates a new GitHub client with optional token authentication
// This is a factory function to create the GitHub client that can be injected into the service
func NewGitHubClient(token string) *github.Client {
	log := logger.GetLogger()

	if token != "" {
		return github.NewClient(nil).WithAuthToken(token)
	} else {
		log.Warn("No GitHub token provided. Rate limits will be more restrictive.")
		return github.NewClient(nil)
	}
}

// GetIssueStatsForRepo gets issue statistics for a single repository
func (s *gitHubService) GetIssueStatsForRepo(ctx context.Context, org, repoName string) (*IssueStats, error) {
	s.log.Info("Fetching issue count", "org", org, "repo", repoName)

	// Verify repository exists
	_, _, err := s.client.Repositories.Get(ctx, org, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository %s/%s: %w", org, repoName, err)
	}

	stats := &IssueStats{RepoName: repoName}

	// List issues (excluding pull requests)
	opts := &github.IssueListByRepoOptions{
		State: "all", // Get both open and closed issues
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		issues, resp, err := s.client.Issues.ListByRepo(ctx, org, repoName, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list issues for %s/%s: %w", org, repoName, err)
		}

		for _, issue := range issues {
			// Skip pull requests (issues with PullRequestLinks are PRs)
			if issue.PullRequestLinks == nil {
				stats.TotalIssues++
				if issue.GetState() == "open" {
					stats.OpenIssues++
				} else {
					stats.ClosedIssues++
				}
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return stats, nil
}

// GetRepositoriesWithPrefix gets all repositories in an organization that match a prefix
func (s *gitHubService) GetRepositoriesWithPrefix(ctx context.Context, org, prefix string) ([]*github.Repository, error) {
	s.log.Info("Fetching repositories with prefix", "org", org, "prefix", prefix)

	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var matchingRepos []*github.Repository

	for {
		repos, resp, err := s.client.Repositories.ListByOrg(ctx, org, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories for org %s: %w", org, err)
		}

		for _, repo := range repos {
			if strings.HasPrefix(repo.GetName(), prefix) {
				matchingRepos = append(matchingRepos, repo)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return matchingRepos, nil
}

// GetIssueStatsForReposWithPrefix gets issue statistics for all repositories matching a prefix
func (s *gitHubService) GetIssueStatsForReposWithPrefix(ctx context.Context, org, prefix string) ([]*IssueStats, error) {
	repos, err := s.GetRepositoriesWithPrefix(ctx, org, prefix)
	if err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		s.log.Info("No repositories found with prefix", "prefix", prefix)
		return nil, nil
	}

	s.log.Info("Found repositories with prefix", "count", len(repos), "prefix", prefix)

	var allStats []*IssueStats
	statsChan := make(chan *IssueStats, len(repos))
	errChan := make(chan error, len(repos))
	sem := make(chan struct{}, s.maxConcurrency) // Limit concurrency to maxConcurrency workers

	for _, repo := range repos {
		sem <- struct{}{}
		go func(repoName string) {
			defer func() { <-sem }()
			stats, err := s.GetIssueStatsForRepo(ctx, org, repoName)
			if err != nil {
				errChan <- fmt.Errorf("failed to get issues for repository %s: %w", repoName, err)
				return
			}
			statsChan <- stats
		}(repo.GetName())
	}

	for i := 0; i < len(repos); i++ {
		select {
		case stats := <-statsChan:
			allStats = append(allStats, stats)
		case err := <-errChan:
			s.log.Error("Error fetching repository stats", "error", err)
		}
	}

	return allStats, nil
}
