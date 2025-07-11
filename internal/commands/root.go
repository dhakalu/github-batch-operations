package commands

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-repo-manager",
	Short: "A CLI tool to manage Go repositories",
	Long:  `A command-line interface for managing multiple Go repositories efficiently.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Initialize subcommands here
	rootCmd.AddCommand(newGetIssueCountCmd())
	rootCmd.AddCommand(newCodeownersCmd())
}
