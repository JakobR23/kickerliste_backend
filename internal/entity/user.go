package entity

// Role represents the access level of a user.
type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

// User is the public representation of a user. Credential fields (password,
// hashsalt) and the force_password_change flag are unexported so they are
// never accidentally serialised in HTTP responses.
type User struct {
	Id         int     `json:"id"`
	Username   string  `json:"username"`
	Role       Role    `json:"role"`
	TotalScore float64 `json:"totalScore"`
	password   string
	hashsalt   string
	// ForcePasswordChange signals that the user must change their password
	// before performing any other action. Set to true when an admin creates
	// the account; cleared when the user successfully changes their password.
	// json:"-" ensures it is never exposed in HTTP responses.
	ForcePasswordChange bool `json:"-"`
}

func (u User) GetPassword() string { return u.password }
func (u User) GetHashsalt() string { return u.hashsalt }

// NewUser constructs a fully-populated User including the credential fields.
// Intended for use by the repository layer when loading rows that include
// password, hashsalt, and force_password_change (e.g. for authentication).
func NewUser(id int, username string, role Role, password, hashsalt string, totalScore float64, forcePasswordChange bool) User {
	return User{
		Id:                  id,
		Username:            username,
		Role:                role,
		password:            password,
		hashsalt:            hashsalt,
		TotalScore:          totalScore,
		ForcePasswordChange: forcePasswordChange,
	}
}
