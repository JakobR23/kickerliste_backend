package auth

import (
	"testing"

	"bierliste_backend/internal/entity"
	"bierliste_backend/internal/hash"
)

// TestVerifyPassword guards the change-password / login verification path for
// both hashing schemes. The bcrypt cases are the regression guard for the bug
// where ChangePassword only verified legacy SHA-256 hashes and so always
// rejected the correct current password for bcrypt accounts.
func TestVerifyPassword(t *testing.T) {
	bcryptHash, err := hash.HashPassword("correct-horse")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	// bcrypt accounts store an empty hashsalt sentinel.
	bcryptUser := entity.NewUser(1, "alice", entity.RoleUser, bcryptHash, "", 0, false, true)

	// Legacy SHA-256 account: password = sha256(plaintext + salt), salt stored.
	const legacySalt = "deadbeef"
	legacyUser := entity.NewUser(2, "bob", entity.RoleUser,
		hash.Password("correct-horse", legacySalt), legacySalt, 0, false, true)

	tests := []struct {
		name      string
		user      entity.User
		plaintext string
		want      bool
	}{
		{"bcrypt correct password", bcryptUser, "correct-horse", true},
		{"bcrypt wrong password", bcryptUser, "wrong", false},
		{"legacy correct password", legacyUser, "correct-horse", true},
		{"legacy wrong password", legacyUser, "wrong", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := verifyPassword(tt.plaintext, tt.user); got != tt.want {
				t.Errorf("verifyPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}
