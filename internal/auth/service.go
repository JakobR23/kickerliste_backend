package auth

import (
	"context"
	"net/http"

	"bierliste_backend/internal/entity"
	"bierliste_backend/internal/hash"
	"bierliste_backend/internal/httputil"
	"bierliste_backend/internal/user"
)

// ErrInvalidCredentials is returned for both unknown username and wrong password
// so callers cannot distinguish the two (prevents user-enumeration attacks).
var ErrInvalidCredentials = httputil.NewStatusError(http.StatusUnauthorized, "invalid username or password")

// ErrPasswordMismatch is returned when the supplied current password does not
// match the one stored for the user during a change-password request.
var ErrPasswordMismatch = httputil.NewStatusError(http.StatusUnprocessableEntity, "current password is incorrect")

// ErrAccountNotActive is returned when a user attempts to log in but their
// account has not yet been activated by an admin.
var ErrAccountNotActive = httputil.NewStatusError(http.StatusForbidden, "account pending activation")

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

// verifyPassword reports whether plaintext matches the user's stored password.
// It handles both bcrypt hashes and legacy SHA-256 hashes (which use the
// separate hashsalt column). Both Login and ChangePassword use this so their
// verification logic cannot drift apart.
func verifyPassword(plaintext string, u entity.User) bool {
	if hash.IsLegacy(u.GetPassword()) {
		return hash.Password(plaintext, u.GetHashsalt()) == u.GetPassword()
	}
	return hash.CheckPassword(plaintext, u.GetPassword())
}

// Login verifies the credentials and returns a signed JWT on success.
// Returns ErrAccountNotActive if the account exists but has not been activated.
//
// Password migration: accounts whose password was hashed with the legacy
// SHA-256 scheme are transparently re-hashed with bcrypt on successful login.
func (s *service) Login(ctx context.Context, req LoginRequest) (string, error) {
	u, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		// Return a generic error so callers cannot tell whether the user exists.
		return "", ErrInvalidCredentials
	}

	if !verifyPassword(req.Password, u) {
		return "", ErrInvalidCredentials
	}
	// Legacy SHA-256 accounts are transparently upgraded to bcrypt on a
	// successful login. Best-effort; if it fails the user can still log in today.
	if hash.IsLegacy(u.GetPassword()) {
		_ = s.userRepo.UpdatePassword(ctx, u.Id, req.Password)
	}

	if !u.Active {
		return "", ErrAccountNotActive
	}

	return generateToken(u, s.secret)
}

// Register creates an inactive account. No token is returned — the user must
// wait for an admin to activate the account before they can log in.
// force_password_change is false because the user chose their own password.
func (s *service) Register(ctx context.Context, req RegisterRequest) error {
	_, err := s.userRepo.Create(ctx, req.Username, req.Password, false, false)
	return err
}

// ChangePassword verifies the current password, updates it to the new value,
// and returns a fresh JWT with forcePasswordChange = false.
func (s *service) ChangePassword(ctx context.Context, userId int, req ChangePasswordRequest) (string, error) {
	u, err := s.userRepo.GetByIdWithCredentials(ctx, userId)
	if err != nil {
		return "", err
	}

	if !verifyPassword(req.CurrentPassword, u) {
		return "", ErrPasswordMismatch
	}

	if err := s.userRepo.UpdatePassword(ctx, userId, req.NewPassword); err != nil {
		return "", err
	}

	// Build the updated user locally to avoid an extra round-trip.
	u.ForcePasswordChange = false
	return generateToken(u, s.secret)
}
