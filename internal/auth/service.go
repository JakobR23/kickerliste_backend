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

// ErrPasswordMismatch is returned when the supplied current password does not
// match the one stored for the user during a change-password request.
var ErrPasswordMismatch = errors.New("current password is incorrect")

// Service handles login, registration, and credential management.
type Service interface {
	Login(ctx context.Context, req LoginRequest) (string, error)
	Register(ctx context.Context, req RegisterRequest) (string, error)
	ChangePassword(ctx context.Context, userId int, req ChangePasswordRequest) (string, error)
}

// LoginRequest holds the credentials supplied by the client.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest holds the fields required for self-registration.
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=1,max=255"`
	Password string `json:"password" binding:"required,min=8"`
}

// ChangePasswordRequest holds the current and desired new password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword"     binding:"required,min=8"`
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

// Register creates a new user account and returns a signed JWT.
// The account is created with force_password_change = false because the user
// chose their own password during registration.
func (s *service) Register(ctx context.Context, req RegisterRequest) (string, error) {
	u, err := s.userRepo.Create(ctx, req.Username, req.Password, false)
	if err != nil {
		return "", err
	}
	return generateToken(u, s.secret)
}

// ChangePassword verifies the current password, updates it to the new value,
// and returns a fresh JWT with forcePasswordChange = false.
func (s *service) ChangePassword(ctx context.Context, userId int, req ChangePasswordRequest) (string, error) {
	u, err := s.userRepo.GetByIdWithCredentials(ctx, userId)
	if err != nil {
		return "", err
	}

	if hash.Password(req.CurrentPassword, u.GetHashsalt()) != u.GetPassword() {
		return "", ErrPasswordMismatch
	}

	if err := s.userRepo.UpdatePassword(ctx, userId, req.NewPassword); err != nil {
		return "", err
	}

	// Build the updated user locally to avoid an extra round-trip.
	u.ForcePasswordChange = false
	return generateToken(u, s.secret)
}
