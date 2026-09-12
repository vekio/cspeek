package cli

import (
	"context"
	"fmt"

	vekconfig "github.com/vekio/config"
	appconfig "github.com/vekio/cspeek/internal/config"
	"github.com/vekio/cspeek/pkg/liquipedia"
)

type liquipediaClient interface {
	Matches(context.Context, liquipedia.MatchesOptions) (liquipedia.MatchesPage, error)
}

type liquipediaClientFactory func(appconfig.Config) (liquipediaClient, error)

func newLiquipediaClient(config appconfig.Config) (liquipediaClient, error) {
	return liquipedia.NewClient(liquipedia.Config{APIKey: config.APIKey})
}

func loadLiquipediaClient(
	configFile *vekconfig.ConfigFile[appconfig.Config],
	newClient liquipediaClientFactory,
) (liquipediaClient, error) {
	config, err := configFile.Load()
	if err != nil {
		return nil, fmt.Errorf("load configuration: %w", err)
	}
	return newClient(config)
}
