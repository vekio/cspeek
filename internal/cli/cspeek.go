package cli

import (
	urfavecli "github.com/urfave/cli/v3"
)

// New creates the root CSpeek command.
func New() (*urfavecli.Command, error) {
	return newRootCommand(), nil
}

func newRootCommand() *urfavecli.Command {
	return &urfavecli.Command{
		Name:     "cspeek",
		Usage:    "",
		Commands: []*urfavecli.Command{},
	}
}
