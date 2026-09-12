package cli

import (
	"context"
	"errors"

	urfavecli "github.com/urfave/cli/v3"
	vekconfig "github.com/vekio/config"
	appconfig "github.com/vekio/cspeek/internal/config"
)

func newUpcomingCommand(
	configFile *vekconfig.ConfigFile[appconfig.Config],
	newClient liquipediaClientFactory,
) *urfavecli.Command {
	return &urfavecli.Command{
		Name:  "upcoming",
		Usage: "list ongoing and scheduled matches",
		Flags: []urfavecli.Flag{
			tierFlag(),
			includeQualifiersFlag(),
			limitFlag(),
			offsetFlag(),
			jsonFlag(),
		},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			if command.NArg() != 0 {
				return errors.New("upcoming does not accept positional arguments")
			}
			service, err := loadMatchService(configFile, newClient)
			if err != nil {
				return err
			}
			result, err := service.Upcoming(ctx, matchFilter(command))
			if err != nil {
				return err
			}
			return writeMatchList(command, result)
		},
	}
}
