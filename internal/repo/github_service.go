package repo

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/go-github/v62/github"

	"go-repo-manager/internal/logger"
)

const (
	// Default concurrency value.
	defaultConcurrency = 10
)

// IssueStats represents issue statistics for a repository.
type IssueStats struct {
	RepoName     string
	TotalIssues  int
	OpenIssues   int
	ClosedIssues int
}

// GitHubClient defines the interface for GitHub API operations.
type GitHubClient interface {
	// GetIssueStatsForRepo retrieves issue statistics for a single repository.
	// It returns the total count of issues, open issues, and closed issues.
	// Pull requests are excluded from the count as they are separate from issues in GitHub's API.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - owner: GitHub organization or username
	//   - repoName: Name of the repository within the organization or user account
	//
	// Returns:
	//   - *IssueStats: Statistics containing issue counts for the repository
	//   - error: Any error encountered during the API calls
	GetIssueStatsForRepo(ctx context.Context, owner, repoName string) (*IssueStats, error)

	// GetRepositoriesWithPrefix retrieves all repositories for an owner (organization or user) that have names
	// starting with the specified prefix. If prefix is empty, it returns all repositories.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - owner: GitHub organization or username
	//   - prefix: Repository name prefix to filter by (empty string matches all)
	//   - isUser: true if owner is a user, false if it's an organization
	//
	// Returns:
	//   - []*github.Repository: Slice of repositories matching the prefix
	//   - error: Any error encountered during the API calls
	GetRepositoriesWithPrefix(ctx context.Context, owner, prefix string, isUser bool) ([]*github.Repository, error)

	// GetIssueStatsForReposWithPrefix retrieves issue statistics for all repositories
	// for an owner (organization or user) that match the specified prefix. This is a convenience method
	// that combines GetRepositoriesWithPrefix and GetIssueStatsForRepo.
	//
	// If any individual repository fails to return stats, it logs the error and continues
	// processing other repositories rather than failing the entire operation.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - owner: GitHub organization or username
	//   - prefix: Repository name prefix to filter by (empty string matches all)
	//   - isUser: true if owner is a user, false if it's an organization
	//
	// Returns:
	//   - []*IssueStats: Slice of issue statistics for each matching repository
	//   - error: Any error encountered during repository discovery (individual repo errors are logged)
	GetIssueStatsForReposWithPrefix(ctx context.Context, owner, prefix string, isUser bool) ([]*IssueStats, error)

	// CreateOrUpdateFile creates or updates a file in a repository
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - owner: GitHub organization or username
	//   - repoName: Name of the repository
	//   - filePath: Path to the file within the repository (e.g., ".github/CODEOWNERS")
	//   - content: Content of the file
	//   - commitMessage: Commit message for the file change
	//
	// Returns:
	//   - error: Any error encountered during the file creation/update
	CreateOrUpdateFile(ctx context.Context, owner, repoName, filePath, content, commitMessage string) error

	// AddCodeownersToReposWithPrefix adds a CODEOWNERS file to all repositories
	// for an owner (organization or user) that match the specified prefix.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - owner: GitHub organization or username
	//   - prefix: Repository name prefix to filter by (empty string matches all)
	//   - isUser: true if owner is a user, false if it's an organization
	//   - codeownersContent: Content of the CODEOWNERS file
	//
	// Returns:
	//   - []string: Slice of repository names that were successfully updated
	//   - []string: Slice of repository names that failed to update
	//   - error: Any error encountered during repository discovery
	AddCodeownersToReposWithPrefix(ctx context.Context, owner, prefix string, isUser bool,
		codeownersContent string) ([]string, []string, error)
}

// gitHubService is the concrete implementation of GitHubClient.
type gitHubService struct {
	client         *github.Client
	log            *slog.Logger
	maxConcurrency int
}

// The GitHub client is injected as a dependency for better testability and flexibility.
func NewGitHubService(client *github.Client) GitHubClient {
	return NewGitHubServiceWithConcurrency(client, 1) // Default concurrency of 1
}

// The maxConcurrency parameter controls how many repositories can be processed concurrently.
func NewGitHubServiceWithConcurrency(client *github.Client, maxConcurrency int) GitHubClient {
	log := logger.GetLogger()

	return NewGitHubServiceWithLogger(client, maxConcurrency, log)
}

// This constructor is primarily used for testing to inject a custom logger.
func NewGitHubServiceWithLogger(client *github.Client, maxConcurrency int, log *slog.Logger) GitHubClient {
	if maxConcurrency <= 0 {
		log.Warn("Invalid maxConcurrency value, using default", "provided", maxConcurrency, "default", defaultConcurrency)
		maxConcurrency = defaultConcurrency
	}

	return &gitHubService{
		client:         client,
		log:            log,
		maxConcurrency: maxConcurrency,
	}
}

// This is a factory function to create the GitHub client that can be injected into the service.
func NewGitHubClient(token string) *github.Client {
	log := logger.GetLogger()

	if token != "" {
		return github.NewClient(nil).WithAuthToken(token)
	} else {
		log.Warn("No GitHub token provided. Rate limits will be more restrictive.")

		return github.NewClient(nil)
	}
}

