package fixture

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

// GetAll returns fixtures ordered newest first.
// teamId, when non-nil, restricts results to fixtures involving that team.
// status, when non-nil, filters by that status; returns all statuses when nil.
func (r *Repository) GetAll(ctx context.Context, teamId *int, status *entity.FixtureStatus) ([]entity.Fixture, error) {
	var statusStr *string
	if status != nil {
		s := string(*status)
		statusStr = &s
	}
	rows, err := r.db.With(ctx).Query(ctx, `
		SELECT id, team_1_id, team_2_id, result::text, score_team_1, score_team_2,
		       played_at, value, status::text, submitted_by
		FROM fixture
		WHERE ($1::int IS NULL OR team_1_id = $1 OR team_2_id = $1)
		  AND ($2::fixture_status IS NULL OR status = $2::fixture_status)
		ORDER BY played_at DESC`,
		teamId, statusStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fixtures []entity.Fixture
	for rows.Next() {
		f, err := scanFixture(rows)
		if err != nil {
			return nil, err
		}
		fixtures = append(fixtures, f)
	}
	return fixtures, rows.Err()
}

// GetById returns a single fixture regardless of status, so submitters can
// check on their own pending or rejected submissions.
func (r *Repository) GetById(ctx context.Context, id int) (entity.Fixture, error) {
	rows, err := r.db.With(ctx).Query(ctx, `
		SELECT id, team_1_id, team_2_id, result::text, score_team_1, score_team_2,
		       played_at, value, status::text, submitted_by
		FROM fixture
		WHERE id = $1`, id)
	if err != nil {
		return entity.Fixture{}, err
	}
	defer rows.Close()

	if !rows.Next() {
		return entity.Fixture{}, pgx.ErrNoRows
	}
	return scanFixture(rows)
}

// Create inserts a new fixture with status 'pending'.
// submittedBy is the ID of the authenticated user making the submission.
func (r *Repository) Create(ctx context.Context, submittedBy, team1Id, team2Id int, result entity.MatchResult, team1Score, team2Score *int, playedAt time.Time, value int) (entity.Fixture, error) {
	var id int
	err := r.db.With(ctx).QueryRow(ctx, `
		INSERT INTO fixture
		    (team_1_id, team_2_id, result, score_team_1, score_team_2, played_at, value, submitted_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`,
		team1Id, team2Id, result, team1Score, team2Score, playedAt, value, submittedBy).Scan(&id)
	if err != nil {
		return entity.Fixture{}, err
	}
	return r.GetById(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id int, result entity.MatchResult, team1Score, team2Score *int, value int) (entity.Fixture, error) {
	_, err := r.db.With(ctx).Exec(ctx, `
		UPDATE fixture
		SET result = $1, score_team_1 = $2, score_team_2 = $3, value = $4
		WHERE id = $5`,
		result, team1Score, team2Score, value, id)
	if err != nil {
		return entity.Fixture{}, err
	}
	return r.GetById(ctx, id)
}

// UpdateStatus transitions a fixture to the given status.
func (r *Repository) UpdateStatus(ctx context.Context, id int, status entity.FixtureStatus) (entity.Fixture, error) {
	_, err := r.db.With(ctx).Exec(ctx,
		`UPDATE fixture SET status = $1 WHERE id = $2`, string(status), id)
	if err != nil {
		return entity.Fixture{}, err
	}
	return r.GetById(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id int) error {
	_, err := r.db.With(ctx).Exec(ctx, `DELETE FROM fixture WHERE id = $1`, id)
	return err
}

func scanFixture(rows pgx.Rows) (entity.Fixture, error) {
	var f entity.Fixture
	var result, status string
	err := rows.Scan(
		&f.Id, &f.Team1Id, &f.Team2Id, &result,
		&f.Team1Score, &f.Team2Score, &f.PlayedAt, &f.Value,
		&status, &f.SubmittedBy,
	)
	if err != nil {
		return entity.Fixture{}, err
	}
	f.Result = entity.MatchResult(result)
	f.Status = entity.FixtureStatus(status)
	return f, nil
}
