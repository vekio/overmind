package config

// Config configures the local Overmind application.
//
// A future API process can define its own process configuration without
// coupling the current CLI to server concerns.
type Config struct {
	Vault Vault `yaml:"vault"`
	Index Index `yaml:"index"`
}

// Vault configures the source-of-truth document storage.
type Vault struct {
	Driver   string `yaml:"driver"`
	RootPath string `yaml:"localRoot"`
}

// Index configures the derived document read model.
type Index struct {
	Driver string `yaml:"driver"`
	Path   string `yaml:"path"`
}
