package scoreadjustment

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

func (r *Repository) GetByUserId(ctx context.Context, userId int) ([]entity.ScoreAdjustment, error) {
	rows, err := r.db.With(ctx).Query(ctx, `
		SELECT id, user_id, amount, reason, created_by, created_at
		FROM score_adjustment
		WHERE user_id = $1
		ORDER BY created_at DESC`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var adjustments []entity.ScoreAdjustment
	for rows.Next() {
		var a entity.ScoreAdjustment
		if err := rows.Scan(&a.Id, &a.UserId, &a.Amount, &a.Reason, &a.CreatedBy, &a.CreatedAt); err != nil {
			return nil, err
		}
		adjustments = append(adjustments, a)
	}
	return adjustments, rows.Err()
}

func (r *Repository) Create(ctx context.Context, userId, createdBy, amount int, reason string) (entity.ScoreAdjustment, error) {
	var a entity.ScoreAdjustment
	err := r.db.With(ctx).QueryRow(ctx, `
		INSERT INTO score_adjustment (user_id, amount, reason, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, amount, reason, created_by, created_at`,
		userId, amount, reason, createdBy).
		Scan(&a.Id, &a.UserId, &a.Amount, &a.Reason, &a.CreatedBy, &a.CreatedAt)
	if err != nil {
		return entity.ScoreAdjustment{}, err
	}
	return a, nil
}
