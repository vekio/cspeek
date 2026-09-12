package cli

import (
	"errors"
	"strings"

	urfavecli "github.com/urfave/cli/v3"
	vekconfig "github.com/vekio/config"
	configurfave "github.com/vekio/config/urfave"
	appconfig "github.com/vekio/cspeek/internal/config"
)

const defaultTier = "curated"

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
	if _, ok := tierCondition(value); !ok {
		return errors.New("tier must be curated, S, A, B, or C")
	}
	return nil
}

func tierCondition(value string) (string, bool) {
	switch strings.ToUpper(value) {
	case "CURATED":
		return "[[extradata_featured::1]]", true
	case "S":
		return "[[liquipediatier::1]]", true
	case "A":
		return "[[liquipediatier::2]]", true
	case "B":
		return "[[liquipediatier::3]]", true
	case "C":
		return "[[liquipediatier::4]]", true
	default:
		return "", false
	}
}
