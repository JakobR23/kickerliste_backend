package entity

type User struct {
	Id         int    `json:"id"`
	Username   string `json:"username"`
	password   string
	hashsalt   string
	TotalScore int `json:"totalScore"`
}

func (user User) GetPassword() string {
	return user.password
}

func (user User) GetHashsalt() string {
	return user.hashsalt
}
