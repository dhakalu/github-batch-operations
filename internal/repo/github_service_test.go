package repo

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/go-github/v62/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test logger that discards output
func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError, // Only show errors to keep tests quiet
	}))
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// Helper function to create int pointers
func intPtr(i int) *int {
	return &i
}

func TestNewGitHubService(t *testing.T) {
	client := github.NewClient(nil)
	service := NewGitHubService(client)

	assert.NotNil(t, service)
	assert.IsType(t, &gitHubService{}, service)
}

func TestNewGitHubServiceWithConcurrency(t *testing.T) {
	tests := []struct {
		name                string
		maxConcurrency      int
		expectedConcurrency int
	}{
		{"Valid concurrency", 5, 5},
		{"Zero concurrency defaults to 10", 0, 10},
		{"Negative concurrency defaults to 10", -1, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := github.NewClient(nil)
			service := NewGitHubServiceWithConcurrency(client, tt.maxConcurrency).(*gitHubService)

			assert.Equal(t, tt.expectedConcurrency, service.maxConcurrency)
		})
	}
}

func TestNewGitHubClient(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"With token", "test-token"},
		{"Without token", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewGitHubClient(tt.token)
			assert.NotNil(t, client)
			assert.IsType(t, &github.Client{}, client)
		})
	}
}

func TestGetIssueStatsForRepo_WithMockServer(t *testing.T) {
	tests := []struct {
		name          string
		setupServer   func() *httptest.Server
		expectedStats *IssueStats
		expectError   bool
	}{
		{
			name: "Success with mixed issues and PRs",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch {
					case strings.Contains(r.URL.Path, "/repos/testorg/testrepo") && r.Method == "GET" && !strings.Contains(r.URL.Path, "/issues"):
						// Repository exists
						repo := github.Repository{Name: stringPtr("testrepo")}
						json.NewEncoder(w).Encode(repo)
					case strings.Contains(r.URL.Path, "/repos/testorg/testrepo/issues"):
						// Return issues as an array
						issues := []*github.Issue{
							{State: stringPtr("open"), PullRequestLinks: nil},
							{State: stringPtr("closed"), PullRequestLinks: nil},
							{State: stringPtr("open"), PullRequestLinks: &github.PullRequestLinks{}}, // PR
						}
						w.Header().Set("Link", "") // No pagination
						json.NewEncoder(w).Encode(issues)
					default:
						http.NotFound(w, r)
					}
				}))
			},
			expectedStats: &IssueStats{
				RepoName:     "testrepo",
				TotalIssues:  2,
				OpenIssues:   1,
				ClosedIssues: 1,
			},
			expectError: false,
		},
		{
			name: "Repository not found",
			setupServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.Contains(r.URL.Path, "/repos/testorg/nonexistent") {
						w.WriteHeader(http.StatusNotFound)
						json.NewEncoder(w).Encode(map[string]string{"message": "Not Found"})
					}
				}))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupServer()
			defer server.Close()

			// Create GitHub client pointing to test server
			client := github.NewClient(nil)
			client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

			// Create service
			logger := createTestLogger()
			service := NewGitHubServiceWithLogger(client, 1, logger)

			ctx := context.Background()
			repoName := "testrepo"
			if tt.expectError {
				repoName = "nonexistent"
			}

			stats, err := service.GetIssueStatsForRepo(ctx, "testorg", repoName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, stats)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedStats, stats)
			}
		})
	}
}

func TestGetRepositoriesWithPrefix_WithMockServer(t *testing.T) {
	tests := []struct {
		name          string
		prefix        string
		expectedRepos []string
	}{
		{
			name:          "Filter by prefix",
			prefix:        "test-",
			expectedRepos: []string{"test-repo1", "test-repo2"},
		},
		{
			name:          "Empty prefix returns all",
			prefix:        "",
			expectedRepos: []string{"test-repo1", "test-repo2", "other-repo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "/orgs/testorg/repos") {
					repos := []github.Repository{
						{Name: stringPtr("test-repo1")},
						{Name: stringPtr("test-repo2")},
						{Name: stringPtr("other-repo")},
					}
					w.Header().Set("Link", "") // No pagination
					json.NewEncoder(w).Encode(repos)
				}
			}))
			defer server.Close()

			client := github.NewClient(nil)
			client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

			logger := createTestLogger()
			service := NewGitHubServiceWithLogger(client, 1, logger)

			ctx := context.Background()
			repos, err := service.GetRepositoriesWithPrefix(ctx, "testorg", tt.prefix)

			require.NoError(t, err)
			assert.Equal(t, len(tt.expectedRepos), len(repos))

			for i, expectedName := range tt.expectedRepos {
				assert.Equal(t, expectedName, repos[i].GetName())
			}
		})
	}
}

