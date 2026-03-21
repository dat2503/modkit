package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listRuntime string

// listCmd lists all available modules in the registry.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available modules in the registry",
	Long: `List all modules available in the module registry.

Reads orchestration/registry.yaml and prints each module's name, category,
phase, and available implementations. Use --runtime to filter by runtime support.

Exit codes:
  0  success
  1  general error`,
	Example: `  modkit list
  modkit list --runtime go
  modkit list --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: implement list logic
		// 1. Load orchestration/registry.yaml
		// 2. Filter by --runtime if specified
		// 3. Print table or JSON based on --output flag
		fmt.Println("Available modules (not yet implemented):")
		fmt.Println("  auth          mvp  Clerk (go, bun)")
		fmt.Println("  payments      mvp  Stripe (go, bun)")
		fmt.Println("  email         mvp  Resend (go, bun)")
		fmt.Println("  storage       mvp  S3 (go, bun)")
		fmt.Println("  cache         mvp  Redis (go, bun)")
		fmt.Println("  observability mvp  OpenTelemetry (go, bun)")
		fmt.Println("  error-tracking mvp Sentry (go, bun)")
		fmt.Println("  jobs          mvp  Asynq (go) / BullMQ (bun)")
		fmt.Println("  realtime      v2   WebSocket (go, bun)")
		fmt.Println("  search        v2   Elasticsearch (go, bun)")
		fmt.Println("  feature-flags v2   Flagsmith (go, bun)")
		fmt.Println("  cicd          v2   GitHub Actions (go, bun)")
		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listRuntime, "runtime", "", `Filter by runtime: "go" or "bun"`)
}
