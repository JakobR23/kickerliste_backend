package entity

import (
	"time"
)

type MatchResult string

const (
	Team1 MatchResult = "team_1"
	Team2 MatchResult = "team_2"
	Draw  MatchResult = "draw"
)

type FixtureStatus string

const (
	StatusPending  FixtureStatus = "pending"
	StatusApproved FixtureStatus = "approved"
	StatusRejected FixtureStatus = "rejected"
)

// allFixtureStatuses is the single source of truth for valid FixtureStatus values.
// Add new statuses here and IsValid() will pick them up automatically.
var allFixtureStatuses = []FixtureStatus{StatusPending, StatusApproved, StatusRejected}

// IsValid reports whether s is a known FixtureStatus value.
func (s FixtureStatus) IsValid() bool {
	for _, v := range allFixtureStatuses {
		if s == v {
			return true
		}
	}
	return false
}

type Fixture struct {
	Id          int           `json:"id"`
	Team1Id     int           `json:"team1Id"`
	Team2Id     int           `json:"team2Id"`
	Result      MatchResult   `json:"result"`
	Team1Score  *int          `json:"team1Score"`
	Team2Score  *int          `json:"team2Score"`
	PlayedAt    time.Time     `json:"playedAt"`
	Value       int           `json:"value"`
	Status      FixtureStatus `json:"status"`
	SubmittedBy *int          `json:"submittedBy"`
}
