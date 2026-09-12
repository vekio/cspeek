package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	urfavecli "github.com/urfave/cli/v3"
	vekconfig "github.com/vekio/config"
	appconfig "github.com/vekio/cspeek/internal/config"
	"github.com/vekio/cspeek/pkg/liquipedia"
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
			client, err := loadLiquipediaClient(configFile, newClient)
			if err != nil {
				return err
			}
			return runMatchList(
				ctx,
				command,
				client,
				liveMatchesOptions(command, now().UTC()),
			)
		},
	}
}

func liveMatchesOptions(command *urfavecli.Command, now time.Time) liquipedia.MatchesOptions {
	formattedNow := formatAPITime(now)
	conditions := []string{
		"[[game::cs2]]",
		"[[finished::0]]",
		"[[dateexact::1]]",
		fmt.Sprintf("([[date::<%s]] OR [[date::%s]])", formattedNow, formattedNow),
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

func formatAPITime(value time.Time) string {
	return value.UTC().Format("2006-01-02 15:04:05")
}
