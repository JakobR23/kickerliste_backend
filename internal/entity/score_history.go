package entity

import "time"

// Score-history event types.
const (
	ScoreHistoryTypeFixture    = "fixture"
	ScoreHistoryTypeAdjustment = "adjustment"
)

// ScoreHistoryEntry is one score-changing event in a user's history: a read
// model projected from approved fixtures and manual score adjustments, not a
// stored table. It is a discriminated union keyed by Type — a "fixture" entry
// carries FixtureId (and omits Reason); an "adjustment" entry carries Reason
// (and omits FixtureId).
type ScoreHistoryEntry struct {
	At           time.Time `json:"at"`
	Delta        float64   `json:"delta"`
	RunningTotal float64   `json:"runningTotal"`
	Type         string    `json:"type"`
	FixtureId    *int      `json:"fixtureId,omitempty"`
	Reason       *string   `json:"reason,omitempty"`
}
