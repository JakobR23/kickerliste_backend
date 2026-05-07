package teammember

import (
	"context"

	"bierliste_backend/internal/database"
	"bierliste_backend/internal/entity"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByTeamId(ctx context.Context, teamId int) ([]entity.TeamMember, error) {
	rows, err := r.db.With(ctx).Query(ctx,
		`SELECT id, team_id, user_id FROM team_member WHERE team_id = $1 ORDER BY id`, teamId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []entity.TeamMember
	for rows.Next() {
		var m entity.TeamMember
		if err := rows.Scan(&m.Id, &m.TeamId, &m.UserId); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *Repository) GetById(ctx context.Context, id int) (entity.TeamMember, error) {
	var m entity.TeamMember
	err := r.db.With(ctx).QueryRow(ctx,
		`SELECT id, team_id, user_id FROM team_member WHERE id = $1`, id).
		Scan(&m.Id, &m.TeamId, &m.UserId)
	if err != nil {
		return entity.TeamMember{}, err
	}
	return m, nil
}

func (r *Repository) Create(ctx context.Context, teamId, userId int) (entity.TeamMember, error) {
	var id int
	err := r.db.With(ctx).QueryRow(ctx,
		`INSERT INTO team_member (team_id, user_id) VALUES ($1, $2) RETURNING id`,
		teamId, userId).Scan(&id)
	if err != nil {
		return entity.TeamMember{}, err
	}
	return r.GetById(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, teamId, userId int) error {
	_, err := r.db.With(ctx).Exec(ctx,
		`DELETE FROM team_member WHERE team_id = $1 AND user_id = $2`, teamId, userId)
	return err
}
