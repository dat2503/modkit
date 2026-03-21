package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	initName      string
	initRuntime   string
	initModules   []string
	initGoModule  string
	initGoVersion string
	initBunVersion string
)

// initCmd scaffolds a new project from the registry templates.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Scaffold a new project from the module registry",
	Long: `Scaffold a new project with selected modules wired and configured.

Reads the registry, prompts for module selection (unless --modules is set),
then generates a project from the appropriate runtime template.

Exit codes:
  0  success
  1  general error
  4  config error (invalid runtime, missing required flags)`,
	Example: `  # Interactive scaffold
  modkit init

  # Non-interactive
  modkit init --name my-app --runtime go --modules auth,payments,email \
    --go-module github.com/user/my-app --no-prompt`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: implement scaffold logic
		// 1. Validate flags / prompt for missing values
		// 2. Load registry from orchestration/registry.yaml
		// 3. Resolve module selection and their default impls
		// 4. Select template directory based on --runtime
		// 5. Execute Go text/template on each .tmpl file
		// 6. Write output to --name directory
		// 7. Run go mod tidy / bun install as post-step
		fmt.Printf("modkit init: would scaffold %q (runtime: %s)\n", initName, initRuntime)
		fmt.Printf("  modules: %v\n", initModules)
		fmt.Println("(not yet implemented)")
		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&initName, "name", "", "Project name (required)")
	initCmd.Flags().StringVar(&initRuntime, "runtime", "", `Backend runtime: "go" or "bun" (required)`)
	initCmd.Flags().StringSliceVar(&initModules, "modules", nil,
		`Comma-separated list of module names to include (e.g. auth,payments,email)`)
	initCmd.Flags().StringVar(&initGoModule, "go-module", "", `Go module path (e.g. github.com/user/my-app); required when --runtime=go`)
	initCmd.Flags().StringVar(&initGoVersion, "go-version", "1.22", "Go version to use in go.mod")
	initCmd.Flags().StringVar(&initBunVersion, "bun-version", "1.1", "Bun version to target")
}
