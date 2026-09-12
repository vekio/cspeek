package cli

import (
	"context"
	"errors"
	"strings"

	urfavecli "github.com/urfave/cli/v3"
	vekconfig "github.com/vekio/config"
	appconfig "github.com/vekio/cspeek/internal/config"
	"github.com/vekio/cspeek/pkg/liquipedia"
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
			client, err := loadLiquipediaClient(configFile, newClient)
			if err != nil {
				return err
			}
			return runMatchList(ctx, command, client, upcomingMatchesOptions(command))
		},
	}
}

func upcomingMatchesOptions(command *urfavecli.Command) liquipedia.MatchesOptions {
	conditions := []string{
		"[[game::cs2]]",
		"[[finished::0]]",
		"[[dateexact::1]]",
	}
	tier, _ := tierCondition(command.String("tier"))
	conditions = append(conditions, tier)
	if !command.Bool("include-qualifiers") {
		conditions = append(conditions, "[[liquipediatiertype::!Qualifier]]")
	}
	return liquipedia.MatchesOptions{
		Wiki:       "counterstrike",
		Conditions: strings.Join(conditions, " AND "),
		Fields:     matchFields,
		Order:      "date ASC,objectname ASC",
		Limit:      command.Int("limit"),
		Offset:     command.Int("offset"),
	}
}
