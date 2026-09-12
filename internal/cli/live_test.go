package cli

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	vekconfig "github.com/vekio/config"
	appconfig "github.com/vekio/cspeek/internal/config"
	matchservice "github.com/vekio/cspeek/internal/matches"
	"github.com/vekio/cspeek/pkg/liquipedia"
)

type fakeMatchesClient struct {
	page    liquipedia.MatchesPage
	err     error
	options liquipedia.MatchesOptions
}

func (f *fakeMatchesClient) Matches(_ context.Context, options liquipedia.MatchesOptions) (liquipedia.MatchesPage, error) {
	f.options = options
	return f.page, f.err
}

func assertMatchTable(t *testing.T, output string, values []string) {
	t.Helper()
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("table has %d lines: %q", len(lines), output)
	}
	headers := []string{"DATE", "TEAM 1", "SCORE", "TEAM 2", "FORMAT", "TOURNAMENT"}
	for index, header := range headers {
		headerColumn := strings.Index(lines[0], header)
		valueColumn := strings.Index(lines[1], values[index])
		if headerColumn < 0 || headerColumn != valueColumn {
			t.Errorf("column %q starts at %d; value %q starts at %d:\n%s", header, headerColumn, values[index], valueColumn, output)
		}
	}
	if strings.Contains(output, "\t") {
		t.Errorf("table contains unexpanded tabs: %q", output)
	}
}

func testConfigFile(t *testing.T, config appconfig.Config) *vekconfig.ConfigFile[appconfig.Config] {
	t.Helper()
	file, err := vekconfig.NewYAMLConfigFile[appconfig.Config]("cspeek-test", "config.yml")
	if err != nil {
		t.Fatal(err)
	}
	if err := file.SetPath(t.TempDir() + "/config.yml"); err != nil {
		t.Fatal(err)
	}
	if err := file.Save(config); err != nil {
		t.Fatal(err)
	}
	return file
}

func TestLivePlainOutputAndQuery(t *testing.T) {
	one, zero := 1, 0
	fake := &fakeMatchesClient{page: liquipedia.MatchesPage{
		Matches: []liquipedia.Match{
			{
				ObjectName: "match-1", MatchID: "m1", PageName: "event/page",
				Date: time.Date(2026, 9, 11, 9, 15, 0, 0, time.UTC), BestOf: 3,
				Tournament: "Full Event", TickerName: "Event", Tier: "1",
				Opponents: []liquipedia.Opponent{
					{Name: "Alpha source", Score: &one, TeamTemplate: &liquipedia.TeamTemplate{Name: "Alpha"}},
					{Name: "Beta", Score: &zero},
				},
			},
			{ObjectName: "match-1"}, // Exact API duplicate is removed before output.
		},
		Warnings: []string{"partial data"},
	}}
	var stdout, stderr bytes.Buffer
	command := newLiveCommand(testConfigFile(t, appconfig.Config{APIKey: "secret"}), func(config appconfig.Config) (matchservice.Source, error) {
		if config.APIKey != "secret" {
			t.Fatalf("API key = %q", config.APIKey)
		}
		return fake, nil
	}, func() time.Time { return time.Date(2026, 9, 11, 14, 30, 0, 0, time.FixedZone("test", 3600)) })
	command.Writer, command.ErrWriter = &stdout, &stderr
	if err := command.Run(context.Background(), []string{"live", "--tier", "C", "--limit", "25", "--offset", "50"}); err != nil {
		t.Fatal(err)
	}
	wantConditions := "[[game::cs2]] AND [[finished::0]] AND [[dateexact::1]] AND ([[date::<2026-09-11 13:30:00]] OR [[date::2026-09-11 13:30:00]]) AND [[liquipediatier::4]] AND [[liquipediatiertype::!Qualifier]]"
	if fake.options.Conditions != wantConditions || fake.options.Limit != 25 || fake.options.Offset != 50 {
		t.Fatalf("unexpected options: %+v", fake.options)
	}
	start := time.Date(2026, 9, 11, 9, 15, 0, 0, time.UTC).In(time.Local).Format(matchTimeFormat)
	assertMatchTable(t, stdout.String(), []string{start, "Alpha", "1-0", "Beta", "BO3", "Event"})
	if stderr.String() != "warning: partial data\n" {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestLiveJSONOutput(t *testing.T) {
	unknown := -1
	fake := &fakeMatchesClient{page: liquipedia.MatchesPage{Matches: []liquipedia.Match{{
		ObjectName: "match-1", MatchID: "m1", Date: time.Date(2026, 9, 11, 8, 0, 0, 0, time.UTC),
		Tournament: "Event", BestOf: 3, Tier: "1",
		Opponents: []liquipedia.Opponent{{Name: "Alpha", Score: &unknown}, {Name: ""}},
	}}}}
	var stdout bytes.Buffer
	command := newLiveCommand(testConfigFile(t, appconfig.Config{APIKey: "secret"}), func(appconfig.Config) (matchservice.Source, error) { return fake, nil }, time.Now)
	command.Writer, command.ErrWriter = &stdout, &bytes.Buffer{}
	if err := command.Run(context.Background(), []string{"live", "--json", "--include-qualifiers"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(fake.options.Conditions, "[[extradata_featured::1]]") || strings.Contains(fake.options.Conditions, "liquipediatier") {
		t.Fatalf("curated selection generated unexpected conditions: %s", fake.options.Conditions)
	}
	var output []matchDTO
	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		t.Fatal(err)
	}
	if len(output) != 1 || output[0].Start.Location() != time.UTC || output[0].Opponents[1].Name != "TBD" {
		t.Fatalf("unexpected JSON: %+v", output)
	}
	if output[0].Opponents[0].Score == nil || *output[0].Opponents[0].Score != -1 || output[0].Opponents[1].Score != nil {
		t.Fatalf("scores lost their source meaning: %+v", output[0].Opponents)
	}
}

func TestLiveErrors(t *testing.T) {
	t.Run("missing API key", func(t *testing.T) {
		file := testConfigFile(t, appconfig.Config{APIKey: "temporary"})
		if err := os.WriteFile(file.Path(), []byte("api_key: \"\"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		command := newLiveCommand(file, func(appconfig.Config) (matchservice.Source, error) {
			t.Fatal("client must not be created")
			return nil, nil
		}, time.Now)
		if err := command.Run(context.Background(), []string{"live"}); err == nil || !strings.Contains(err.Error(), "API key cannot be empty") {
			t.Fatalf("error = %v", err)
		}
	})
	t.Run("client error", func(t *testing.T) {
		want := errors.New("API unavailable")
		command := newLiveCommand(testConfigFile(t, appconfig.Config{APIKey: "secret"}), func(appconfig.Config) (matchservice.Source, error) {
			return &fakeMatchesClient{err: want}, nil
		}, time.Now)
		if err := command.Run(context.Background(), []string{"live"}); !errors.Is(err, want) {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestTierFlag(t *testing.T) {
	for _, value := range []string{"curated", "S", "A", "B", "C"} {
		if err := validateTier(value); err != nil {
			t.Errorf("validateTier(%q) error = %v", value, err)
		}
	}
	if err := validateTier("4"); err == nil {
		t.Fatal("numeric tier was accepted")
	}
}
