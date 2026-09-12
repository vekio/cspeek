package matches

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/vekio/cspeek/pkg/liquipedia"
)

type sourceStub struct {
	page    liquipedia.MatchesPage
	err     error
	options liquipedia.MatchesOptions
}

func (source *sourceStub) Matches(
	_ context.Context,
	options liquipedia.MatchesOptions,
) (liquipedia.MatchesPage, error) {
	source.options = options
	return source.page, source.err
}

func TestLiveBuildsQueryAndMapsUniqueMatches(t *testing.T) {
	score := 1
	source := &sourceStub{page: liquipedia.MatchesPage{
		Matches: []liquipedia.Match{
			{
				ObjectName: "match-1",
				Date:       time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC),
				Opponents: []liquipedia.Opponent{{
					Name:         "source name",
					Score:        &score,
					TeamTemplate: &liquipedia.TeamTemplate{Name: "Display Name"},
				}},
			},
			{ObjectName: "match-1"},
		},
		Warnings: []string{"partial result"},
	}}
	service := NewService(source)
	result, err := service.Live(
		context.Background(),
		time.Date(2026, 9, 12, 14, 30, 0, 0, time.FixedZone("test", 3600)),
		Filter{Tier: TierC, Limit: 25, Offset: 50},
	)
	if err != nil {
		t.Fatal(err)
	}
	wantConditions := "[[game::cs2]] AND [[finished::0]] AND [[dateexact::1]] AND ([[date::<2026-09-12 13:30:00]] OR [[date::2026-09-12 13:30:00]]) AND [[liquipediatier::4]] AND [[liquipediatiertype::!Qualifier]]"
	if source.options.Conditions != wantConditions || source.options.Limit != 25 || source.options.Offset != 50 {
		t.Fatalf("unexpected options: %+v", source.options)
	}
	if len(result.Matches) != 1 || result.Matches[0].Opponents[0].Name != "Display Name" {
		t.Fatalf("unexpected matches: %+v", result.Matches)
	}
	if len(result.Warnings) != 1 || result.Warnings[0] != "partial result" {
		t.Fatalf("unexpected warnings: %v", result.Warnings)
	}
}

func TestUpcomingUsesCuratedWithoutDateOrQualifierCondition(t *testing.T) {
	source := &sourceStub{}
	service := NewService(source)
	_, err := service.Upcoming(context.Background(), Filter{
		Tier:              TierCurated,
		IncludeQualifiers: true,
		Limit:             100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if source.options.Conditions != "[[game::cs2]] AND [[finished::0]] AND [[dateexact::1]] AND [[extradata_featured::1]]" {
		t.Fatalf("conditions = %q", source.options.Conditions)
	}
	if strings.Contains(source.options.Conditions, "[[date::") {
		t.Fatalf("upcoming contains a date condition: %s", source.options.Conditions)
	}
}

func TestParseTier(t *testing.T) {
	for input, want := range map[string]Tier{
		"curated": TierCurated,
		"s":       TierS,
		"A":       TierA,
		"B":       TierB,
		"C":       TierC,
	} {
		got, err := ParseTier(input)
		if err != nil || got != want {
			t.Errorf("ParseTier(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
	if _, err := ParseTier("4"); err == nil {
		t.Fatal("numeric tier was accepted")
	}
}
