//go:build windows

package data

import (
	"os/exec"
	"os/user"
)

// platformLock sets file permissions on Windows to restrict write access.
// This is a security measure to prevent unauthorized modification of the blocklist files.
// It uses the `icacls` command to:
// 1. Disable inheritance of ACLs from the parent directory (`/inheritance:d`).
// 2. Grant the current user write permissions (`/grant:r`).
// 3. Remove the `Everyone` group's permissions, effectively locking the file to the current user.
func platformLock(path string) error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	return exec.Command("icacls", path, "/inheritance:d", "/grant:r", currentUser.Username+":(W)", "/remove", "Everyone").Run()
}
