package liquipedia

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"time"
)

// Match contains the fields needed for schedules, results, and match details.
// Optional, irregular metadata stays as raw JSON, including empty arrays.
// Projected-out fields have their zero value; request state fields explicitly
// when classifying matches, or leave MatchesOptions.Fields empty.
type Match struct {
	ObjectName    string         `json:"objectname"`
	MatchID       string         `json:"match2id"`
	PageName      string         `json:"pagename"`
	Game          string         `json:"game"`
	Date          time.Time      `json:"date"`
	DateExact     bool           `json:"dateexact"`
	Finished      bool           `json:"finished"`
	Status        string         `json:"status"`
	Winner        string         `json:"winner"`
	Walkover      string         `json:"walkover"`
	ResultType    string         `json:"resulttype"`
	BestOf        int            `json:"bestof"`
	Tournament    string         `json:"tournament"`
	TickerName    string         `json:"tickername"`
	Tier          string         `json:"liquipediatier"`
	TierType      string         `json:"liquipediatiertype"`
	PublisherTier string         `json:"publishertier"`
	Section       string         `json:"section"`
	Opponents     []Opponent     `json:"match2opponents"`
	Games         []Game         `json:"match2games"`
	ExtraData     jsontext.Value `json:"extradata"`
	Streams       jsontext.Value `json:"stream"`
	Links         jsontext.Value `json:"links"`
}

type Opponent struct {
	ID           int            `json:"id"`
	Type         string         `json:"type"`
	Name         string         `json:"name"`
	Score        *int           `json:"score"` // nil is absent; -1 is an unknown score, not zero.
	Status       string         `json:"status"`
	TeamTemplate *TeamTemplate  `json:"teamtemplate"`
	Players      jsontext.Value `json:"match2players"`
}

type TeamTemplate struct {
	Name      string `json:"name"`
	ShortName string `json:"shortname"`
}

type Game struct {
	ID        int            `json:"match2gameid"`
	Map       string         `json:"map"`
	Scores    []int          `json:"scores"`
	Winner    string         `json:"winner"`
	Status    string         `json:"status"`
	ExtraData jsontext.Value `json:"extradata"`
}

// UnmarshalJSON normalizes API dates and integer booleans without leaking those
// wire-format details into the application's match model.
func (m *Match) UnmarshalJSON(data []byte) error {
	type matchAlias Match
	var decoded matchAlias
	wire := struct {
		*matchAlias
		Date      string  `json:"date"`
		Finished  apiBool `json:"finished"`
		DateExact apiBool `json:"dateexact"`
	}{matchAlias: &decoded}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	if wire.Date != "" {
		date, err := time.Parse("2006-01-02 15:04:05", wire.Date)
		if err != nil {
			return fmt.Errorf("invalid match date: %w", err)
		}
		decoded.Date = date
	}
	decoded.Finished = bool(wire.Finished)
	decoded.DateExact = bool(wire.DateExact)
	*m = Match(decoded)
	return nil
}

type apiBool bool

func (b *apiBool) UnmarshalJSON(data []byte) error {
	switch string(bytes.TrimSpace(data)) {
	case "true", "1":
		*b = true
	case "false", "0":
		*b = false
	default:
		return fmt.Errorf("expected a boolean or 0/1, got %s", data)
	}
	return nil
}
