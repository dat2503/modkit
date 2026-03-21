// Package cmd implements the modkit CLI commands.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	outputFormat string
	noPrompt     bool
)

// rootCmd is the base command for the modkit CLI.
var rootCmd = &cobra.Command{
	Use:   "modkit",
	Short: "modkit — module registry CLI for SaaS scaffolding",
	Long: `modkit is the CLI tool for the modkit module registry.

It scaffolds new projects, manages pluggable modules, and validates
that all selected modules are correctly wired together.

Documentation: https://github.com/dat2503/modkit`,
	SilenceUsage: true,
}

// Execute runs the root command. Called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&outputFormat, "output", "o", "table",
		`Output format: "table" or "json"`,
	)
	rootCmd.PersistentFlags().BoolVar(
		&noPrompt, "no-prompt", false,
		"Disable interactive prompts; fail on missing required input",
	)

	rootCmd.AddCommand(
		initCmd,
		listCmd,
		infoCmd,
		pullCmd,
		validateCmd,
		upgradeCmd,
		doctorCmd,
		runtimesCmd,
	)
}
