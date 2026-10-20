package scorehistory

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

// GetByUserId returns the user's score-changing events oldest first, each with
// the cumulative running total after it. It reads the user_score_history view,
// which mirrors the leaderboard's scoring so the last entry's running total
// equals the user's total_score. The ORDER BY reproduces the deterministic
// order the view's running total was computed with.
func (r *Repository) GetByUserId(ctx context.Context, userId int) ([]entity.ScoreHistoryEntry, error) {
	rows, err := r.db.With(ctx).Query(ctx, `
		SELECT at, delta, type, fixture_id, reason, running_total
		FROM user_score_history
		WHERE user_id = $1
		ORDER BY at, type, seq`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Non-nil slice so an empty history serialises as [] rather than null.
	entries := []entity.ScoreHistoryEntry{}
	for rows.Next() {
		var e entity.ScoreHistoryEntry
		if err := rows.Scan(&e.At, &e.Delta, &e.Type, &e.FixtureId, &e.Reason, &e.RunningTotal); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
