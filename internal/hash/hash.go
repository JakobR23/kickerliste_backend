package hash

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = bcrypt.DefaultCost // 10 rounds

// HashPassword hashes a plaintext password using bcrypt and returns the hash.
// The salt is embedded in the returned string — no separate salt is needed.
func HashPassword(plaintext string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CheckPassword reports whether plaintext matches the given bcrypt hash.
func CheckPassword(plaintext, hashed string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plaintext)) == nil
}

// IsLegacy reports whether hashed is a legacy SHA-256 hex hash rather than a
// bcrypt hash. Bcrypt hashes always start with "$2"; SHA-256 hashes are 64-char
// lowercase hex strings and never start with "$".
func IsLegacy(hashed string) bool {
	return !strings.HasPrefix(hashed, "$2")
}

// --- Legacy functions kept for verifying old SHA-256 hashes during migration ---

// GenerateSalt returns a random 16-byte hex-encoded salt.
// Deprecated: new passwords use bcrypt which embeds its own salt.
func GenerateSalt() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Password hashes a plaintext password combined with the given salt using SHA-256.
// Deprecated: use HashPassword instead. Kept only for verifying legacy hashes
// during the login-time migration to bcrypt.
func Password(plaintext, salt string) string {
	h := sha256.Sum256([]byte(plaintext + salt))
	return hex.EncodeToString(h[:])
}
