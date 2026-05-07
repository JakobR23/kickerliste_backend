package fixture

import (
	"context"
	"errors"
	"time"

	"bierliste_backend/internal/entity"
)

// ErrSameTeam is returned when both sides of a fixture reference the same team.
var ErrSameTeam = errors.New("team1Id and team2Id must be different")

// ErrScoresInconsistent is returned when only one score is provided.
var ErrScoresInconsistent = errors.New("team1Score and team2Score must both be set or both be null")

// Service encapsulates the business logic for fixture management.
type Service interface {
	GetAll(ctx context.Context, teamId *int) ([]entity.Fixture, error)
	GetById(ctx context.Context, id int) (entity.Fixture, error)
	Create(ctx context.Context, req CreateRequest) (entity.Fixture, error)
	Update(ctx context.Context, id int, req UpdateRequest) (entity.Fixture, error)
	Delete(ctx context.Context, id int) error
}

// CreateRequest holds the fields required to record a new fixture.
type CreateRequest struct {
	Team1Id    int               `json:"team1Id"    binding:"required"`
	Team2Id    int               `json:"team2Id"    binding:"required"`
	Result     entity.MatchResult `json:"result"     binding:"required"`
	Team1Score *int              `json:"team1Score"`
	Team2Score *int              `json:"team2Score"`
	PlayedAt   time.Time         `json:"playedAt"`
	Value      int               `json:"value"      binding:"min=0"`
}

// UpdateRequest holds the fields that can be changed on an existing fixture.
type UpdateRequest struct {
	Result     entity.MatchResult `json:"result"     binding:"required"`
	Team1Score *int              `json:"team1Score"`
	Team2Score *int              `json:"team2Score"`
	Value      int               `json:"value"      binding:"min=0"`
}

type service struct {
	repo *Repository
}

// NewService creates a Service backed by the given repository.
func NewService(repo *Repository) Service {
	return &service{repo: repo}
}

// GetAll returns all fixtures, optionally filtered to those involving teamId.
func (s *service) GetAll(ctx context.Context, teamId *int) ([]entity.Fixture, error) {
	if teamId != nil {
		return s.repo.GetByTeamId(ctx, *teamId)
	}
	return s.repo.GetAll(ctx)
}

func (s *service) GetById(ctx context.Context, id int) (entity.Fixture, error) {
	return s.repo.GetById(ctx, id)
}

func (s *service) Create(ctx context.Context, req CreateRequest) (entity.Fixture, error) {
	if req.Team1Id == req.Team2Id {
		return entity.Fixture{}, ErrSameTeam
	}
	if (req.Team1Score == nil) != (req.Team2Score == nil) {
		return entity.Fixture{}, ErrScoresInconsistent
	}
	playedAt := req.PlayedAt
	if playedAt.IsZero() {
		playedAt = time.Now()
	}
	return s.repo.Create(ctx, req.Team1Id, req.Team2Id, req.Result, req.Team1Score, req.Team2Score, playedAt, req.Value)
}

func (s *service) Update(ctx context.Context, id int, req UpdateRequest) (entity.Fixture, error) {
	if (req.Team1Score == nil) != (req.Team2Score == nil) {
		return entity.Fixture{}, ErrScoresInconsistent
	}
	return s.repo.Update(ctx, id, req.Result, req.Team1Score, req.Team2Score, req.Value)
}

func (s *service) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
