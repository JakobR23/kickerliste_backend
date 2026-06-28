package team

import (
	"context"
	"net/http"

	"bierliste_backend/internal/entity"
	"bierliste_backend/internal/httputil"
	"bierliste_backend/internal/teammember"
	"bierliste_backend/internal/user"
)

// ErrTeamFull is returned when a team already has the maximum of 2 members.
var ErrTeamFull = httputil.NewStatusError(http.StatusConflict, "team already has 2 members")

// Service encapsulates the business logic for team and team-member management.
type Service interface {
	GetAll(ctx context.Context) ([]entity.Team, error)
	GetById(ctx context.Context, id int) (entity.Team, error)
	Create(ctx context.Context, req CreateRequest) (entity.Team, error)
	Update(ctx context.Context, id int, req UpdateRequest) (entity.Team, error)
	Delete(ctx context.Context, id int) error

	AddMember(ctx context.Context, teamId int, req AddMemberRequest) (entity.User, error)
	RemoveMember(ctx context.Context, teamId, userId int) error
}

// CreateRequest holds the fields required to create a new team.
type CreateRequest struct {
	Name *string `json:"name"`
}

// UpdateRequest holds the fields that can be changed on an existing team.
type UpdateRequest struct {
	Name *string `json:"name"`
}

// AddMemberRequest holds the user to add to a team.
type AddMemberRequest struct {
	UserId int `json:"userId" binding:"required"`
}

type service struct {
	teamRepo *Repository
	tmRepo   *teammember.Repository
	userRepo *user.Repository
}

// NewService creates a Service backed by the given repositories.
func NewService(teamRepo *Repository, tmRepo *teammember.Repository, userRepo *user.Repository) Service {
	return &service{
		teamRepo: teamRepo,
		tmRepo:   tmRepo,
		userRepo: userRepo,
	}
}

func (s *service) GetAll(ctx context.Context) ([]entity.Team, error) {
	return s.teamRepo.GetAll(ctx)
}

func (s *service) GetById(ctx context.Context, id int) (entity.Team, error) {
	return s.teamRepo.GetById(ctx, id)
}

func (s *service) Create(ctx context.Context, req CreateRequest) (entity.Team, error) {
	return s.teamRepo.Create(ctx, req.Name)
}

func (s *service) Update(ctx context.Context, id int, req UpdateRequest) (entity.Team, error) {
	return s.teamRepo.Update(ctx, id, req.Name)
}

func (s *service) Delete(ctx context.Context, id int) error {
	return s.teamRepo.Delete(ctx, id)
}

func (s *service) AddMember(ctx context.Context, teamId int, req AddMemberRequest) (entity.User, error) {
	if _, err := s.teamRepo.GetById(ctx, teamId); err != nil {
		return entity.User{}, err
	}
	members, err := s.tmRepo.GetByTeamId(ctx, teamId)
	if err != nil {
		return entity.User{}, err
	}
	if len(members) >= 2 {
		return entity.User{}, ErrTeamFull
	}
	if _, err := s.tmRepo.Create(ctx, teamId, req.UserId); err != nil {
		return entity.User{}, err
	}
	return s.userRepo.GetById(ctx, req.UserId)
}

func (s *service) RemoveMember(ctx context.Context, teamId, userId int) error {
	return s.tmRepo.Delete(ctx, teamId, userId)
}
