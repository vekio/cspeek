package matches

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/vekio/cspeek/pkg/liquipedia"
)

var fields = []string{
	"objectname", "match2id", "pagename", "date", "bestof", "tournament",
	"tickername", "liquipediatier", "liquipediatiertype", "section",
	"match2opponents",
}

// Source supplies raw match pages to the application service.
type Source interface {
	Matches(context.Context, liquipedia.MatchesOptions) (liquipedia.MatchesPage, error)
}

// Service implements match-list use cases independently of their presentation.
type Service struct {
	source Source
}

// NewService creates a match service backed by a Liquipedia source.
func NewService(source Source) *Service {
	return &Service{source: source}
}

// Live returns unfinished matches with an exact start at or before now.
func (service *Service) Live(ctx context.Context, now time.Time, filter Filter) (Result, error) {
	formattedNow := now.UTC().Format("2006-01-02 15:04:05")
	conditions := []string{
		"[[game::cs2]]",
		"[[finished::0]]",
		"[[dateexact::1]]",
		fmt.Sprintf("([[date::<%s]] OR [[date::%s]])", formattedNow, formattedNow),
	}
	return service.list(ctx, conditions, filter)
}

// Upcoming returns unfinished matches with exact start times and no date cutoff.
func (service *Service) Upcoming(ctx context.Context, filter Filter) (Result, error) {
	return service.list(ctx, []string{
		"[[game::cs2]]",
		"[[finished::0]]",
		"[[dateexact::1]]",
	}, filter)
}

func (service *Service) list(ctx context.Context, conditions []string, filter Filter) (Result, error) {
	if err := filter.validate(); err != nil {
		return Result{}, err
	}
	conditions = append(conditions, filter.Tier.condition())
	if !filter.IncludeQualifiers {
		conditions = append(conditions, "[[liquipediatiertype::!Qualifier]]")
	}
	page, err := service.source.Matches(ctx, liquipedia.MatchesOptions{
		Wiki:       "counterstrike",
		Conditions: strings.Join(conditions, " AND "),
		Fields:     fields,
		Order:      "date ASC,objectname ASC",
		Limit:      filter.Limit,
		Offset:     filter.Offset,
	})
	if err != nil {
		return Result{}, err
	}
	return Result{
		Matches:  mapMatches(uniqueMatches(page.Matches)),
		Warnings: append([]string(nil), page.Warnings...),
	}, nil
}

func uniqueMatches(matches []liquipedia.Match) []liquipedia.Match {
	unique := make([]liquipedia.Match, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		if match.ObjectName == "" {
			unique = append(unique, match)
			continue
		}
		if _, exists := seen[match.ObjectName]; exists {
			continue
		}
		seen[match.ObjectName] = struct{}{}
		unique = append(unique, match)
	}
	return unique
}

func mapMatches(source []liquipedia.Match) []Match {
	matches := make([]Match, 0, len(source))
	for _, sourceMatch := range source {
		opponents := make([]Opponent, 0, len(sourceMatch.Opponents))
		for _, sourceOpponent := range sourceMatch.Opponents {
			name := strings.TrimSpace(sourceOpponent.Name)
			if sourceOpponent.TeamTemplate != nil && strings.TrimSpace(sourceOpponent.TeamTemplate.Name) != "" {
				name = strings.TrimSpace(sourceOpponent.TeamTemplate.Name)
			}
			if name == "" {
				name = "TBD"
			}
			opponents = append(opponents, Opponent{Name: name, Score: cloneScore(sourceOpponent.Score)})
		}
		matches = append(matches, Match{
			ID:         sourceMatch.ObjectName,
			MatchID:    sourceMatch.MatchID,
			PageName:   sourceMatch.PageName,
			Start:      sourceMatch.Date,
			Opponents:  opponents,
			BestOf:     sourceMatch.BestOf,
			Tournament: sourceMatch.Tournament,
			TickerName: sourceMatch.TickerName,
			Tier:       sourceMatch.Tier,
			TierType:   sourceMatch.TierType,
			Section:    sourceMatch.Section,
		})
	}
	return matches
}

func cloneScore(score *int) *int {
	if score == nil {
		return nil
	}
	value := *score
	return &value
}
