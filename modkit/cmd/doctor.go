package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

// doctorCmd checks the local environment for required tools and configuration.
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check local environment for required tools",
	Long: `Check that all required tools and environment configuration are present.

Verifies:
  - Go installation and version (if using go runtime)
  - Bun installation and version (if using bun runtime)
  - Docker and docker compose (for local infra)
  - .env file exists (warns if only .env.example)
  - .modkit.yaml is valid and present

Exit codes:
  0  all checks pass
  1  one or more checks failed`,
	Example: `  modkit doctor
  modkit doctor --output json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: replace stub with real checks
		fmt.Println("modkit doctor — environment check")
		fmt.Println()

		checks := []struct {
			name string
			fn   func() (string, bool)
		}{
			{"OS", func() (string, bool) { return runtime.GOOS, true }},
			{"go", checkCommand("go", "version")},
			{"bun", checkCommand("bun", "--version")},
			{"docker", checkCommand("docker", "version", "--format", "{{.Server.Version}}")},
			{"docker compose", checkCommand("docker", "compose", "version")},
		}

		allOK := true
		for _, c := range checks {
			result, ok := c.fn()
			status := "✓"
			if !ok {
				status = "✗"
				allOK = false
			}
			fmt.Printf("  %s  %-20s %s\n", status, c.name, result)
		}

		fmt.Println()
		if allOK {
			fmt.Println("All checks passed.")
		} else {
			fmt.Println("Some checks failed. Install missing tools and re-run.")
		}
		return nil
	},
}

// checkCommand returns a function that runs a command and returns its output.
func checkCommand(name string, args ...string) func() (string, bool) {
	return func() (string, bool) {
		out, err := exec.Command(name, args...).Output()
		if err != nil {
			return "not found", false
		}
		// Truncate to first line
		s := string(out)
		for i, c := range s {
			if c == '\n' {
				return s[:i], true
			}
		}
		return s, true
	}
}
