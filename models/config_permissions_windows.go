//go:build windows

package models

import (
	"fmt"
	"os/exec"
	"os/user"
	"strings"
)

const (
	windowsLocalSystemSID    = "*S-1-5-18"
	windowsAdministratorsSID = "*S-1-5-32-544"
	windowsSIDPrefix         = "S-1-"
)

func restrictConfigDirectory(path string) error {
	ownerSID, err := currentUserSID()
	if err != nil {
		return err
	}
	if err := runICACLS(path, "/reset"); err != nil {
		return err
	}
	return runICACLS(path, "/inheritance:r", "/grant:r",
		"*"+ownerSID+":(OI)(CI)(F)", windowsLocalSystemSID+":(OI)(CI)(F)", windowsAdministratorsSID+":(OI)(CI)(F)")
}

func restrictConfigFile(path string) error {
	ownerSID, err := currentUserSID()
	if err != nil {
		return err
	}
	if err := runICACLS(path, "/reset"); err != nil {
		return err
	}
	return runICACLS(path, "/inheritance:r", "/grant:r",
		"*"+ownerSID+":(F)", windowsLocalSystemSID+":(F)", windowsAdministratorsSID+":(F)")
}

func currentUserSID() (string, error) {
	current, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("get current Windows user: %w", err)
	}
	if !strings.HasPrefix(current.Uid, windowsSIDPrefix) {
		return "", fmt.Errorf("current Windows user has an invalid SID")
	}
	for _, char := range current.Uid {
		if (char < '0' || char > '9') && char != 'S' && char != '-' {
			return "", fmt.Errorf("current Windows user has an invalid SID")
		}
	}
	return current.Uid, nil
}

func runICACLS(path string, args ...string) error {
	commandArgs := append([]string{path}, args...)
	output, err := exec.Command("icacls.exe", commandArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("icacls failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
