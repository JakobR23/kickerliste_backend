package entity

import (
	"time"
)

type result string

const (
	Team1 result = "team_1"
	Team2 result = "team_2"
	Draw  result = "draw"
)

type Fixture struct {
	Id         int       `json:"id"`
	Team1Id    int       `json:"team1Id"`
	Team2Id    int       `json:"team2Id"`
	Result     result    `json:"result"`
	Team1Score *int      `json:"team1Score"`
	Team2Score *int      `json:"team2Score"`
	PlayedAt   time.Time `json:"playedAt"`
	Value      int       `json:"value"`
}
