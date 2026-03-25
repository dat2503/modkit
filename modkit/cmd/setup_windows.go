//go:build windows

package cmd

import (
	"os"
	"os/exec"
)

// createLink creates a directory junction on Windows (no elevation required).
func createLink(target, link string) error {
	return exec.Command("cmd", "/c", "mklink", "/J", link, target).Run()
}

func removeLink(link string) error {
	// Junctions are removed with RemoveAll (they don't follow the link).
	return os.RemoveAll(link)
}
