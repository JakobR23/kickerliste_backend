package user

import (
	"context"

	"bierliste_backend/internal/database"
	"bierliste_backend/internal/entity"
	"bierliste_backend/internal/hash"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll(ctx context.Context) ([]entity.User, error) {
	rows, err := r.db.With(ctx).Query(ctx,
		`SELECT id, username, role::text, total_score FROM user_total_score ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var u entity.User
		var role string
		if err := rows.Scan(&u.Id, &u.Username, &role, &u.TotalScore); err != nil {
			return nil, err
		}
		u.Role = entity.Role(role)
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *Repository) GetById(ctx context.Context, id int) (entity.User, error) {
	var u entity.User
	var role string
	err := r.db.With(ctx).QueryRow(ctx,
		`SELECT id, username, role::text, total_score FROM user_total_score WHERE id = $1`, id).
		Scan(&u.Id, &u.Username, &role, &u.TotalScore)
	if err != nil {
		return entity.User{}, err
	}
	u.Role = entity.Role(role)
	return u, nil
}

// GetByUsername fetches a user with full credentials for authentication.
// Queries the base table directly (not the view) to retrieve password and hashsalt.
func (r *Repository) GetByUsername(ctx context.Context, username string) (entity.User, error) {
	var id, totalScore int
	var uname, role, password, hashsalt string
	err := r.db.With(ctx).QueryRow(ctx, `
		SELECT u.id, u.username, u.role::text, u.password, u.hashsalt,
		       COALESCE(SUM(f.value), 0) AS total_score
		FROM "user" u
		LEFT JOIN team_member tm ON tm.user_id = u.id
		LEFT JOIN fixture f ON (f.team_1_id = tm.team_id AND f.result = 'team_1')
		                    OR (f.team_2_id = tm.team_id AND f.result = 'team_2')
		WHERE u.username = $1
		GROUP BY u.id, u.username, u.role, u.password, u.hashsalt`, username).
		Scan(&id, &uname, &role, &password, &hashsalt, &totalScore)
	if err != nil {
		return entity.User{}, err
	}
	return entity.NewUser(id, uname, entity.Role(role), password, hashsalt, totalScore), nil
}

func (r *Repository) Create(ctx context.Context, username, password string) (entity.User, error) {
	salt := hash.GenerateSalt()
	hashed := hash.Password(password, salt)

	var id int
	err := r.db.With(ctx).QueryRow(ctx,
		`INSERT INTO "user" (username, password, hashsalt) VALUES ($1, $2, $3) RETURNING id`,
		username, hashed, salt).Scan(&id)
	if err != nil {
		return entity.User{}, err
	}
	return r.GetById(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id int, username string) (entity.User, error) {
	_, err := r.db.With(ctx).Exec(ctx,
		`UPDATE "user" SET username = $1 WHERE id = $2`, username, id)
	if err != nil {
		return entity.User{}, err
	}
	return r.GetById(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	_, err := r.db.With(ctx).Exec(ctx, `DELETE FROM "user" WHERE id = $1`, id)
	return err
}
