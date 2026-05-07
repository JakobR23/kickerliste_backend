package hash

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateSalt returns a random 16-byte hex-encoded salt.
func GenerateSalt() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Password hashes a plaintext password combined with the given salt using SHA-256.
func Password(plaintext, salt string) string {
	h := sha256.Sum256([]byte(plaintext + salt))
	return hex.EncodeToString(h[:])
}
