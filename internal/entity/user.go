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

// NewUser constructs a fully-populated User including the credential fields.
// Intended for use by the repository layer when loading rows that include
// password and hashsalt (e.g. for authentication).
func NewUser(id int, username, password, hashsalt string, totalScore int) User {
	return User{
		Id:         id,
		Username:   username,
		password:   password,
		hashsalt:   hashsalt,
		TotalScore: totalScore,
	}
}
