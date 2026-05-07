package entity

type TeamMember struct {
	Id     int `json:"id"`
	TeamId int `json:"teamId"`
	UserId int `json:"userId"`
}
