package entity

// Role represents the access level of a user.
type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type User struct {
	Id         int    `json:"id"`
	Username   string `json:"username"`
	Role       Role   `json:"role"`
	TotalScore int    `json:"totalScore"`
	password   string
	hashsalt   string
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
func NewUser(id int, username string, role Role, password, hashsalt string, totalScore int) User {
	return User{
		Id:         id,
		Username:   username,
		Role:       role,
		password:   password,
		hashsalt:   hashsalt,
		TotalScore: totalScore,
	}
}
