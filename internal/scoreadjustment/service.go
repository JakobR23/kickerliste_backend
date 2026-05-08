package scoreadjustment

import (
	"context"
	"errors"

	"bierliste_backend/internal/entity"
)

// ErrZeroAmount is returned when an adjustment of zero is submitted.
var ErrZeroAmount = errors.New("amount must not be zero")

// Service encapsulates business logic for manual score adjustments.
type Service interface {
	GetByUserId(ctx context.Context, userId int) ([]entity.ScoreAdjustment, error)
	Create(ctx context.Context, adminUserId, targetUserId int, req CreateRequest) (entity.ScoreAdjustment, error)
}

// CreateRequest holds the fields required to record a manual score adjustment.
type CreateRequest struct {
	Amount int    `json:"amount" binding:"required"`
	Reason string `json:"reason" binding:"required,min=1"`
}

type service struct {
	repo *Repository
}

func NewService(repo *Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetByUserId(ctx context.Context, userId int) ([]entity.ScoreAdjustment, error) {
	return s.repo.GetByUserId(ctx, userId)
}

func (s *service) Create(ctx context.Context, adminUserId, targetUserId int, req CreateRequest) (entity.ScoreAdjustment, error) {
	if req.Amount == 0 {
		return entity.ScoreAdjustment{}, ErrZeroAmount
	}
	return s.repo.Create(ctx, targetUserId, adminUserId, req.Amount, req.Reason)
}
