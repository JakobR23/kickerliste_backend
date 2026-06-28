package team

import (
	"context"
	"time"

	"bierliste_backend/internal/database"
	"bierliste_backend/internal/entity"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// GetAll returns all teams with their members using a single JOIN query.
func (r *Repository) GetAll(ctx context.Context) ([]entity.Team, error) {
	rows, err := r.db.With(ctx).Query(ctx, `
		SELECT t.id, t.name, t.created_at,
		       u.id, u.username, u.role::text, u.active,
		       COALESCE(v.total_score, 0) AS total_score
		FROM team t
		LEFT JOIN team_member tm ON tm.team_id = t.id
		LEFT JOIN "user" u ON u.id = tm.user_id
		LEFT JOIN user_total_score v ON v.id = u.id
		ORDER BY t.id, u.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []entity.Team
	index := map[int]int{} // team id → index in result

	for rows.Next() {
		var teamId int
		var teamName *string
		var teamCreatedAt time.Time
		var uid *int
		var username, role *string
		var active *bool
		var totalScore *float64

		if err := rows.Scan(
			&teamId, &teamName, &teamCreatedAt,
			&uid, &username, &role, &active, &totalScore,
		); err != nil {
			return nil, err
		}

		idx, seen := index[teamId]
		if !seen {
			t := entity.Team{Members: []entity.User{}}
			t.Id = teamId
			if teamName != nil {
				t.Name = *teamName
			}
			t.CreatedAt = teamCreatedAt
			result = append(result, t)
			idx = len(result) - 1
			index[teamId] = idx
		}

		if uid != nil {
			result[idx].Members = append(result[idx].Members, entity.User{
				Id:         *uid,
				Username:   *username,
				Role:       entity.Role(*role),
				Active:     *active,
				TotalScore: *totalScore,
			})
		}
	}
	return result, rows.Err()
}

func (r *Repository) GetById(ctx context.Context, id int) (entity.Team, error) {
	rows, err := r.db.With(ctx).Query(ctx, `
		SELECT t.id, t.name, t.created_at,
		       u.id, u.username, u.role::text, u.active,
		       COALESCE(v.total_score, 0) AS total_score
		FROM team t
		LEFT JOIN team_member tm ON tm.team_id = t.id
		LEFT JOIN "user" u ON u.id = tm.user_id
		LEFT JOIN user_total_score v ON v.id = u.id
		WHERE t.id = $1
		ORDER BY u.id`, id)
	if err != nil {
		return entity.Team{}, err
	}
	defer rows.Close()

	var t entity.Team
	t.Members = []entity.User{}
	found := false

	for rows.Next() {
		var teamId int
		var teamName *string
		var teamCreatedAt time.Time
		var uid *int
		var username, role *string
		var active *bool
		var totalScore *float64

		if err := rows.Scan(
			&teamId, &teamName, &teamCreatedAt,
			&uid, &username, &role, &active, &totalScore,
		); err != nil {
			return entity.Team{}, err
		}

		if !found {
			t.Id = teamId
			if teamName != nil {
				t.Name = *teamName
			}
			t.CreatedAt = teamCreatedAt
			found = true
		}

		if uid != nil {
			t.Members = append(t.Members, entity.User{
				Id:         *uid,
				Username:   *username,
				Role:       entity.Role(*role),
				Active:     *active,
				TotalScore: *totalScore,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return entity.Team{}, err
	}
	if !found {
		return entity.Team{}, pgx.ErrNoRows
	}
	return t, nil
}

func (r *Repository) Create(ctx context.Context, name *string) (entity.Team, error) {
	var id int
	err := r.db.With(ctx).QueryRow(ctx,
		`INSERT INTO team (name) VALUES ($1) RETURNING id`, name).Scan(&id)
	if err != nil {
		return entity.Team{}, err
	}
	return r.GetById(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id int, name *string) (entity.Team, error) {
	_, err := r.db.With(ctx).Exec(ctx, `UPDATE team SET name = $1 WHERE id = $2`, name, id)
	if err != nil {
		return entity.Team{}, err
	}
	return r.GetById(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	_, err := r.db.With(ctx).Exec(ctx, `DELETE FROM team WHERE id = $1`, id)
	return err
}
