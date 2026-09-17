package config

// Config configures the local Overmind application.
//
// A future API process can define its own process configuration without
// coupling the current CLI to server concerns.
type Config struct {
	DataDir string `yaml:"dataDir"`
}
