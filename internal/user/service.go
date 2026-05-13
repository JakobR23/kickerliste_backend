package user

import (
	"context"

	"bierliste_backend/internal/entity"
)

// Service encapsulates the business logic for user management.
type Service interface {
	GetAll(ctx context.Context, active bool) ([]entity.User, error)
	GetById(ctx context.Context, id int) (entity.User, error)
	Create(ctx context.Context, req CreateRequest) (entity.User, error)
	Update(ctx context.Context, id int, req UpdateRequest) (entity.User, error)
	Delete(ctx context.Context, id int) error
	Activate(ctx context.Context, id int) error
}

// CreateRequest holds the fields required to register a new user.
type CreateRequest struct {
	Username string `json:"username" binding:"required,min=1,max=255"`
	Password string `json:"password" binding:"required,min=8"`
}

// UpdateRequest holds the fields that can be changed on an existing user.
type UpdateRequest struct {
	Username string `json:"username" binding:"required,min=1,max=255"`
}

type service struct {
	repo *Repository
}

// NewService creates a Service backed by the given repository.
func NewService(repo *Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetAll(ctx context.Context, active bool) ([]entity.User, error) {
	return s.repo.GetAll(ctx, active)
}

func (s *service) GetById(ctx context.Context, id int) (entity.User, error) {
	return s.repo.GetById(ctx, id)
}

// Create is used by admins to create accounts directly. Admin-created accounts
// are active immediately and require a password change on first login.
func (s *service) Create(ctx context.Context, req CreateRequest) (entity.User, error) {
	return s.repo.Create(ctx, req.Username, req.Password, true, true)
}

func (s *service) Update(ctx context.Context, id int, req UpdateRequest) (entity.User, error) {
	return s.repo.Update(ctx, id, req.Username)
}

func (s *service) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *service) Activate(ctx context.Context, id int) error {
	return s.repo.Activate(ctx, id)
}