// GetIssueStatsForRepo gets issue statistics for a single repository.
func (s *gitHubService) GetIssueStatsForRepo(ctx context.Context, owner, repoName string) (*IssueStats, error) {
	s.log.Info("Fetching issue count", "owner", owner, "repo", repoName)

	// Verify repository exists
	_, _, err := s.client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository %s/%s: %w", owner, repoName, err)
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
		issues, resp, err := s.client.Issues.ListByRepo(ctx, owner, repoName, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list issues for %s/%s: %w", owner, repoName, err)
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

// GetRepositoriesWithPrefix gets all repositories for an owner that match a prefix.
func (s *gitHubService) GetRepositoriesWithPrefix(ctx context.Context, owner, prefix string, isUser bool) ([]*github.Repository, error) {
	s.log.Info("Fetching repositories with prefix", "owner", owner, "prefix", prefix, "isUser", isUser)

	if isUser {
		return s.getUserRepositoriesWithPrefix(ctx, owner, prefix)
	}

	return s.getOrgRepositoriesWithPrefix(ctx, owner, prefix)
}

func (s *gitHubService) getUserRepositoriesWithPrefix(ctx context.Context, owner, prefix string) ([]*github.Repository, error) {
	var matchingRepos []*github.Repository

	opts := &github.RepositoryListByUserOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		repos, resp, err := s.client.Repositories.ListByUser(ctx, owner, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories for user %s: %w", owner, err)
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

func (s *gitHubService) getOrgRepositoriesWithPrefix(ctx context.Context, owner, prefix string) ([]*github.Repository, error) {
	var matchingRepos []*github.Repository

	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		repos, resp, err := s.client.Repositories.ListByOrg(ctx, owner, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories for org %s: %w", owner, err)
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

// GetIssueStatsForReposWithPrefix gets issue statistics for all repositories matching a prefix.
func (s *gitHubService) GetIssueStatsForReposWithPrefix(ctx context.Context, owner, prefix string, isUser bool) ([]*IssueStats, error) {
	repos, err := s.GetRepositoriesWithPrefix(ctx, owner, prefix, isUser)
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

			stats, err := s.GetIssueStatsForRepo(ctx, owner, repoName)
			if err != nil {
				errChan <- fmt.Errorf("failed to get issues for repository %s: %w", repoName, err)

				return
			}
			statsChan <- stats
		}(repo.GetName())
	}

	for range repos {
		select {
		case stats := <-statsChan:
			allStats = append(allStats, stats)
		case err := <-errChan:
			s.log.Error("Error fetching repository stats", "error", err)
		}
	}

	return allStats, nil
}

// CreateOrUpdateFile creates or updates a file in a repository.
func (s *gitHubService) CreateOrUpdateFile(ctx context.Context, owner, repoName, filePath, content, commitMessage string) error {
	s.log.Info("Creating or updating file", "owner", owner, "repo", repoName, "file", filePath)

	// Get the current file to check if it exists and get its SHA
	fileContent, _, resp, err := s.client.Repositories.GetContents(ctx, owner, repoName, filePath, nil)

	var sha *string

	if err != nil {
		// Check if error is 404 (file doesn't exist)
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			// File doesn't exist, we'll create it (sha remains nil)
			s.log.Info("File doesn't exist, will create new file", "file", filePath)
		} else {
			return fmt.Errorf("failed to check if file exists %s/%s:%s: %w", owner, repoName, filePath, err)
		}
	} else {
		// File exists, get its SHA for updating
		sha = fileContent.SHA
		s.log.Info("File exists, will update existing file", "file", filePath, "sha", *sha)
	}

	// Create the file update options
	opts := &github.RepositoryContentFileOptions{
		Message: github.String(commitMessage),
		Content: []byte(content),
		SHA:     sha, // nil for new files, existing SHA for updates
	}

	// Create or update the file
	_, _, err = s.client.Repositories.CreateFile(ctx, owner, repoName, filePath, opts)
	if err != nil {
		return fmt.Errorf("failed to create/update file %s/%s:%s: %w", owner, repoName, filePath, err)
	}

	s.log.Info("Successfully created/updated file", "owner", owner, "repo", repoName, "file", filePath)

	return nil
}

// AddCodeownersToReposWithPrefix adds a CODEOWNERS file to all repositories matching a prefix.
func (s *gitHubService) AddCodeownersToReposWithPrefix(ctx context.Context, owner, prefix string,
	isUser bool, codeownersContent string,
) ([]string, []string, error) {
	repos, err := s.GetRepositoriesWithPrefix(ctx, owner, prefix, isUser)
	if err != nil {
		return nil, nil, err
	}

	if len(repos) == 0 {
		s.log.Info("No repositories found with prefix", "prefix", prefix)

		return nil, nil, nil
	}

	s.log.Info("Found repositories with prefix", "count", len(repos), "prefix", prefix)

	var successRepos []string

	var failedRepos []string

	successChan := make(chan string, len(repos))
	failChan := make(chan string, len(repos))
	sem := make(chan struct{}, s.maxConcurrency) // Limit concurrency

	for _, repo := range repos {
		sem <- struct{}{}

		go func(repoName string) {
			defer func() { <-sem }()

			commitMessage := "Add/Update CODEOWNERS file"

			err := s.CreateOrUpdateFile(ctx, owner, repoName, ".github/CODEOWNERS", codeownersContent, commitMessage)
			if err != nil {
				s.log.Error("Failed to add CODEOWNERS to repository", "owner", owner, "repo", repoName, "error", err)
				failChan <- repoName

				return
			}
			successChan <- repoName
		}(repo.GetName())
	}

	// Collect results
	for range repos {
		select {
		case repoName := <-successChan:
			successRepos = append(successRepos, repoName)
		case repoName := <-failChan:
			failedRepos = append(failedRepos, repoName)
		}
	}

	return successRepos, failedRepos, nil
}
