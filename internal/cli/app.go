package cli

import (
	"fmt"

	"git.casta.me/alberto/overmind/internal/bootstrap"
	"git.casta.me/alberto/overmind/internal/config"
	configlib "github.com/vekio/config"
)

func loadApp(configFile *configlib.ConfigFile[config.Config]) (bootstrap.Application, error) {
	cfg, err := configFile.Load()
	if err != nil {
		return bootstrap.Application{}, fmt.Errorf("load configuration: %w", err)
	}

	application, err := bootstrap.New(cfg)
	if err != nil {
		return bootstrap.Application{}, fmt.Errorf("configure application: %w", err)
	}
	return application, nil
}
