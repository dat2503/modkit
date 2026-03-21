package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var validateStrict bool

// validateCmd checks that all selected modules are correctly wired.
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate module wiring in the current project",
	Long: `Validate that all selected modules are correctly wired together.

Reads .modkit.yaml and checks:
  - All required dependencies are present
  - Interface implementations satisfy the contracts
  - Config env vars are referenced correctly
  - Module initialization order is correct in bootstrap

Exit codes:
  0  success (all checks pass)
  1  general error
  2  validation failure (wiring errors found)
  3  missing dependency`,
	Example: `  modkit validate
  modkit validate --strict
  modkit validate --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: implement validate logic
		// 1. Read .modkit.yaml
		// 2. For each module, check deps are present
		// 3. Parse bootstrap file and verify init order matches rulebook §3
		// 4. Check that required env vars are declared in config
		// 5. If --strict, also check for unused imports and lint issues
		fmt.Println("modkit validate (not yet implemented)")
		fmt.Println("All checks passed (stub)")
		return nil
	},
}

func init() {
	validateCmd.Flags().BoolVar(&validateStrict, "strict", false,
		"Enable additional strict checks (unused imports, lint)")
}
