package main

import (
	"context"
	"fmt"
	"os"

	cspeekcli "github.com/vekio/cspeek/internal/cli"
)

func main() {
	command, err := cspeekcli.New()
	if err == nil {
		err = command.Run(context.Background(), os.Args)
	}
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
