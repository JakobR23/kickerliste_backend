package team

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
