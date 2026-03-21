package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var infoAgent bool

// infoCmd prints details about a specific module.
var infoCmd = &cobra.Command{
	Use:   "info <module>",
	Short: "Show details for a module",
	Long: `Show detailed information about a module from the registry.

Prints the module manifest (module.yaml) including available implementations,
dependencies, config schema, and wiring notes. Use --agent to print the full
AGENT.md for AI-assisted wiring.

Exit codes:
  0  success
  1  general error (module not found)`,
	Args:    cobra.ExactArgs(1),
	Example: `  modkit info auth
  modkit info payments --agent
  modkit info cache --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		module := args[0]
		// TODO: implement info logic
		// 1. Find modules/{module}/module.yaml
		// 2. If --agent, read modules/{module}/docs/AGENT.md and print it
		// 3. Otherwise print module.yaml fields as table or JSON
		fmt.Printf("modkit info %q (not yet implemented)\n", module)
		if infoAgent {
			fmt.Printf("Would print AGENT.md for module %q\n", module)
		}
		return nil
	},
}

func init() {
	infoCmd.Flags().BoolVar(&infoAgent, "agent", false, "Print the full AGENT.md for AI-assisted wiring")
}
