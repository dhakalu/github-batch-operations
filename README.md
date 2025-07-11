# Go Repository Manager

Go Repository Manager is a command-line interface (CLI) tool designed to help developers manage multiple repositories efficiently. This tool provides commands to analyze and manage GitHub repositories at scale.

## Features

- **Issue Count Analysis**: Get issue counts from GitHub repositories individually or by prefix
- Count open, closed, and total issues across repositories
- Support for organization-wide repository analysis
- GitHub API integration with token-based authentication
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

Get the count of issues from GitHub repositories. This command supports three modes:

**Single Repository Mode:**
```bash
./bin/go-repo-manager get-issue-count --org myorg --repo myrepo
```

**Repository Prefix Mode:**
```bash
./bin/go-repo-manager get-issue-count --org myorg --repo-prefix api-
```

**All Repositories Mode:**
```bash
./bin/go-repo-manager get-issue-count --org myorg
```

**Flags:**
- `--org string`: GitHub organization name (required)
- `--repo string`: Specific repository name (optional)
- `--repo-prefix string`: Repository name prefix to filter repositories (optional)
- `--token string`: GitHub personal access token (optional, can also be set via GITHUB_TOKEN env var)

**Examples:**
```bash
# Get issue count for a specific repository
./bin/go-repo-manager get-issue-count --org kubernetes --repo kubernetes

# Get issue count for all repositories starting with "api-"
./bin/go-repo-manager get-issue-count --org myorg --repo-prefix api-

# Get issue count for ALL repositories in the organization
./bin/go-repo-manager get-issue-count --org myorg

# Use GitHub token for higher rate limits
export GITHUB_TOKEN=your_personal_access_token
./bin/go-repo-manager get-issue-count --org myorg --repo-prefix service-
```

**Output:**
The command provides detailed information including:
- Total issues per repository
- Open issues count
- Closed issues count
- Summary statistics for prefix mode

**Note:** The command excludes pull requests and only counts actual issues.

### Authentication

For better rate limits and access to private repositories, set your GitHub personal access token:

```bash
export GITHUB_TOKEN=your_personal_access_token
```

Or pass it directly as a flag:

```bash
./bin/go-repo-manager get-issue-count --token your_token --org myorg --repo myrepo
```

## Dependencies

- [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- [github.com/google/go-github/v62](https://github.com/google/go-github) - GitHub API client

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## License

This project is licensed under the MIT License. See the LICENSE file for details.