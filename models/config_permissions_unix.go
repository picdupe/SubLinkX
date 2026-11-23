//go:build !windows

package models

import "os"

func restrictConfigDirectory(path string) error {
	return os.Chmod(path, 0700)
}

func restrictConfigFile(path string) error {
	return os.Chmod(path, 0600)
}