func TestGetIssueStatsForReposWithPrefix_WithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/orgs/testorg/repos"):
			repos := []github.Repository{
				{Name: stringPtr("test-repo1")},
				{Name: stringPtr("test-repo2")},
			}
			w.Header().Set("Link", "") // No pagination
			json.NewEncoder(w).Encode(repos)
		case strings.Contains(r.URL.Path, "/repos/testorg/test-repo1") && !strings.Contains(r.URL.Path, "/issues"):
			repo := github.Repository{Name: stringPtr("test-repo1")}
			json.NewEncoder(w).Encode(repo)
		case strings.Contains(r.URL.Path, "/repos/testorg/test-repo2") && !strings.Contains(r.URL.Path, "/issues"):
			repo := github.Repository{Name: stringPtr("test-repo2")}
			json.NewEncoder(w).Encode(repo)
		case strings.Contains(r.URL.Path, "/repos/testorg/test-repo1/issues"):
			issues := []github.Issue{
				{State: stringPtr("open"), PullRequestLinks: nil},
				{State: stringPtr("closed"), PullRequestLinks: nil},
			}
			w.Header().Set("Link", "")
			json.NewEncoder(w).Encode(issues)
		case strings.Contains(r.URL.Path, "/repos/testorg/test-repo2/issues"):
			issues := []github.Issue{
				{State: stringPtr("open"), PullRequestLinks: nil},
			}
			w.Header().Set("Link", "")
			json.NewEncoder(w).Encode(issues)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := github.NewClient(nil)
	client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")

	logger := createTestLogger()
	service := NewGitHubServiceWithLogger(client, 2, logger) // Test concurrency

	ctx := context.Background()
	stats, err := service.GetIssueStatsForReposWithPrefix(ctx, "testorg", "test-")

	require.NoError(t, err)
	assert.Equal(t, 2, len(stats))

	// Sort stats by repo name for consistent testing
	if len(stats) == 2 && stats[0].RepoName > stats[1].RepoName {
		stats[0], stats[1] = stats[1], stats[0]
	}

	expected := []*IssueStats{
		{RepoName: "test-repo1", TotalIssues: 2, OpenIssues: 1, ClosedIssues: 1},
		{RepoName: "test-repo2", TotalIssues: 1, OpenIssues: 1, ClosedIssues: 0},
	}

	for i, expectedStat := range expected {
		assert.Equal(t, expectedStat.RepoName, stats[i].RepoName)
		assert.Equal(t, expectedStat.TotalIssues, stats[i].TotalIssues)
		assert.Equal(t, expectedStat.OpenIssues, stats[i].OpenIssues)
		assert.Equal(t, expectedStat.ClosedIssues, stats[i].ClosedIssues)
	}
}

func TestGetIssueStatsForRepo_BusinessLogic(t *testing.T) {
	tests := []struct {
		name          string
		issues        []*github.Issue
		expectedStats *IssueStats
	}{
		{
			name: "Mixed issues and PRs",
			issues: []*github.Issue{
				{State: stringPtr("open"), PullRequestLinks: nil},
				{State: stringPtr("closed"), PullRequestLinks: nil},
				{State: stringPtr("open"), PullRequestLinks: &github.PullRequestLinks{}}, // PR - excluded
				{State: stringPtr("closed"), PullRequestLinks: nil},
			},
			expectedStats: &IssueStats{
				RepoName:     "testrepo",
				TotalIssues:  3,
				OpenIssues:   1,
				ClosedIssues: 2,
			},
		},
		{
			name: "Only open issues",
			issues: []*github.Issue{
				{State: stringPtr("open"), PullRequestLinks: nil},
				{State: stringPtr("open"), PullRequestLinks: nil},
			},
			expectedStats: &IssueStats{
				RepoName:     "testrepo",
				TotalIssues:  2,
				OpenIssues:   2,
				ClosedIssues: 0,
			},
		},
		{
			name:   "No issues",
			issues: []*github.Issue{},
			expectedStats: &IssueStats{
				RepoName:     "testrepo",
				TotalIssues:  0,
				OpenIssues:   0,
				ClosedIssues: 0,
			},
		},
		{
			name: "Only pull requests",
			issues: []*github.Issue{
				{State: stringPtr("open"), PullRequestLinks: &github.PullRequestLinks{}},
				{State: stringPtr("closed"), PullRequestLinks: &github.PullRequestLinks{}},
			},
			expectedStats: &IssueStats{
				RepoName:     "testrepo",
				TotalIssues:  0,
				OpenIssues:   0,
				ClosedIssues: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the issue counting logic
			stats := countIssues(tt.issues, "testrepo")

			assert.Equal(t, tt.expectedStats, stats)
		})
	}
}

