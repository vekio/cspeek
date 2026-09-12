package cli

import (
	urfavecli "github.com/urfave/cli/v3"
	"github.com/vekio/cspeek/internal/matches"
)

func matchFilter(command *urfavecli.Command) matches.Filter {
	tier, _ := matches.ParseTier(command.String("tier"))
	return matches.Filter{
		Tier:              tier,
		IncludeQualifiers: command.Bool("include-qualifiers"),
		Limit:             command.Int("limit"),
		Offset:            command.Int("offset"),
	}
}

func writeMatchList(command *urfavecli.Command, result matches.Result) error {
	if err := writeWarnings(command.ErrWriter, result.Warnings); err != nil {
		return err
	}
	return writeMatchOutput(command.Writer, newMatchDTOs(result.Matches), command.Bool("json"))
}
