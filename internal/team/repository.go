package team

import (
	"context"
	"time"

	"bierliste_backend/internal/database"
	"bierliste_backend/internal/entity"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll(ctx context.Context) ([]entity.Team, error) {
	rows, err := r.db.With(ctx).Query(ctx, `SELECT id, name, created_at FROM team ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []entity.Team
	for rows.Next() {
		var t entity.Team
		var name *string
		if err := rows.Scan(&t.Id, &name, &t.CreatedAt); err != nil {
			return nil, err
		}
		if name != nil {
			t.Name = *name
		}
		teams = append(teams, t)
	}
	return teams, rows.Err()
}

func (r *Repository) GetById(ctx context.Context, id int) (entity.Team, error) {
	var t entity.Team
	var name *string
	err := r.db.With(ctx).QueryRow(ctx,
		`SELECT id, name, created_at FROM team WHERE id = $1`, id).
		Scan(&t.Id, &name, &t.CreatedAt)
	if err != nil {
		return entity.Team{}, err
	}
	if name != nil {
		t.Name = *name
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

// GetAllWithMembers returns every team together with its members in a single
// query. Teams with no members are included with an empty Members slice.
func (r *Repository) GetAllWithMembers(ctx context.Context) ([]entity.TeamWithMembers, error) {
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

	var result []entity.TeamWithMembers
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

		// Upsert team into result slice.
		idx, seen := index[teamId]
		if !seen {
			t := entity.TeamWithMembers{Members: []entity.User{}}
			t.Id = teamId
			if teamName != nil {
				t.Name = *teamName
			}
			t.CreatedAt = teamCreatedAt
			result = append(result, t)
			idx = len(result) - 1
			index[teamId] = idx
		}

		// Append member if the LEFT JOIN produced a row.
		if uid != nil {
			u := entity.User{
				Id:         *uid,
				Username:   *username,
				Role:       entity.Role(*role),
				Active:     *active,
				TotalScore: *totalScore,
			}
			result[idx].Members = append(result[idx].Members, u)
		}
	}
	return result, rows.Err()
}
