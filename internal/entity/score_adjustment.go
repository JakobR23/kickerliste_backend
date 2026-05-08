package entity

import "time"

type ScoreAdjustment struct {
	Id        int       `json:"id"`
	UserId    int       `json:"userId"`
	Amount    int       `json:"amount"`
	Reason    string    `json:"reason"`
	CreatedBy int       `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
}
