package auth

import (
	"context"
	"errors"

	"bierliste_backend/internal/hash"
	"bierliste_backend/internal/user"
)

// ErrInvalidCredentials is returned for both unknown username and wrong password
// so callers cannot distinguish the two (prevents user-enumeration attacks).
var ErrInvalidCredentials = errors.New("invalid username or password")

// Service handles login and token issuance.
type Service interface {
	Login(ctx context.Context, req LoginRequest) (string, error)
}

// LoginRequest holds the credentials supplied by the client.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type service struct {
	userRepo *user.Repository
	secret   string
}

// NewService creates an auth Service.
func NewService(userRepo *user.Repository, secret string) Service {
	return &service{userRepo: userRepo, secret: secret}
}

// Login verifies the credentials and returns a signed JWT on success.
func (s *service) Login(ctx context.Context, req LoginRequest) (string, error) {
	u, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		// Return a generic error so callers cannot tell whether the user exists.
		return "", ErrInvalidCredentials
	}

	if hash.Password(req.Password, u.GetHashsalt()) != u.GetPassword() {
		return "", ErrInvalidCredentials
	}

	return generateToken(u, s.secret)
}
