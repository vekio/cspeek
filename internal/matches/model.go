// Package matches provides match queries shared by the CLI and TUI.
package matches

import "time"

// Match is the application representation of a Counter-Strike match.
type Match struct {
	ID         string
	MatchID    string
	PageName   string
	Start      time.Time
	Opponents  []Opponent
	BestOf     int
	Tournament string
	TickerName string
	Tier       string
	TierType   string
	Section    string
}

// Opponent contains the display name and series score of one participant.
type Opponent struct {
	Name  string
	Score *int
}

// Result contains one page of unique matches and non-fatal API warnings.
type Result struct {
	Matches  []Match
	Warnings []string
}
