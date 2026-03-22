package cmd

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Registry represents the parsed orchestration/registry.yaml.
type Registry struct {
	DefaultRuntime  string          `yaml:"default_runtime"`
	DefaultFrontend string          `yaml:"default_frontend"`
	Runtimes        []RegistryRuntime `yaml:"runtimes"`
	Modules         []RegistryModule  `yaml:"modules"`
}

// RegistryRuntime describes a supported backend runtime.
type RegistryRuntime struct {
	Name string `yaml:"name"`
}

// RegistryModule describes a module entry in the registry.
type RegistryModule struct {
	Name            string         `yaml:"name"`
	AlwaysInclude   bool           `yaml:"always_include"`
	DefaultImpl     string         `yaml:"default_impl"`
	DefaultImplGo   string         `yaml:"default_impl_go"`
	DefaultImplBun  string         `yaml:"default_impl_bun"`
	Implementations []RegistryImpl `yaml:"implementations"`
}

// RegistryImpl describes one implementation of a module.
type RegistryImpl struct {
	Name     string            `yaml:"name"`
	Runtimes map[string]string `yaml:"runtimes"`
}

// ScaffoldModule is a resolved (name, impl) pair for scaffolding.
type ScaffoldModule struct {
	Name string
	Impl string
}

// loadRegistry reads and parses the registry YAML file.
func loadRegistry(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read registry: %w", err)
	}
	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parse registry: %w", err)
	}
	return &reg, nil
}

// findModule looks up a module by name in the registry. Returns nil if not found.
func findModule(reg *Registry, name string) *RegistryModule {
	for i := range reg.Modules {
		if reg.Modules[i].Name == name {
			return &reg.Modules[i]
		}
	}
	return nil
}

// defaultImpl returns the default implementation name for a module + runtime.
func defaultImpl(mod *RegistryModule, runtime string) string {
	if runtime == "go" && mod.DefaultImplGo != "" {
		return mod.DefaultImplGo
	}
	if runtime == "bun" && mod.DefaultImplBun != "" {
		return mod.DefaultImplBun
	}
	return mod.DefaultImpl
}

// implSupportsRuntime checks whether an impl name supports the given runtime.
func implSupportsRuntime(mod *RegistryModule, implName, runtime string) bool {
	for _, impl := range mod.Implementations {
		if impl.Name == implName {
			_, ok := impl.Runtimes[runtime]
			return ok
		}
	}
	return false
}

// resolveModules resolves the user-supplied module list against the registry,
// auto-prepending always_include modules, and returning the ordered list.
func resolveModules(reg *Registry, names []string, runtime string) ([]ScaffoldModule, error) {
	resolved := make([]ScaffoldModule, 0, len(names)+2)
	seen := map[string]bool{}

	for _, entry := range names {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		var name, impl string
		if idx := strings.Index(entry, ":"); idx >= 0 {
			name = entry[:idx]
			impl = entry[idx+1:]
		} else {
			name = entry
		}

		mod := findModule(reg, name)
		if mod == nil {
			return nil, fmt.Errorf("module %q not found in registry", name)
		}

		if impl == "" {
			impl = defaultImpl(mod, runtime)
		}
		if impl == "" {
			return nil, fmt.Errorf("module %q has no default implementation for runtime %q", name, runtime)
		}
		if !implSupportsRuntime(mod, impl, runtime) {
			return nil, &runtimeMismatchError{module: name, impl: impl, runtime: runtime}
		}

		resolved = append(resolved, ScaffoldModule{Name: name, Impl: impl})
		seen[name] = true
	}

	// Prepend always_include modules not already in the list.
	var prepend []ScaffoldModule
	for _, mod := range reg.Modules {
		if mod.AlwaysInclude && !seen[mod.Name] {
			impl := defaultImpl(&mod, runtime)
			if impl == "" {
				continue
			}
			prepend = append(prepend, ScaffoldModule{Name: mod.Name, Impl: impl})
			seen[mod.Name] = true
		}
	}

	return append(prepend, resolved...), nil
}

// runtimeMismatchError is returned when an impl doesn't support the chosen runtime (exit code 6).
type runtimeMismatchError struct {
	module, impl, runtime string
}

func (e *runtimeMismatchError) Error() string {
	return fmt.Sprintf("module %q implementation %q does not support runtime %q", e.module, e.impl, e.runtime)
}
