package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	initName       string
	initRuntime    string
	initFrontend   string
	initModules    []string
	initGoModule   string
	initGoVersion  string
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
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringVar(&initName, "name", "", "Project name (required)")
	initCmd.Flags().StringVar(&initRuntime, "runtime", "", `Backend runtime: "go" or "bun" (default: bun)`)
	initCmd.Flags().StringVar(&initFrontend, "frontend", "", `Frontend framework: "vite" or "next" (default: vite)`)
	initCmd.Flags().StringSliceVar(&initModules, "modules", nil,
		`Comma-separated list of module names to include (e.g. auth,payments,email)`)
	initCmd.Flags().StringVar(&initGoModule, "go-module", "", `Go module path (e.g. github.com/user/my-app); required when --runtime=go`)
	initCmd.Flags().StringVar(&initGoVersion, "go-version", "1.24", "Go version to use in go.mod")
	initCmd.Flags().StringVar(&initBunVersion, "bun-version", "1.1", "Bun version to target")
}

func runInit(cmd *cobra.Command, args []string) error {
	// 1. Load registry first so we can use defaults.
	regFile := filepath.Join(registryPath, "orchestration", "registry.yaml")
	reg, err := loadRegistry(regFile)
	if err != nil {
		return fmt.Errorf("load registry from %s: %w", regFile, err)
	}

	// 2. Validate / prompt for required flags.
	if initName == "" {
		if noPrompt {
			return configError("--name is required")
		}
		initName = prompt("Project name: ")
	}

	if initRuntime == "" {
		defaultRuntime := reg.DefaultRuntime
		if defaultRuntime == "" {
			defaultRuntime = "bun"
		}
		if noPrompt {
			initRuntime = defaultRuntime
		} else {
			initRuntime = promptDefault(fmt.Sprintf("Runtime (go/bun) [%s]: ", defaultRuntime), defaultRuntime)
		}
	}
	if initRuntime != "go" && initRuntime != "bun" {
		return configError(fmt.Sprintf("invalid runtime %q — must be \"go\" or \"bun\"", initRuntime))
	}

	if initFrontend == "" {
		defaultFrontend := reg.DefaultFrontend
		if defaultFrontend == "" {
			defaultFrontend = "vite"
		}
		if noPrompt {
			initFrontend = defaultFrontend
		} else {
			initFrontend = promptDefault(fmt.Sprintf("Frontend (vite/next) [%s]: ", defaultFrontend), defaultFrontend)
		}
	}
	if initFrontend != "vite" && initFrontend != "next" {
		return configError(fmt.Sprintf("invalid frontend %q — must be \"vite\" or \"next\"", initFrontend))
	}

	if initRuntime == "go" && initGoModule == "" {
		if noPrompt {
			return configError("--go-module is required when --runtime=go")
		}
		initGoModule = prompt("Go module path (e.g. github.com/org/project): ")
	}

	// 3. Ensure output directory doesn't already exist.
	if _, err := os.Stat(initName); err == nil {
		return fmt.Errorf("directory %q already exists", initName)
	}

	// 4. Resolve modules.
	modules, err := resolveModules(reg, initModules, initRuntime)
	if err != nil {
		if _, ok := err.(*runtimeMismatchError); ok {
			os.Exit(6)
		}
		return err
	}

	// 5. Build template data.
	absRegistry, err := filepath.Abs(registryPath)
	if err != nil {
		return err
	}
	data := TemplateData{
		Name:            initName,
		Modules:         modules,
		Frontend:        initFrontend,
		GoModule:        initGoModule,
		GoVersion:       initGoVersion,
		BunVersion:      initBunVersion,
		RegistryVersion: "1.0",
		RegistryPath:    absRegistry,
	}

	// 6. Process templates.
	templateDir := filepath.Join(registryPath, "templates", "project-"+initRuntime+"-"+initFrontend)
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		// Fallback to legacy template path (project-{runtime}) for backwards compatibility.
		legacyDir := filepath.Join(registryPath, "templates", "project-"+initRuntime)
		if _, lerr := os.Stat(legacyDir); os.IsNotExist(lerr) {
			return fmt.Errorf("template directory not found: %s", templateDir)
		}
		templateDir = legacyDir
	}
	fmt.Printf("Scaffolding %q (runtime: %s, frontend: %s)...\n", initName, initRuntime, initFrontend)
	if err := processTemplates(templateDir, initName, data); err != nil {
		return fmt.Errorf("process templates: %w", err)
	}

	// 7. Copy module implementation files.
	if err := copyModuleImpls(registryPath, initName, initRuntime, modules); err != nil {
		return fmt.Errorf("copy module impls: %w", err)
	}

	// 8. Post-processing (non-fatal).
	if initRuntime == "go" {
		runPostStep("go mod tidy", filepath.Join(initName, "apps", "api"))
	} else {
		runPostStep("bun install", filepath.Join(initName, "apps", "api"))
		runPostStep("bun install", filepath.Join(initName, "apps", "web"))
	}

	// 9. Output summary.
	if outputFormat == "json" {
		return printInitJSON(initName, initRuntime, initFrontend, modules)
	}
	printInitTable(initName, initRuntime, initFrontend, modules)
	return nil
}

