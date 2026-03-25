//go:build !windows

package cmd

import "os"

func createLink(target, link string) error {
	return os.Symlink(target, link)
}

func removeLink(link string) error {
	return os.Remove(link)
}
