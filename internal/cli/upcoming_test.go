package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	appconfig "github.com/vekio/cspeek/internal/config"
	"github.com/vekio/cspeek/pkg/liquipedia"
)

func TestUpcomingUsesSharedMatchOutputAndFilters(t *testing.T) {
	fake := &fakeMatchesClient{page: liquipedia.MatchesPage{Matches: []liquipedia.Match{{
		ObjectName: "upcoming-1",
		Date:       time.Date(2026, 9, 12, 10, 0, 0, 0, time.UTC),
		BestOf:     3,
		Tournament: "Future Event",
		Opponents: []liquipedia.Opponent{
			{Name: "Alpha"},
			{Name: "Beta"},
		},
	}}}}
	command := newUpcomingCommand(
		testConfigFile(t, appconfig.Config{APIKey: "secret"}),
		func(appconfig.Config) (liquipediaClient, error) { return fake, nil },
	)
	var output bytes.Buffer
	command.Writer, command.ErrWriter = &output, &bytes.Buffer{}
	if err := command.Run(context.Background(), []string{"upcoming"}); err != nil {
		t.Fatal(err)
	}
	wantConditions := "[[game::cs2]] AND [[finished::0]] AND [[dateexact::1]] AND [[extradata_featured::1]] AND [[liquipediatiertype::!Qualifier]]"
	if fake.options.Conditions != wantConditions || fake.options.Limit != 100 || fake.options.Offset != 0 {
		t.Fatalf("unexpected options: %+v", fake.options)
	}
	start := time.Date(2026, 9, 12, 10, 0, 0, 0, time.UTC).In(time.Local).Format(matchTimeFormat)
	assertMatchTable(t, output.String(), []string{start, "Alpha", "---", "Beta", "BO3", "Future Event"})
}

func TestUpcomingTierAndQualifierOptions(t *testing.T) {
	fake := &fakeMatchesClient{}
	command := newUpcomingCommand(
		testConfigFile(t, appconfig.Config{APIKey: "secret"}),
		func(appconfig.Config) (liquipediaClient, error) { return fake, nil },
	)
	command.Writer, command.ErrWriter = &bytes.Buffer{}, &bytes.Buffer{}
	if err := command.Run(context.Background(), []string{
		"upcoming", "--tier", "A", "--include-qualifiers", "--limit", "25", "--offset", "50",
	}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(fake.options.Conditions, "liquipediatiertype") ||
		!strings.Contains(fake.options.Conditions, "[[liquipediatier::2]]") ||
		fake.options.Limit != 25 || fake.options.Offset != 50 {
		t.Fatalf("unexpected options: %+v", fake.options)
	}
}

func TestUpcomingRejectsArguments(t *testing.T) {
	command := newUpcomingCommand(
		testConfigFile(t, appconfig.Config{APIKey: "secret"}),
		func(appconfig.Config) (liquipediaClient, error) {
			t.Fatal("client must not be created")
			return nil, nil
		},
	)
	if err := command.Run(context.Background(), []string{"upcoming", "extra"}); err == nil {
		t.Fatal("positional argument was accepted")
	}
}