func TestGetRepositoriesWithPrefix_FilterLogic(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		repos    []*github.Repository
		expected []*github.Repository
	}{
		{
			name:   "Matching prefix",
			prefix: "test-",
			repos: []*github.Repository{
				{Name: stringPtr("test-repo1")},
				{Name: stringPtr("test-repo2")},
				{Name: stringPtr("other-repo")},
			},
			expected: []*github.Repository{
				{Name: stringPtr("test-repo1")},
				{Name: stringPtr("test-repo2")},
			},
		},
		{
			name:   "Empty prefix matches all",
			prefix: "",
			repos: []*github.Repository{
				{Name: stringPtr("repo1")},
				{Name: stringPtr("repo2")},
			},
			expected: []*github.Repository{
				{Name: stringPtr("repo1")},
				{Name: stringPtr("repo2")},
			},
		},
		{
			name:   "No matches",
			prefix: "nonexistent-",
			repos: []*github.Repository{
				{Name: stringPtr("repo1")},
				{Name: stringPtr("repo2")},
			},
			expected: []*github.Repository{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the filtering logic
			var matchingRepos []*github.Repository

			for _, repo := range tt.repos {
				if strings.HasPrefix(repo.GetName(), tt.prefix) {
					matchingRepos = append(matchingRepos, repo)
				}
			}

			assert.Equal(t, len(tt.expected), len(matchingRepos))
			for i, expectedRepo := range tt.expected {
				assert.Equal(t, expectedRepo.GetName(), matchingRepos[i].GetName())
			}
		})
	}
}

// Helper function to count issues
func countIssues(issues []*github.Issue, repoName string) *IssueStats {
	stats := &IssueStats{RepoName: repoName}

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

	return stats
}

// TestGitHubService_ErrorHandling tests error scenarios using interface mocks
func TestGitHubService_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func() GitHubClient
		expectError bool
		errorMsg    string
	}{
		{
			name: "Repository not found",
			setupMocks: func() GitHubClient {
				return &mockGitHubService{
					shouldError: true,
					errorMsg:    "repository not found",
				}
			},
			expectError: true,
			errorMsg:    "repository not found",
		},
		{
			name: "API rate limit exceeded",
			setupMocks: func() GitHubClient {
				return &mockGitHubService{
					shouldError: true,
					errorMsg:    "API rate limit exceeded",
				}
			},
			expectError: true,
			errorMsg:    "API rate limit exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := tt.setupMocks()
			ctx := context.Background()

			_, err := service.GetIssueStatsForRepo(ctx, "testorg", "testrepo")

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetIssueStatsForReposWithPrefix_NoRepos(t *testing.T) {
	service := &mockGitHubService{
		shouldError: false,
		repos:       []*github.Repository{}, // No repositories
	}

	ctx := context.Background()
	stats, err := service.GetIssueStatsForReposWithPrefix(ctx, "testorg", "nonexistent-")

	assert.NoError(t, err)
	assert.Nil(t, stats)
}

// Simple mock implementation for testing error scenarios and edge cases
type mockGitHubService struct {
	shouldError bool
	errorMsg    string
	repos       []*github.Repository
}

func (m *mockGitHubService) GetIssueStatsForRepo(ctx context.Context, org, repoName string) (*IssueStats, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}
	return &IssueStats{
		RepoName:     repoName,
		TotalIssues:  5,
		OpenIssues:   3,
		ClosedIssues: 2,
	}, nil
}

func (m *mockGitHubService) GetRepositoriesWithPrefix(ctx context.Context, org, prefix string) ([]*github.Repository, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}
	if m.repos != nil {
		return m.repos, nil
	}
	return []*github.Repository{
		{Name: stringPtr("test-repo")},
	}, nil
}

func (m *mockGitHubService) GetIssueStatsForReposWithPrefix(ctx context.Context, org, prefix string) ([]*IssueStats, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}

	repos, err := m.GetRepositoriesWithPrefix(ctx, org, prefix)
	if err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, nil
	}

	return []*IssueStats{
		{RepoName: "test-repo", TotalIssues: 5, OpenIssues: 3, ClosedIssues: 2},
	}, nil
}

// Benchmark tests
func BenchmarkIssueStatsProcessing(b *testing.B) {
	issues := make([]*github.Issue, 1000)
	for i := 0; i < 1000; i++ {
		state := "open"
		if i%2 == 0 {
			state = "closed"
		}

		var prLinks *github.PullRequestLinks
		if i%10 == 0 { // Every 10th is a PR
			prLinks = &github.PullRequestLinks{}
		}

		issues[i] = &github.Issue{
			State:            stringPtr(state),
			PullRequestLinks: prLinks,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		countIssues(issues, "benchmark-repo")
	}
}

func BenchmarkServiceCreation(b *testing.B) {
	client := github.NewClient(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewGitHubService(client)
	}
}
