package cli

import (
	"errors"

	urfavecli "github.com/urfave/cli/v3"
	vekconfig "github.com/vekio/config"
	configurfave "github.com/vekio/config/urfave"
	appconfig "github.com/vekio/cspeek/internal/config"
	"github.com/vekio/cspeek/internal/matches"
)

const defaultTier = string(matches.TierCurated)

func configFlag(file *vekconfig.ConfigFile[appconfig.Config]) urfavecli.Flag {
	return configurfave.NewConfigFlag(file)
}

func tierFlag() urfavecli.Flag {
	return &urfavecli.StringFlag{
		Name:  "tier",
		Usage: "match selection: curated, S, A, B, or C",
		Value: defaultTier,
		Config: urfavecli.StringConfig{
			TrimSpace: true,
		},
		Validator: validateTier,
	}
}

func includeQualifiersFlag() urfavecli.Flag {
	return &urfavecli.BoolFlag{
		Name:  "include-qualifiers",
		Usage: "include qualifier events in the selected tier",
	}
}

func limitFlag() urfavecli.Flag {
	return &urfavecli.IntFlag{
		Name:  "limit",
		Usage: "number of API records to request",
		Value: 100,
		Validator: func(value int) error {
			if value < 1 || value > 1000 {
				return errors.New("limit must be between 1 and 1000")
			}
			return nil
		},
	}
}

func offsetFlag() urfavecli.Flag {
	return &urfavecli.IntFlag{
		Name:  "offset",
		Usage: "number of API records to skip",
		Validator: func(value int) error {
			if value < 0 {
				return errors.New("offset must be nonnegative")
			}
			return nil
		},
	}
}

func jsonFlag() urfavecli.Flag {
	return &urfavecli.BoolFlag{
		Name:  "json",
		Usage: "write the matches as JSON",
	}
}

func validateTier(value string) error {
	_, err := matches.ParseTier(value)
	return err
}
