package main

import (
	"go-repo-manager/internal/commands"
	"go-repo-manager/internal/logger"
)

func main() {
	// Initialize logger
	logger.Setup()

	// Execute commands
	commands.Execute()
}
