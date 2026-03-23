package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// TemplateData is the data passed to every .tmpl file during scaffolding.
type TemplateData struct {
	Name            string
	Modules         []ScaffoldModule
	Frontend        string // "vite" or "next"
	GoModule        string
	GoVersion       string
	BunVersion      string
	RegistryVersion string
	RegistryPath    string // absolute path to registry root; used in go.mod replace directive
}

// HasModule returns true if the named module is in the scaffold list.
func (d TemplateData) HasModule(name string) bool {
	for _, m := range d.Modules {
		if m.Name == name {
			return true
		}
	}
	return false
}

// ImplFor returns the implementation name for the named module, or "" if not found.
// Used in templates for impl-aware conditionals, e.g. {{if eq (.ImplFor "auth") "better-auth"}}.
func (d TemplateData) ImplFor(name string) string {
	for _, m := range d.Modules {
		if m.Name == name {
			return m.Impl
		}
	}
	return ""
}

// processTemplates walks templateDir, renders .tmpl files with data, and copies
// all other files verbatim into outDir, preserving the directory structure.
func processTemplates(templateDir, outDir string, data TemplateData) error {
	return filepath.WalkDir(templateDir, func(src string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(templateDir, src)
		if err != nil {
			return err
		}

		dst := filepath.Join(outDir, rel)

		if d.IsDir() {
			return os.MkdirAll(dst, 0755)
		}

		if strings.HasSuffix(src, ".tmpl") {
			dst = strings.TrimSuffix(dst, ".tmpl")
			return renderTemplate(src, dst, data)
		}

		return copyFile(src, dst)
	})
}

// renderTemplate executes a Go text/template file and writes the result to dst.
func renderTemplate(src, dst string, data TemplateData) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read template %s: %w", src, err)
	}

	name := filepath.Base(src)
	tmpl, err := template.New(name).Parse(string(content))
	if err != nil {
		return fmt.Errorf("parse template %s: %w", src, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("render template %s: %w", src, err)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	return os.WriteFile(dst, buf.Bytes(), 0644)
}

// copyFile copies a file verbatim from src to dst.
func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// copyModuleImpls copies each module's implementation directory into the
// generated project at outDir/apps/api/modules/{name}/.
func copyModuleImpls(registryPath, outDir string, modules []ScaffoldModule) error {
	for _, mod := range modules {
		if mod.ImplDir == "" {
			continue // no impl directory for this module (e.g. cicd has no code)
		}

		src := filepath.Join(registryPath, mod.ImplDir)

		// If the impl directory doesn't exist, skip it.
		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue
		}

		dst := filepath.Join(outDir, "apps", "api", "modules", mod.Name)
		if err := copyDirWithImportRewrite(src, dst); err != nil {
			return fmt.Errorf("copy module %s: %w", mod.Name, err)
		}
	}
	return nil
}

// copyContracts copies the contracts directory for the given runtime into the
// generated project at outDir/apps/api/contracts/{lang}/.
func copyContracts(registryPath, outDir, runtime string) error {
	lang := "ts"
	if runtime == "go" {
		lang = "go"
	}
	src := filepath.Join(registryPath, "contracts", lang)
	dst := filepath.Join(outDir, "apps", "api", "contracts", lang)
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil // no contracts to copy
	}
	return copyDir(src, dst)
}

// copyDir recursively copies all files from src to dst.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(path, target)
	})
}

// copyDirWithImportRewrite copies files from src to dst, rewriting import paths
// in .ts and .go files so that references to the registry's contracts directory
// point to the project-local contracts directory instead.
func copyDirWithImportRewrite(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		ext := filepath.Ext(path)
		if ext == ".ts" || ext == ".go" {
			return copyFileWithImportRewrite(path, target)
		}
		return copyFile(path, target)
	})
}

// copyFileWithImportRewrite copies a file, rewriting contract import paths.
// Registry impls import from '../../../contracts/{lang}/...' but in the generated
// project, modules live at apps/api/modules/{name}/ and contracts at apps/api/contracts/{lang}/.
func copyFileWithImportRewrite(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	s := string(content)
	// TS: ../../../contracts/ts/ → ../../contracts/ts/  (from modules/{name}/ to contracts/ts/)
	s = strings.ReplaceAll(s, "'../../../contracts/ts/", "'../../contracts/ts/")
	s = strings.ReplaceAll(s, "\"../../../contracts/ts/", "\"../../contracts/ts/")
	// Go: github.com/dat2503/modkit/contracts/go → same (Go uses module paths, not relative)
	// No change needed for Go — the go.mod replace directive handles it.

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	return os.WriteFile(dst, []byte(s), 0644)
}
