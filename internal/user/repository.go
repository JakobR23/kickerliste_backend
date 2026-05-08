package user

import (
	"context"
	"errors"

	"bierliste_backend/internal/database"
	"bierliste_backend/internal/entity"
	"bierliste_backend/internal/hash"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrUsernameTaken is returned when a CREATE violates the unique constraint on username.
var ErrUsernameTaken = errors.New("username already taken")

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
// Queries the base table directly (not the view) to retrieve password, hashsalt,
// and force_password_change.
func (r *Repository) GetByUsername(ctx context.Context, username string) (entity.User, error) {
	var id, totalScore int
	var uname, role, password, hashsalt string
	var forcePasswordChange bool
	err := r.db.With(ctx).QueryRow(ctx, `
		SELECT u.id, u.username, u.role::text, u.password, u.hashsalt,
		       u.force_password_change,
		       COALESCE(SUM(f.value), 0) AS total_score
		FROM "user" u
		LEFT JOIN team_member tm ON tm.user_id = u.id
		LEFT JOIN fixture f ON (f.team_1_id = tm.team_id AND f.result = 'team_1')
		                    OR (f.team_2_id = tm.team_id AND f.result = 'team_2')
		WHERE u.username = $1
		GROUP BY u.id, u.username, u.role, u.password, u.hashsalt, u.force_password_change`,
		username).
		Scan(&id, &uname, &role, &password, &hashsalt, &forcePasswordChange, &totalScore)
	if err != nil {
		return entity.User{}, err
	}
	return entity.NewUser(id, uname, entity.Role(role), password, hashsalt, totalScore, forcePasswordChange), nil
}

// GetByIdWithCredentials fetches a user with full credentials by ID.
// Used by the change-password flow to verify the current password.
func (r *Repository) GetByIdWithCredentials(ctx context.Context, id int) (entity.User, error) {
	var uid int
	var username, role, password, hashsalt string
	var forcePasswordChange bool
	err := r.db.With(ctx).QueryRow(ctx, `
		SELECT id, username, role::text, password, hashsalt, force_password_change
		FROM "user"
		WHERE id = $1`, id).
		Scan(&uid, &username, &role, &password, &hashsalt, &forcePasswordChange)
	if err != nil {
		return entity.User{}, err
	}
	return entity.NewUser(uid, username, entity.Role(role), password, hashsalt, 0, forcePasswordChange), nil
}

// Create inserts a new user and returns the created record.
// Set forcePasswordChange = true for admin-created accounts so the user is
// prompted to set their own password on first login.
func (r *Repository) Create(ctx context.Context, username, password string, forcePasswordChange bool) (entity.User, error) {
	salt := hash.GenerateSalt()
	hashed := hash.Password(password, salt)

	var id int
	err := r.db.With(ctx).QueryRow(ctx,
		`INSERT INTO "user" (username, password, hashsalt, force_password_change) VALUES ($1, $2, $3, $4) RETURNING id`,
		username, hashed, salt, forcePasswordChange).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			return entity.User{}, ErrUsernameTaken
		}
		return entity.User{}, err
	}

	u, err := r.GetById(ctx, id)
	if err != nil {
		return entity.User{}, err
	}
	// GetById queries the view which does not include force_password_change;
	// carry it forward from the known insert value.
	u.ForcePasswordChange = forcePasswordChange
	return u, nil
}

// UpdatePassword replaces the user's hashed password and salt, and clears the
// force_password_change flag.
func (r *Repository) UpdatePassword(ctx context.Context, id int, newPassword string) error {
	salt := hash.GenerateSalt()
	hashed := hash.Password(newPassword, salt)
	_, err := r.db.With(ctx).Exec(ctx,
		`UPDATE "user" SET password = $1, hashsalt = $2, force_password_change = FALSE WHERE id = $3`,
		hashed, salt, id)
	return err
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

// isUniqueViolation reports whether err is a PostgreSQL unique-constraint violation.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation
}
