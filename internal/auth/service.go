package auth

import (
	"context"
	"errors"

	"bierliste_backend/internal/entity"
	"bierliste_backend/internal/hash"
	"bierliste_backend/internal/user"
)

// ErrInvalidCredentials is returned for both unknown username and wrong password
// so callers cannot distinguish the two (prevents user-enumeration attacks).
var ErrInvalidCredentials = errors.New("invalid username or password")

// ErrPasswordMismatch is returned when the supplied current password does not
// match the one stored for the user during a change-password request.
var ErrPasswordMismatch = errors.New("current password is incorrect")

// ErrAccountNotActive is returned when a user attempts to log in but their
// account has not yet been activated by an admin.
var ErrAccountNotActive = errors.New("account pending activation")

// Service handles login, registration, and credential management.
type Service interface {
	Login(ctx context.Context, req LoginRequest) (string, error)
	// Register creates an inactive account pending admin activation. Returns
	// no token — the user must be activated before they can log in.
	Register(ctx context.Context, req RegisterRequest) error
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
// Returns ErrAccountNotActive if the account exists but has not been activated.
func (s *service) Login(ctx context.Context, req LoginRequest) (string, error) {
	u, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		// Return a generic error so callers cannot tell whether the user exists.
		return "", ErrInvalidCredentials
	}

	if hash.Password(req.Password, u.GetHashsalt()) != u.GetPassword() {
		return "", ErrInvalidCredentials
	}

	if !u.Active {
		return "", ErrAccountNotActive
	}

	return generateToken(u, s.secret)
}

// Register creates an inactive account. No token is returned — the user must
// wait for an admin to activate the account before they can log in.
// force_password_change is false because the user chose their own password.
// Self-registered accounts are always role=user.
func (s *service) Register(ctx context.Context, req RegisterRequest) error {
	_, err := s.userRepo.Create(ctx, req.Username, req.Password, entity.RoleUser, false, false)
	return err
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
