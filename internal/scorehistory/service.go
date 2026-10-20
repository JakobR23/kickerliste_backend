package scorehistory

import (
	"context"

	"bierliste_backend/internal/entity"
	"bierliste_backend/internal/user"
)

// Service encapsulates the business logic for the per-user score history.
type Service interface {
	GetByUserId(ctx context.Context, userId int) ([]entity.ScoreHistoryEntry, error)
}

type service struct {
	repo     *Repository
	userRepo *user.Repository
}

// NewService creates a Service. userRepo is used to tell an unknown user (404)
// apart from a known user with no history (empty list).
func NewService(repo *Repository, userRepo *user.Repository) Service {
	return &service{repo: repo, userRepo: userRepo}
}

func (s *service) GetByUserId(ctx context.Context, userId int) ([]entity.ScoreHistoryEntry, error) {
	// A missing user must surface as 404, not an empty history.
	if _, err := s.userRepo.GetById(ctx, userId); err != nil {
		return nil, err
	}
	return s.repo.GetByUserId(ctx, userId)
}