// configError returns a config error (exit code 4).
func configError(msg string) error {
	fmt.Fprintln(os.Stderr, "config error:", msg)
	os.Exit(4)
	return nil // unreachable
}

// prompt reads a line from stdin.
func prompt(label string) string {
	fmt.Print(label)
	var s string
	fmt.Scanln(&s)
	return s
}

// promptDefault reads a line from stdin, returning defaultVal if the user enters nothing.
func promptDefault(label, defaultVal string) string {
	s := prompt(label)
	if s == "" {
		return defaultVal
	}
	return s
}

// runPostStep executes a shell command in the given directory, printing a warning on failure.
func runPostStep(command, dir string) {
	parts := splitCommand(command)
	c := exec.Command(parts[0], parts[1:]...) //nolint:gosec
	c.Dir = dir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: post-step %q in %s failed: %v\n", command, dir, err)
	}
}

func splitCommand(s string) []string {
	var parts []string
	for _, p := range filepath.SplitList(s) {
		parts = append(parts, p)
	}
	if len(parts) == 1 {
		// filepath.SplitList uses OS path separator; for space-split use manual split.
		parts = nil
		word := ""
		for _, c := range s {
			if c == ' ' {
				if word != "" {
					parts = append(parts, word)
					word = ""
				}
			} else {
				word += string(c)
			}
		}
		if word != "" {
			parts = append(parts, word)
		}
	}
	return parts
}

func printInitTable(name, runtime, frontend string, modules []ScaffoldModule) {
	fmt.Printf("\n✓ Created project %q\n\n", name)
	fmt.Printf("  Runtime:  %s\n", runtime)
	fmt.Printf("  Frontend: %s\n", frontend)
	fmt.Printf("  Modules:\n")
	for _, m := range modules {
		fmt.Printf("    %-20s %s\n", m.Name, m.Impl)
	}
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  cd %s\n", name)
	fmt.Printf("  cp .env.example .env   # fill in your keys\n")
	fmt.Printf("  make setup             # start infra + install deps\n")
	fmt.Printf("  make dev               # start API + web\n")
}

func printInitJSON(name, runtime, frontend string, modules []ScaffoldModule) error {
	type modJSON struct {
		Name string `json:"name"`
		Impl string `json:"impl"`
	}
	mods := make([]modJSON, len(modules))
	for i, m := range modules {
		mods[i] = modJSON{Name: m.Name, Impl: m.Impl}
	}
	out := map[string]any{
		"name":     name,
		"runtime":  runtime,
		"frontend": frontend,
		"modules":  mods,
		"path":     name,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
