package cli

import (
	"context"

	urfavecli "github.com/urfave/cli/v3"
	"github.com/vekio/cspeek/pkg/liquipedia"
)

var matchFields = []string{
	"objectname", "match2id", "pagename", "game", "date", "dateexact",
	"finished", "status", "winner", "walkover", "resulttype", "bestof",
	"tournament", "tickername", "liquipediatier", "liquipediatiertype",
	"section", "extradata", "match2opponents", "match2games", "stream", "links",
}

// runMatchList performs the behavior shared by filtered match commands.
func runMatchList(
	ctx context.Context,
	command *urfavecli.Command,
	client liquipediaClient,
	options liquipedia.MatchesOptions,
) error {
	page, err := client.Matches(ctx, options)
	if err != nil {
		return err
	}
	if err := writeWarnings(command.ErrWriter, page.Warnings); err != nil {
		return err
	}
	matches := newMatchDTOs(uniqueMatches(page.Matches))
	return writeMatchOutput(command.Writer, matches, command.Bool("json"))
}
