// Package cmd contains all CLI commands for asentric.
package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "asentric",
	Short: "Asentric - Blockchain Security Framework",
	Long: `Asentric is a blockchain security framework for building
real-time smart contract monitoring systems.

Inspired by ponder.sh, Asentric provides:
  • SDK for custom security rules
  • CLI for project scaffolding
  • Runtime for event processing

Commands:
  init      Create a new Asentric project
  version   Show version information

Documentation: https://github.com/asentric/asentric`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(versionCmd)
}
