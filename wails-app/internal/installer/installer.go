package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/windows/registry"
)

const (
	InstallDirName = "ProcGuard"
	ExeName        = "ProcGuard.exe"
)

// Install performs the full installation process
// 1. Copies executable to Program Files
// 2. Adds Windows Defender exclusion
// 3. Sets up Registry Run key
// 4. Registers Native Messaging Host
func Install() error {
	// 1. Get Program Files path
	programFiles := os.Getenv("ProgramFiles")
	if programFiles == "" {
		return fmt.Errorf("could not find Program Files directory")
	}

	installDir := filepath.Join(programFiles, InstallDirName)
	exePath := filepath.Join(installDir, ExeName)

	// 2. Create directory
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %v", err)
	}

	// 3. Copy executable
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %v", err)
	}

	// Only copy if we're not already running from the install directory
	if !strings.EqualFold(currentExe, exePath) {
		if err := copyFile(currentExe, exePath); err != nil {
			return fmt.Errorf("failed to copy executable: %v", err)
		}
	}

	// 4. Add Windows Defender Exclusion (Powershell)
	// This prevents "virus" false positives for our self-protection code
	cmd := exec.Command("powershell", "-Command", 
		fmt.Sprintf("Add-MpPreference -ExclusionPath '%s'", installDir))
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Run(); err != nil {
		// Log warning but continue - maybe Defender isn't running
		fmt.Printf("Warning: Failed to add Defender exclusion: %v\n", err)
	}

	// 5. Add to Registry Run Key (Auto-start for ALL users)
	k, _, err := registry.CreateKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open Run registry key: %v", err)
	}
	defer k.Close()

	if err := k.SetStringValue("ProcGuard", exePath); err != nil {
		return fmt.Errorf("failed to set Run registry value: %v", err)
	}

	// 6. Create Start Menu Shortcut (For Searchability)
	// This makes "ProcGuard" appear in Windows Search
	startMenu := filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Windows", "Start Menu", "Programs")
	if err := os.MkdirAll(startMenu, 0755); err == nil {
		shortcutPath := filepath.Join(startMenu, "ProcGuard.lnk")
		_ = createShortcut(exePath, shortcutPath)
	}

	return nil
}

// Uninstall removes the system-wide installation components
// Note: It cannot delete the running executable itself.
func Uninstall() error {
	// 1. Remove Registry Run Key
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.WRITE)
	if err == nil {
		defer k.Close()
		_ = k.DeleteValue("ProcGuard")
	}

	// 2. Remove Start Menu Shortcut
	startMenu := filepath.Join(os.Getenv("ProgramData"), "Microsoft", "Windows", "Start Menu", "Programs")
	_ = os.Remove(filepath.Join(startMenu, "ProcGuard.lnk"))

	// 3. Remove Windows Defender Exclusion
	programFiles := os.Getenv("ProgramFiles")
	if programFiles != "" {
		installDir := filepath.Join(programFiles, InstallDirName)
		cmd := exec.Command("powershell", "-Command", 
			fmt.Sprintf("Remove-MpPreference -ExclusionPath '%s'", installDir))
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		_ = cmd.Run()
	}

	return nil
}

// IsInstalled checks if the application is installed system-wide
func IsInstalled() bool {
	// Check if running from Program Files
	exePath, err := os.Executable()
	if err != nil {
		return false
	}
	
	programFiles := os.Getenv("ProgramFiles")
	if programFiles == "" {
		return false
	}

	// Normalize paths for comparison
	exePath = strings.ToLower(filepath.Clean(exePath))
	installDir := strings.ToLower(filepath.Clean(filepath.Join(programFiles, InstallDirName)))

	return strings.HasPrefix(exePath, installDir)
}

func createShortcut(target, shortcut string) error {
	// Use PowerShell to create shortcut (simpler than COM/CGO)
	script := fmt.Sprintf("$s=(New-Object -COM WScript.Shell).CreateShortcut('%s');$s.TargetPath='%s';$s.Save()", shortcut, target)
	cmd := exec.Command("powershell", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0755)
}

// IsAdmin checks if the current process has administrator privileges
func IsAdmin() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	return err == nil
}
