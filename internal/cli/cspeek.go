package cli

import (
	"time"

	urfavecli "github.com/urfave/cli/v3"
	"github.com/vekio/config"
	configurfave "github.com/vekio/config/urfave"
	appconfig "github.com/vekio/cspeek/internal/config"
)

// New creates the root CSpeek command.
func New() (*urfavecli.Command, error) {
	configFile, err := appconfig.New()
	if err != nil {
		return nil, err
	}
	return newRootCommand(configFile, appconfig.Default(), newLiquipediaClient), nil
}

func newRootCommand(
	configFile *config.ConfigFile[appconfig.Config],
	defaults appconfig.Config,
	newClient liquipediaClientFactory,
) *urfavecli.Command {
	return &urfavecli.Command{
		Name:  "cspeek",
		Usage: "check Counter-Strike matches from the terminal",
		Flags: []urfavecli.Flag{
			configFlag(configFile),
		},
		Commands: []*urfavecli.Command{
			configurfave.NewConfigCommand(configFile, defaults),
			newLiveCommand(configFile, newClient, time.Now),
			newUpcomingCommand(configFile, newClient),
		},
	}
}
