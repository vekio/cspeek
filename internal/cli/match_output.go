package cli

import (
	"encoding/json/v2"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/vekio/cspeek/pkg/liquipedia"
)

const (
	matchTimeFormat = "2006-01-02 15:04 MST"
	matchTSVHeader  = "DATE\tTEAM 1\tSCORE\tTEAM 2\tFORMAT\tTOURNAMENT"
)

// matchDTO is the stable match representation exposed by CLI output.
// It keeps the Liquipedia wire model out of output formatting.
type matchDTO struct {
	ID         string             `json:"id"`
	MatchID    string             `json:"match_id"`
	PageName   string             `json:"page_name"`
	Start      time.Time          `json:"start"`
	Opponents  []matchOpponentDTO `json:"opponents"`
	BestOf     int                `json:"best_of"`
	Tournament string             `json:"tournament"`
	Tier       string             `json:"tier"`
	TierType   string             `json:"tier_type"`
	Section    string             `json:"section"`
	tickerName string
}

type matchOpponentDTO struct {
	Name  string `json:"name"`
	Score *int   `json:"score"`
}

func newMatchDTOs(matches []liquipedia.Match) []matchDTO {
	items := make([]matchDTO, 0, len(matches))
	for _, match := range matches {
		items = append(items, newMatchDTO(match))
	}
	return items
}

func newMatchDTO(match liquipedia.Match) matchDTO {
	opponents := make([]matchOpponentDTO, 0, len(match.Opponents))
	for _, opponent := range match.Opponents {
		name := strings.TrimSpace(opponent.Name)
		if opponent.TeamTemplate != nil && strings.TrimSpace(opponent.TeamTemplate.Name) != "" {
			name = strings.TrimSpace(opponent.TeamTemplate.Name)
		}
		if name == "" {
			name = "TBD"
		}
		opponents = append(opponents, matchOpponentDTO{Name: name, Score: opponent.Score})
	}
	return matchDTO{
		ID:         match.ObjectName,
		MatchID:    match.MatchID,
		PageName:   match.PageName,
		Start:      match.Date,
		Opponents:  opponents,
		BestOf:     match.BestOf,
		Tournament: match.Tournament,
		Tier:       match.Tier,
		TierType:   match.TierType,
		Section:    match.Section,
		tickerName: match.TickerName,
	}
}

// TSV returns one record suitable for line-oriented Unix tools.
func (match matchDTO) TSV() string {
	left, right := match.opponent(0), match.opponent(1)
	return strings.Join([]string{
		match.Start.In(time.Local).Format(matchTimeFormat),
		cleanField(left.Name),
		scoreText(left.Score) + "-" + scoreText(right.Score),
		cleanField(right.Name),
		fmt.Sprintf("BO%d", match.BestOf),
		cleanField(match.tournamentName()),
	}, "\t")
}

func (match matchDTO) opponent(index int) matchOpponentDTO {
	if index >= len(match.Opponents) {
		return matchOpponentDTO{Name: "TBD"}
	}
	return match.Opponents[index]
}

func (match matchDTO) tournamentName() string {
	if name := strings.TrimSpace(match.tickerName); name != "" {
		return name
	}
	return match.Tournament
}

func scoreText(score *int) string {
	if score == nil || *score < 0 {
		return "-"
	}
	return strconv.Itoa(*score)
}

func writeMatchOutput(output io.Writer, matches []matchDTO, asJSON bool) error {
	if asJSON {
		data, err := json.Marshal(matches)
		if err != nil {
			return fmt.Errorf("encode matches as JSON: %w", err)
		}
		if _, err := output.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("write matches: %w", err)
		}
		return nil
	}
	table := tabwriter.NewWriter(output, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(table, matchTSVHeader); err != nil {
		return fmt.Errorf("write matches: %w", err)
	}
	for _, match := range matches {
		if _, err := fmt.Fprintln(table, match.TSV()); err != nil {
			return fmt.Errorf("write matches: %w", err)
		}
	}
	if err := table.Flush(); err != nil {
		return fmt.Errorf("write matches: %w", err)
	}
	return nil
}

func writeWarnings(output io.Writer, warnings []string) error {
	for _, warning := range warnings {
		if _, err := fmt.Fprintf(output, "warning: %s\n", warning); err != nil {
			return fmt.Errorf("write warning: %w", err)
		}
	}
	return nil
}

func uniqueMatches(matches []liquipedia.Match) []liquipedia.Match {
	unique := make([]liquipedia.Match, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		key := match.ObjectName
		if key == "" {
			unique = append(unique, match)
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, match)
	}
	return unique
}

func cleanField(value string) string {
	return strings.NewReplacer("\t", " ", "\r", " ", "\n", " ").Replace(value)
}
