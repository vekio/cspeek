package matches

import (
	"errors"
	"strings"
)

// Tier selects Liquipedia's curated matches or one event tier.
type Tier string

const (
	TierCurated Tier = "curated"
	TierS       Tier = "S"
	TierA       Tier = "A"
	TierB       Tier = "B"
	TierC       Tier = "C"
)

// ParseTier converts a user-facing tier label into a supported Tier.
func ParseTier(value string) (Tier, error) {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "CURATED":
		return TierCurated, nil
	case "S":
		return TierS, nil
	case "A":
		return TierA, nil
	case "B":
		return TierB, nil
	case "C":
		return TierC, nil
	default:
		return "", errors.New("tier must be curated, S, A, B, or C")
	}
}

// Filter contains the options shared by match-list use cases.
type Filter struct {
	Tier              Tier
	IncludeQualifiers bool
	Limit             int
	Offset            int
}

func (filter Filter) validate() error {
	if _, err := ParseTier(string(filter.Tier)); err != nil {
		return err
	}
	if filter.Limit < 0 || filter.Limit > 1000 {
		return errors.New("limit must be between 0 and 1000")
	}
	if filter.Offset < 0 {
		return errors.New("offset must be nonnegative")
	}
	return nil
}

func (tier Tier) condition() string {
	switch tier {
	case TierCurated:
		return "[[extradata_featured::1]]"
	case TierS:
		return "[[liquipediatier::1]]"
	case TierA:
		return "[[liquipediatier::2]]"
	case TierB:
		return "[[liquipediatier::3]]"
	case TierC:
		return "[[liquipediatier::4]]"
	default:
		return ""
	}
}
