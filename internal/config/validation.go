package config

import "fmt"

// Validate checks that the local application settings are usable.
func (config Config) Validate() error {
	if config.DataDir == "" {
		return fmt.Errorf("dataDir is required")
	}
	return nil
}
