package cli

import (
	"context"
	"errors"
	"time"

	urfavecli "github.com/urfave/cli/v3"
	vekconfig "github.com/vekio/config"
	appconfig "github.com/vekio/cspeek/internal/config"
)

func newLiveCommand(
	configFile *vekconfig.ConfigFile[appconfig.Config],
	newClient liquipediaClientFactory,
	now func() time.Time,
) *urfavecli.Command {
	return &urfavecli.Command{
		Name:  "live",
		Usage: "list matches expected to be in progress",
		Flags: []urfavecli.Flag{
			tierFlag(),
			includeQualifiersFlag(),
			limitFlag(),
			offsetFlag(),
			jsonFlag(),
		},
		Action: func(ctx context.Context, command *urfavecli.Command) error {
			if command.NArg() != 0 {
				return errors.New("live does not accept positional arguments")
			}
			service, err := loadMatchService(configFile, newClient)
			if err != nil {
				return err
			}
			result, err := service.Live(ctx, now(), matchFilter(command))
			if err != nil {
				return err
			}
			return writeMatchList(command, result)
		},
	}
}
