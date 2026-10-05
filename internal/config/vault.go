package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// NormalizeVaultPath expands the home shorthand, resolves relative paths and rejects files.
func NormalizeVaultPath(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("vault directory is required")
	}
	if value == "~" || strings.HasPrefix(value, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		if value == "~" {
			value = home
		} else {
			value = filepath.Join(home, strings.TrimPrefix(value, "~/"))
		}
	} else if strings.HasPrefix(value, "~") {
		return "", fmt.Errorf("vault directory cannot use another user's home")
	}
	path, err := filepath.Abs(value)
	if err != nil {
		return "", fmt.Errorf("resolve vault directory: %w", err)
	}
	if info, err := os.Stat(path); err == nil {
		if !info.IsDir() {
			return "", fmt.Errorf("vault path %q is not a directory", path)
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("inspect vault directory: %w", err)
	}
	return path, nil
}
