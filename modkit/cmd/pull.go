package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var pullImpl string

// pullCmd copies a module implementation into the current project.
var pullCmd = &cobra.Command{
	Use:   "pull <module>",
	Short: "Add a module implementation to the current project",
	Long: `Copy a module's implementation files into the current project.

Reads .modkit.yaml to determine the project runtime, selects the appropriate
implementation, copies files into the correct location, and updates .modkit.yaml.

Exit codes:
  0  success
  1  general error
  3  missing dependency (required module not already pulled)
  5  network error
  6  runtime mismatch`,
	Args:    cobra.ExactArgs(1),
	Example: `  modkit pull auth
  modkit pull payments --impl stripe
  modkit pull realtime --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		module := args[0]
		// TODO: implement pull logic
		// 1. Read .modkit.yaml to get runtime and already-selected modules
		// 2. Check module dependencies are satisfied
		// 3. Copy modules/{module}/impl/{impl}-{runtime}/ into project
		// 4. Append module entry to .modkit.yaml
		// 5. Print wiring instructions from AGENT.md
		fmt.Printf("modkit pull %q (impl: %s) (not yet implemented)\n", module, pullImpl)
		return nil
	},
}

func init() {
	pullCmd.Flags().StringVar(&pullImpl, "impl", "",
		"Specific implementation to use (defaults to module's default_impl)")
}
