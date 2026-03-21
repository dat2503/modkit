package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// Runtime describes a supported project runtime.
type Runtime struct {
	Name         string `json:"name"`
	Label        string `json:"label"`
	BuildCmd     string `json:"build_cmd"`
	TestCmd      string `json:"test_cmd"`
	LintCmd      string `json:"lint_cmd"`
	DockerBase   string `json:"docker_base"`
}

var supportedRuntimes = []Runtime{
	{
		Name:       "go",
		Label:      "Go",
		BuildCmd:   "go build ./...",
		TestCmd:    "go test ./...",
		LintCmd:    "go vet ./...",
		DockerBase: "golang:1.22-alpine",
	},
	{
		Name:       "bun",
		Label:      "Bun",
		BuildCmd:   "bun run build",
		TestCmd:    "bun test",
		LintCmd:    "bunx @biomejs/biome check .",
		DockerBase: "oven/bun:1.1-alpine",
	},
}

// runtimesCmd lists the supported backend runtimes.
var runtimesCmd = &cobra.Command{
	Use:   "runtimes",
	Short: "List supported backend runtimes",
	Long: `List all backend runtimes supported by the module registry.

Each runtime has a corresponding project template under templates/project-{runtime}/.

Exit codes:
  0  success`,
	Example: `  modkit runtimes
  modkit runtimes --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if outputFormat == "json" {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(supportedRuntimes)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tLABEL\tBUILD\tTEST\tDOCKER BASE")
		for _, r := range supportedRuntimes {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				r.Name, r.Label, r.BuildCmd, r.TestCmd, r.DockerBase)
		}
		return w.Flush()
	},
}
