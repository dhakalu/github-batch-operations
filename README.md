# Go Repository Manager

Go Repository Manager is a command-line interface (CLI) tool designed to help developers manage multiple repositories efficiently. This tool provides commands to analyze and manage GitHub repositories at scale.

## Features

- **Issue Count Analysis**: Get issue counts from GitHub repositories individually or by prefix
- **CODEOWNERS Management**: Add or update CODEOWNERS files across multiple repositories efficiently
- **Visual Repository Status**: Clear visual indicators (✅/❌) to quickly identify clean vs problematic repositories
- **Smart Sorting**: Repositories are sorted with clean ones first, then by issue count for easy prioritization
- **Batch Operations**: Process multiple repositories concurrently with configurable concurrency
- Count open, closed, and total issues across repositories
- Support for both organization and user repository analysis
- **Enhanced Reporting**: Rich summary statistics including success rates and clean repository percentages
- GitHub API integration with token-based authentication
- Configurable concurrency for processing multiple repositories
- Utility functions for logging and error handling

## Installation

To install the Go Repository Manager, follow these steps:

1. Ensure you have Go installed on your machine. You can download it from [golang.org](https://golang.org/dl/).
2. Clone the repository:

   ```bash
   git clone https://github.com/yourusername/go-repo-manager.git
   ```

3. Navigate to the project directory:

   ```bash
   cd go-repo-manager
   ```

4. Install the dependencies:

   ```bash
   go mod tidy
   ```

5. Build the project:

   ```bash
   go build -o bin/go-repo-manager ./cmd
   ```

## Usage

To use the Go Repository Manager, run the following command:

```bash
./bin/go-repo-manager [command]
```

### Commands

#### `get-issue-count`

Get the count of issues from GitHub repositories. This command supports three modes and can work with both organizations and user accounts:

**Single Repository Mode:**
```bash
# For organization repositories
./bin/go-repo-manager get-issue-count --org myorg --repo myrepo

# For user repositories
./bin/go-repo-manager get-issue-count --username myusername --repo myrepo
```

**Repository Prefix Mode:**
```bash
# For organization repositories
./bin/go-repo-manager get-issue-count --org myorg --repo-prefix api-

# For user repositories
./bin/go-repo-manager get-issue-count --username myusername --repo-prefix api-
```

**All Repositories Mode:**
```bash
# For organization repositories
./bin/go-repo-manager get-issue-count --org myorg

# For user repositories
./bin/go-repo-manager get-issue-count --username myusername
```

**Flags:**
- `--org string`: GitHub organization name (mutually exclusive with --username)
- `--username string`: GitHub username (mutually exclusive with --org)
- `--repo string`: Specific repository name (optional)
- `--repo-prefix string`: Repository name prefix to filter repositories (optional)
- `--token string`: GitHub personal access token (optional, can also be set via GITHUB_TOKEN env var)
- `--concurrency int`: Maximum number of concurrent workers for processing repositories (default: 1)

**Examples:**
```bash
# Get issue count for a specific repository in an organization
./bin/go-repo-manager get-issue-count --org kubernetes --repo kubernetes

# Get issue count for a specific repository for a user
./bin/go-repo-manager get-issue-count --username octocat --repo Hello-World

# Get issue count for all repositories starting with "api-" in an organization
./bin/go-repo-manager get-issue-count --org myorg --repo-prefix api-

# Get issue count for all repositories starting with "my-" for a user
./bin/go-repo-manager get-issue-count --username myusername --repo-prefix my-

# Get issue count for ALL repositories in the organization
./bin/go-repo-manager get-issue-count --org myorg

# Get issue count for ALL repositories for a user
./bin/go-repo-manager get-issue-count --username myusername

# Use GitHub token for higher rate limits
export GITHUB_TOKEN=your_personal_access_token
./bin/go-repo-manager get-issue-count --org myorg --repo-prefix service-
```

**Output:**
The command provides a visual, easy-to-read report with:
- ✅ Checkmarks for repositories without issues (clean repositories)
- ❌ Cross marks for repositories with issues
- Sorted display: clean repositories first, then repositories with issues
- Detailed breakdown of total, open, and closed issues per repository
- Summary statistics including clean repository percentage
- Enhanced formatting with emojis and visual separators

**Sample Output:**
```
📋 Repository Analysis:
----------------------------------------------------------------------
✅ Repository: myorg/clean-repo (CLEAN)
📊 Total Issues: 0

❌ Repository: myorg/busy-repo (HAS ISSUES)
📊 Total Issues: 15
🔓 Open Issues: 8
✔️  Closed Issues: 7

======================================================================
📊 SUMMARY for all repositories for organization 'myorg':
----------------------------------------------------------------------
📁 Total Repositories: 10
✅ Clean Repositories (no issues): 7
❌ Repositories with issues: 3
----------------------------------------------------------------------
🐛 Total Issues across all repos: 45
🔓 Total Open Issues: 22
✔️  Total Closed Issues: 23
📈 Clean Repository Rate: 70.0%
======================================================================
```

The enhanced output makes it easy to quickly identify:
- Which repositories have issues and which are clean
- Repositories that need attention (those with issues are clearly marked)
- Overall repository health with percentage statistics
- Prioritized view: clean repositories are shown first

**Note:** The command excludes pull requests and only counts actual issues.

#### `codeowners`

Add or update CODEOWNERS files in GitHub repositories. This command supports the same modes as `get-issue-count` and can work with both organizations and user accounts:

**Single Repository Mode:**
```bash
# For organization repositories
./bin/go-repo-manager codeowners --org myorg --repo myrepo --codeowner-file ./CODEOWNERS

# For user repositories
./bin/go-repo-manager codeowners --username myusername --repo myrepo --codeowner-file ./CODEOWNERS
```

**Repository Prefix Mode:**
```bash
# For organization repositories
./bin/go-repo-manager codeowners --org myorg --repo-prefix api- --codeowner-file ./CODEOWNERS

# For user repositories
./bin/go-repo-manager codeowners --username myusername --repo-prefix my- --codeowner-file ./CODEOWNERS
```

**All Repositories Mode:**
```bash
# For organization repositories
./bin/go-repo-manager codeowners --org myorg --codeowner-file ./CODEOWNERS

# For user repositories
./bin/go-repo-manager codeowners --username myusername --codeowner-file ./CODEOWNERS
```

**Flags:**
- `--org string`: GitHub organization name (mutually exclusive with --username)
- `--username string`: GitHub username (mutually exclusive with --org)
- `--repo string`: Specific repository name (optional)
- `--repo-prefix string`: Repository name prefix to filter repositories (optional)
- `--codeowner-file string`: Path to the CODEOWNERS file to add to repositories (required)
- `--token string`: GitHub personal access token (optional, can also be set via GITHUB_TOKEN env var)
- `--concurrency int`: Maximum number of concurrent workers for processing repositories (default: 1)

**Examples:**
```bash
# Add CODEOWNERS to a specific repository in an organization
./bin/go-repo-manager codeowners --org myorg --repo myrepo --codeowner-file ./team-CODEOWNERS

# Add CODEOWNERS to all repositories starting with "api-" in an organization
./bin/go-repo-manager codeowners --org myorg --repo-prefix api- --codeowner-file ./CODEOWNERS

# Add CODEOWNERS to ALL repositories for a user
./bin/go-repo-manager codeowners --username myusername --codeowner-file ./CODEOWNERS

# Use concurrency for faster processing of multiple repositories
./bin/go-repo-manager codeowners --org myorg --repo-prefix service- --codeowner-file ./CODEOWNERS --concurrency 5
```

**Sample Output:**
```
📋 CODEOWNERS Update Results:
----------------------------------------------------------------------
✅ SUCCESSFUL UPDATES (8 repositories):
  ✅ myorg/api-service
  ✅ myorg/api-gateway
  ✅ myorg/api-auth

❌ FAILED UPDATES (1 repositories):
  ❌ myorg/archived-repo

======================================================================
📊 SUMMARY for repositories with prefix 'api-' for organization 'myorg':
----------------------------------------------------------------------
📁 Total Repositories: 9
✅ Successful Updates: 8
❌ Failed Updates: 1
📈 Success Rate: 88.9%
📍 CODEOWNERS files added/updated at: .github/CODEOWNERS
======================================================================
```

**Features:**
- **Smart File Management**: Automatically creates `.github/` directory if it doesn't exist
- **Update Existing Files**: Updates existing CODEOWNERS files or creates new ones
- **Batch Processing**: Process multiple repositories concurrently for efficiency
- **Detailed Reporting**: Visual feedback showing success/failure for each repository
- **Safe Operations**: Each repository operation is independent - failures don't affect other repositories

**Note:** The command creates or updates the CODEOWNERS file at `.github/CODEOWNERS` in each repository with a descriptive commit message.

### Authentication

For better rate limits and access to private repositories, set your GitHub personal access token:

```bash
export GITHUB_TOKEN=your_personal_access_token
```

Or pass it directly as a flag:

```bash
./bin/go-repo-manager get-issue-count --token your_token --org myorg --repo myrepo
```

## Testing

This project includes comprehensive unit tests with mocking strategies to ensure reliability and maintainability. The test suite covers all major functionality including HTTP integration, business logic, error handling, and edge cases.

### Running Tests

#### Run All Tests
```bash
# Run all tests in the project
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test ./internal/repo
go test -v ./internal/repo
```

#### Run Tests with Coverage
```bash
# Generate coverage report
go test -cover ./...

# Generate detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

#### Run Benchmark Tests
```bash
# Run benchmark tests
go test -bench=. ./internal/repo

# Run benchmarks with memory allocation stats
go test -bench=. -benchmem ./internal/repo
```

### Writing Tests

This project follows Go testing best practices with a focus on testability and maintainability. Here's how to write effective tests for this codebase:

#### Test File Structure

Test files should be placed alongside the code they test with a `_test.go` suffix:
```
internal/
  repo/
    github_service.go      # Implementation
    github_service_test.go # Tests
```

#### Testing Patterns Used

**1. Interface-Based Mocking**
The project uses interfaces to enable easy mocking:
```go
// Define interface for testability
type GitHubClient interface {
    GetIssueStatsForRepo(ctx context.Context, org, repoName string) (*IssueStats, error)
    // ...other methods
}

// Implement mock for testing
type mockGitHubService struct {
    shouldError bool
    errorMsg    string
}

func (m *mockGitHubService) GetIssueStatsForRepo(ctx context.Context, org, repoName string) (*IssueStats, error) {
    if m.shouldError {
        return nil, errors.New(m.errorMsg)
    }
    return &IssueStats{...}, nil
}
```

**2. HTTP Mock Server Testing**
For testing HTTP integrations, use `httptest.NewServer`:
```go
func TestWithMockServer(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Mock GitHub API responses
        switch {
        case strings.Contains(r.URL.Path, "/repos/org/repo"):
            json.NewEncoder(w).Encode(mockResponse)
        }
    }))
    defer server.Close()
    
    // Point client to test server
    client := github.NewClient(nil)
    client.BaseURL, _ = client.BaseURL.Parse(server.URL + "/")
}
```

**3. Business Logic Testing**
Test core logic separately from HTTP concerns:
```go
func TestBusinessLogic(t *testing.T) {
    tests := []struct {
        name     string
        input    []*github.Issue
        expected *IssueStats
    }{
        {
            name: "Mixed issues and PRs",
            input: []*github.Issue{
                {State: stringPtr("open"), PullRequestLinks: nil},
                {State: stringPtr("open"), PullRequestLinks: &github.PullRequestLinks{}}, // PR
            },
            expected: &IssueStats{TotalIssues: 1, OpenIssues: 1},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic here
        })
    }
}
```

#### Test Categories

**Unit Tests**
- Test individual functions and methods
- Use mocks for external dependencies
- Focus on business logic and edge cases

**Integration Tests**
- Test interactions between components
- Use HTTP mock servers for API testing
- Verify end-to-end workflows

**Error Handling Tests**
- Test all error scenarios
- Verify appropriate error messages
- Test graceful degradation

**Performance Tests**
- Use benchmark tests for performance-critical code
- Test with realistic data volumes
- Monitor memory allocations

#### Writing Test Guidelines

**1. Test Structure**
Follow the Arrange-Act-Assert pattern:
```go
func TestFunction(t *testing.T) {
    // Arrange
    input := setupTestData()
    expected := expectedResult()
    
    // Act
    result, err := functionUnderTest(input)
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

**2. Use Helper Functions**
Create reusable test utilities:
```go
// Helper to create string pointers
func stringPtr(s string) *string {
    return &s
}

// Helper for test logger
func createTestLogger() *slog.Logger {
    return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: slog.LevelError, // Keep tests quiet
    }))
}
```

**3. Table-Driven Tests**
Use table-driven tests for multiple scenarios:
```go
func TestMultipleScenarios(t *testing.T) {
    tests := []struct {
        name        string
        input       InputType
        expected    OutputType
        expectError bool
    }{
        {"success case", validInput, expectedOutput, false},
        {"error case", invalidInput, nil, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

**4. Mock Complex Dependencies**
For testing services that depend on external APIs:
```go
// Create interface-based mocks
type MockService struct {
    mock.Mock // If using testify/mock
    // Or simple struct with behavior flags
    shouldError bool
    returnData  interface{}
}
```

**5. Test Error Conditions**
Always test error paths:
```go
func TestErrorHandling(t *testing.T) {
    service := &mockService{shouldError: true, errorMsg: "API error"}
    
    result, err := service.DoSomething()
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "API error")
    assert.Nil(t, result)
}
```

#### Testing Tools and Libraries

The project uses these testing libraries:
- **Standard library**: `testing` package for basic test functionality
- **Testify**: `github.com/stretchr/testify` for assertions and requirements
  - `assert`: For non-fatal assertions
  - `require`: For fatal assertions that stop test execution
- **HTTP testing**: `net/http/httptest` for mocking HTTP servers

#### Running Specific Tests

```bash
# Run specific test function
go test -run TestGetIssueStatsForRepo ./internal/repo

# Run tests matching pattern
go test -run "TestGetIssue.*" ./internal/repo

# Run tests with timeout
go test -timeout 30s ./...

# Run tests in parallel
go test -parallel 4 ./...
```

#### Continuous Integration

Tests should be run in CI/CD pipelines:
```bash
# Example CI script
go test -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## Dependencies

- [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- [github.com/google/go-github/v62](https://github.com/google/go-github) - GitHub API client

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## License

This project is licensed under the MIT License. See the LICENSE file for details.