package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	upgradeModule string
	upgradeAll    bool
)

// upgradeCmd upgrades module implementations to the latest registry version.
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade module implementations to the latest registry version",
	Long: `Upgrade one or all module implementations in the current project.

Fetches the latest version of each module's implementation from the registry
and applies changes, showing a diff before applying.

Exit codes:
  0  success
  1  general error
  5  network error`,
	Example: `  modkit upgrade --all
  modkit upgrade --module auth
  modkit upgrade --module payments --no-prompt`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: implement upgrade logic
		// 1. Read .modkit.yaml
		// 2. Fetch latest registry version
		// 3. For each module (or --module), diff current vs latest
		// 4. Show diff, prompt to apply (unless --no-prompt)
		// 5. Apply changes and update .modkit.yaml registry_version
		if upgradeAll {
			fmt.Println("modkit upgrade --all (not yet implemented)")
		} else if upgradeModule != "" {
			fmt.Printf("modkit upgrade --module %q (not yet implemented)\n", upgradeModule)
		} else {
			return fmt.Errorf("specify --all or --module <name>")
		}
		return nil
	},
}

func init() {
	upgradeCmd.Flags().StringVar(&upgradeModule, "module", "", "Upgrade a specific module")
	upgradeCmd.Flags().BoolVar(&upgradeAll, "all", false, "Upgrade all modules")
}
