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

// ErrFixtureNotPending is returned when approve or reject is called on a
// fixture that is no longer in the pending state.
var ErrFixtureNotPending = errors.New("fixture is not pending")

// Service encapsulates the business logic for fixture management.
type Service interface {
	GetAll(ctx context.Context, teamId *int) ([]entity.Fixture, error)
	GetById(ctx context.Context, id int) (entity.Fixture, error)
	Create(ctx context.Context, submittedBy int, req CreateRequest) (entity.Fixture, error)
	Update(ctx context.Context, id int, req UpdateRequest) (entity.Fixture, error)
	Delete(ctx context.Context, id int) error
	Approve(ctx context.Context, id int) (entity.Fixture, error)
	Reject(ctx context.Context, id int) (entity.Fixture, error)
}

// CreateRequest holds the fields required to submit a new fixture for approval.
type CreateRequest struct {
	Team1Id    int                `json:"team1Id"    binding:"required"`
	Team2Id    int                `json:"team2Id"    binding:"required"`
	Result     entity.MatchResult `json:"result"     binding:"required"`
	Team1Score *int               `json:"team1Score"`
	Team2Score *int               `json:"team2Score"`
	PlayedAt   time.Time          `json:"playedAt"`
	Value      int                `json:"value"      binding:"min=0"`
}

// UpdateRequest holds the fields that can be changed on an existing fixture.
type UpdateRequest struct {
	Result     entity.MatchResult `json:"result"     binding:"required"`
	Team1Score *int               `json:"team1Score"`
	Team2Score *int               `json:"team2Score"`
	Value      int                `json:"value"      binding:"min=0"`
}

type service struct {
	repo *Repository
}

// NewService creates a Service backed by the given repository.
func NewService(repo *Repository) Service {
	return &service{repo: repo}
}

// GetAll returns approved fixtures, optionally filtered to those involving teamId.
func (s *service) GetAll(ctx context.Context, teamId *int) ([]entity.Fixture, error) {
	if teamId != nil {
		return s.repo.GetByTeamId(ctx, *teamId)
	}
	return s.repo.GetAll(ctx)
}

func (s *service) GetById(ctx context.Context, id int) (entity.Fixture, error) {
	return s.repo.GetById(ctx, id)
}

// Create submits a new fixture for admin approval.
// submittedBy must be the ID of the authenticated caller.
func (s *service) Create(ctx context.Context, submittedBy int, req CreateRequest) (entity.Fixture, error) {
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
	return s.repo.Create(ctx, submittedBy, req.Team1Id, req.Team2Id, req.Result, req.Team1Score, req.Team2Score, playedAt, req.Value)
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

// Approve marks a pending fixture as approved, making it count toward scores.
// Returns ErrFixtureNotPending if the fixture is not in pending status.
func (s *service) Approve(ctx context.Context, id int) (entity.Fixture, error) {
	f, err := s.repo.GetById(ctx, id)
	if err != nil {
		return entity.Fixture{}, err
	}
	if f.Status != entity.StatusPending {
		return entity.Fixture{}, ErrFixtureNotPending
	}
	return s.repo.UpdateStatus(ctx, id, entity.StatusApproved)
}

// Reject marks a pending fixture as rejected. The record is preserved for
// audit purposes but does not count toward scores.
// Returns ErrFixtureNotPending if the fixture is not in pending status.
func (s *service) Reject(ctx context.Context, id int) (entity.Fixture, error) {
	f, err := s.repo.GetById(ctx, id)
	if err != nil {
		return entity.Fixture{}, err
	}
	if f.Status != entity.StatusPending {
		return entity.Fixture{}, ErrFixtureNotPending
	}
	return s.repo.UpdateStatus(ctx, id, entity.StatusRejected)
}
