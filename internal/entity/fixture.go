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

type Fixture struct {
	Id         int         `json:"id"`
	Team1Id    int         `json:"team1Id"`
	Team2Id    int         `json:"team2Id"`
	Result     MatchResult `json:"result"`
	Team1Score *int      `json:"team1Score"`
	Team2Score *int      `json:"team2Score"`
	PlayedAt   time.Time `json:"playedAt"`
	Value      int       `json:"value"`
}
