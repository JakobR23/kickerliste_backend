package entity

import (
	"time"
)

type Team struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type TeamWithMembers struct {
	Team
	Members []User `json:"members"`
}
