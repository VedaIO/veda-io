package data

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config defines the structure of the application's configuration file.
// It is used to store state and settings that need to persist between runs.
type Config struct {
	// AutostartEnabled tracks whether the Windows autostart task has been created.
	AutostartEnabled bool `json:"autostart_enabled,omitempty"`
	// PasswordHash stores the bcrypt hash of the GUI password.
	PasswordHash string `json:"password_hash,omitempty"`
}

// NewConfig creates a new Config with default values.
func NewConfig() *Config {
	return &Config{}
}

// GetConfigPath returns the path to the configuration file, which is stored in the user's cache directory.
func GetConfigPath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, "procguard", "config", "settings.json"), nil
}

// LoadConfig reads the configuration file from the user's cache directory.
// If the file doesn't exist, it returns a new default configuration, making the application resilient.
func LoadConfig() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// If the config file doesn't exist, create a new one with default values.
		// This makes the application more robust on first run or if the file is deleted.
		return NewConfig(), nil
	}
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(content, &config); err != nil {
		// If the file is corrupted or invalid, it's better to return an error
		// than to proceed with a potentially broken configuration.
		return nil, err
	}

	return &config, nil
}

// Save writes the current configuration to the configuration file.
func (c *Config) Save() error {
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// Marshal the config to JSON with indentation for readability.
	content, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// Write the content to the file with read/write permissions for the current user only.
	return os.WriteFile(path, content, 0644)
}
