# Go Repository Manager

Go Repository Manager is a command-line interface (CLI) tool designed to help developers manage multiple Go repositories efficiently. This tool provides commands to clone, update, and delete Go repositories, streamlining the workflow for managing projects.

## Features

- Clone multiple Go repositories with a single command.
- Update existing repositories to the latest version.
- Delete repositories that are no longer needed.
- Utility functions for logging and error handling.

## Installation

To install the Go Repository Manager, follow these steps:

1. Ensure you have Go installed on your machine. You can download it from [golang.org](https://golang.org/dl/).
2. Clone the repository:

   ```
   git clone https://github.com/yourusername/go-repo-manager.git
   ```

3. Navigate to the project directory:

   ```
   cd go-repo-manager
   ```

4. Install the dependencies:

   ```
   go mod tidy
   ```

## Usage

To use the Go Repository Manager, run the following command:

```
go run cmd/main.go [command]
```

### Commands

- `clone`: Clone a Go repository.
- `update`: Update an existing Go repository.
- `delete`: Delete a Go repository.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## License

This project is licensed under the MIT License. See the LICENSE file for details.