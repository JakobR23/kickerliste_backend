package user

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"bierliste_backend/internal/database"
	"bierliste_backend/internal/entity"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll(ctx context.Context) ([]entity.User, error) {
	rows, err := r.db.With(ctx).Query(ctx,
		`SELECT id, username, total_score FROM user_total_score ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var u entity.User
		if err := rows.Scan(&u.Id, &u.Username, &u.TotalScore); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *Repository) GetById(ctx context.Context, id int) (entity.User, error) {
	var u entity.User
	err := r.db.With(ctx).QueryRow(ctx,
		`SELECT id, username, total_score FROM user_total_score WHERE id = $1`, id).
		Scan(&u.Id, &u.Username, &u.TotalScore)
	if err != nil {
		return entity.User{}, err
	}
	return u, nil
}

func (r *Repository) Create(ctx context.Context, username, password string) (entity.User, error) {
	salt := generateSalt()
	hash := hashPassword(password, salt)

	var id int
	err := r.db.With(ctx).QueryRow(ctx,
		`INSERT INTO "user" (username, password, hashsalt) VALUES ($1, $2, $3) RETURNING id`,
		username, hash, salt).Scan(&id)
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

func generateSalt() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func hashPassword(password, salt string) string {
	h := sha256.Sum256([]byte(password + salt))
	return hex.EncodeToString(h[:])
}
