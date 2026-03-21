package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Install modkit skills, hooks, and Claude Code integration",
	Long: `Sets up your local environment for using modkit with Claude Code.

This command:
  1. Links the registry to ~/.modkit/cache/
  2. Installs the /new-app skill to ~/.claude/commands/
  3. Updates ~/.claude/CLAUDE.md with modkit workflow instructions`,
	RunE: runSetup,
}

func runSetup(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}

	absRegistry, err := filepath.Abs(registryPath)
	if err != nil {
		return fmt.Errorf("resolve registry path: %w", err)
	}

	// 1. Ensure ~/.modkit/cache/ exists and points to the registry.
	cacheDir := filepath.Join(home, ".modkit", "cache")
	if err := linkOrCopyCache(absRegistry, cacheDir); err != nil {
		return err
	}
	fmt.Printf("  Registry linked: %s → %s\n", cacheDir, absRegistry)

	// 2. Install user-level skills.
	claudeCommands := filepath.Join(home, ".claude", "commands")
	if err := os.MkdirAll(claudeCommands, 0755); err != nil {
		return fmt.Errorf("create commands dir: %w", err)
	}

	skillSrc := filepath.Join(absRegistry, "modkit", "skills", "new-app.md")
	skillDst := filepath.Join(claudeCommands, "new-app.md")
	if err := copyFile(skillSrc, skillDst); err != nil {
		return fmt.Errorf("install new-app skill: %w", err)
	}
	fmt.Printf("  Skill installed: %s\n", skillDst)

	// 3. Update ~/.claude/CLAUDE.md.
	claudeMD := filepath.Join(home, ".claude", "CLAUDE.md")
	if err := updateClaudeMD(claudeMD); err != nil {
		return err
	}
	fmt.Printf("  CLAUDE.md updated: %s\n", claudeMD)

	fmt.Println("\nSetup complete. Available skills:")
	fmt.Println("  /new-app — start a new web application")
	fmt.Println("\nInside a scaffolded project, you'll also have:")
	fmt.Println("  /add-module — add a module from the registry")
	fmt.Println("  /validate  — check module wiring")
	return nil
}

// linkOrCopyCache creates ~/.modkit/cache/ as a symlink to the registry,
// or updates it if it already points elsewhere.
func linkOrCopyCache(registryPath, cacheDir string) error {
	parent := filepath.Dir(cacheDir)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return err
	}

	// Check if cache already exists and is correct.
	if target, err := os.Readlink(cacheDir); err == nil {
		if filepath.Clean(target) == filepath.Clean(registryPath) {
			return nil // already correct
		}
		os.Remove(cacheDir) // remove stale link
	} else if fi, err := os.Stat(cacheDir); err == nil && fi.IsDir() {
		// It's a real directory, not a symlink. Leave it.
		return nil
	}

	return os.Symlink(registryPath, cacheDir)
}

const modkitSection = `## Web App Projects

When I say "new webapp", "new web project", "build me an app", "I have an idea",
or any variation implying I want to build a web application — or when you detect
a ` + "`.modkit.yaml`" + ` in the current directory:

### New project (no .modkit.yaml yet)
1. Run the /new-app skill, OR manually follow the workflow below
2. Read the modkit registry context:
   - ~/.modkit/cache/docs/agent-instructions.md
   - ~/.modkit/cache/orchestration/playbook.md
   - ~/.modkit/cache/orchestration/registry.yaml
3. Follow Phase 0 (Intake) → Phase 1 (Module Selection) of the playbook
4. Wait for my approval on the module selection
5. Run modkit init with the approved modules
6. Continue to Phase 2 (Architecture) → Phase 3 (Build)
7. Stop at every checkpoint for my approval

### Existing project (.modkit.yaml exists)
1. Read .modkit.yaml to understand the project's modules and runtime
2. Read ~/.modkit/cache/orchestration/composition-rulebook.md for wiring rules
3. Use /add-module to add modules, /validate to check wiring
4. Follow the composition rulebook for all code changes

### Available skills
- ` + "`/new-app`" + ` — start a new web application from scratch
- ` + "`/add-module`" + ` — add a module to the current project (project-level)
- ` + "`/validate`" + ` — check module wiring is correct (project-level)`

const sectionHeader = "## Web App Projects"

// updateClaudeMD appends or replaces the Web App Projects section in CLAUDE.md.
func updateClaudeMD(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create the file with the section.
			return os.WriteFile(path, []byte("# Personal Preferences\n\n"+modkitSection+"\n"), 0644)
		}
		return err
	}

	text := string(content)

	if idx := strings.Index(text, sectionHeader); idx >= 0 {
		// Find the end of the section (next ## heading or EOF).
		rest := text[idx+len(sectionHeader):]
		endIdx := strings.Index(rest, "\n## ")
		if endIdx >= 0 {
			text = text[:idx] + modkitSection + "\n" + rest[endIdx+1:]
		} else {
			text = text[:idx] + modkitSection + "\n"
		}
	} else {
		// Append the section.
		if !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		text += "\n" + modkitSection + "\n"
	}

	return os.WriteFile(path, []byte(text), 0644)
}
