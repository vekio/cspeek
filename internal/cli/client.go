package cli

import (
	"fmt"

	vekconfig "github.com/vekio/config"
	appconfig "github.com/vekio/cspeek/internal/config"
	"github.com/vekio/cspeek/internal/matches"
	"github.com/vekio/cspeek/pkg/liquipedia"
)

type liquipediaClientFactory func(appconfig.Config) (matches.Source, error)

func newLiquipediaClient(config appconfig.Config) (matches.Source, error) {
	return liquipedia.NewClient(liquipedia.Config{APIKey: config.APIKey})
}

func loadMatchService(
	configFile *vekconfig.ConfigFile[appconfig.Config],
	newClient liquipediaClientFactory,
) (*matches.Service, error) {
	config, err := configFile.Load()
	if err != nil {
		return nil, fmt.Errorf("load configuration: %w", err)
	}
	client, err := newClient(config)
	if err != nil {
		return nil, err
	}
	return matches.NewService(client), nil
}
