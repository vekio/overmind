package config

// CLIConfig configures the local Overmind command-line application.
type CLIConfig struct {
	Vault   Vault   `yaml:"vault"`
	Logging Logging `yaml:"logging"`
	Index   Index   `yaml:"index"`
}

// ServerConfig configures the Overmind HTTP server process.
type ServerConfig struct {
	Vault   Vault   `yaml:"vault"`
	Logging Logging `yaml:"logging"`
	HTTP    HTTP    `yaml:"http"`
	Index   Index   `yaml:"index"`
}

// Vault configures the source-of-truth document storage.
type Vault struct {
	Driver   string `yaml:"driver"`
	RootPath string `yaml:"localRoot"`
}

// Logging configures application log filtering.
type Logging struct {
	Level string `yaml:"level"`
}

// HTTP configures the embedded HTTP server.
type HTTP struct {
	Address string `yaml:"address"`
}

// Index configures the derived document read model.
type Index struct {
	Driver string `yaml:"driver"`
	Path   string `yaml:"path"`
}
