//go:build darwin

package nativehost

// InstallNativeHost is a dummy implementation for non-Windows platforms.
func InstallNativeHost(exePath, extensionId string) error {
	return nil
}

// CreateManifest is a dummy implementation for non-Windows platforms.
func CreateManifest(manifestPath, exePath, extensionId string) error {
	return nil
}

// RegisterExtension is a dummy implementation for non-Windows platforms.
func RegisterExtension(extensionId string) error {
	return nil
}

// Remove is a dummy implementation for non-Windows platforms.
func Remove() error {
	return nil
}
