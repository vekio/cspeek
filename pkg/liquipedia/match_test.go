package liquipedia

import (
	"encoding/json/v2"
	"testing"
	"time"
)

func TestMatchWireVariants(t *testing.T) {
	for _, body := range []string{
		`{"finished":0,"dateexact":1,"stream":[],"links":[]}`,
		`{"finished":false,"dateexact":true,"stream":{"twitch":"channel"},"links":{"hltv":{}}}`,
	} {
		var match Match
		if err := json.Unmarshal([]byte(body), &match); err != nil {
			t.Fatal(err)
		}
		if match.Finished || !match.DateExact || len(match.Streams) == 0 || len(match.Links) == 0 {
			t.Fatalf("unexpected decoded match: %+v", match)
		}
	}
	var match Match
	err := json.Unmarshal([]byte(`{"date":"2026-09-11 08:00:00","finished":1,"dateexact":0,"match2opponents":[{"name":"Alpha","score":-1},{"name":"Beta","score":0},{"name":"TBD"}],"match2games":[{"map":"Inferno","scores":[13,11]},{"map":"Nuke","scores":[]}]}`), &match)
	if err != nil {
		t.Fatal(err)
	}
	if !match.Finished || match.DateExact || match.Date.Location() != time.UTC {
		t.Fatalf("incorrect state/date: %+v", match)
	}
	if *match.Opponents[0].Score != -1 || *match.Opponents[1].Score != 0 || match.Opponents[2].Score != nil {
		t.Fatal("unknown, zero and absent scores must remain distinct")
	}
	if match.Games[0].Scores[0] != 13 || len(match.Games[1].Scores) != 0 {
		t.Fatal("map scores changed")
	}
	// Reusing a value must not retain fields from a previous response.
	if err := json.Unmarshal([]byte(`{}`), &match); err != nil || !match.Date.IsZero() || match.Finished {
		t.Fatalf("stale fields after decoding: %+v, %v", match, err)
	}
}

func TestMatchRejectsInvalidValues(t *testing.T) {
	for _, body := range []string{
		`{"finished":2}`, `{"dateexact":"true"}`, `{"date":"not a date"}`, `{"finished":null}`,
	} {
		var match Match
		if err := json.Unmarshal([]byte(body), &match); err == nil {
			t.Errorf("accepted malformed record: %s", body)
		}
	}
}
