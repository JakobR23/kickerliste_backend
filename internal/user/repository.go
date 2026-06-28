package user

import (
	"context"

	"bierliste_backend/internal/database"
	"bierliste_backend/internal/entity"
	"bierliste_backend/internal/hash"

	"github.com/jackc/pgx/v5"
)

// userSelectCols is the shared SELECT/JOIN for loading the public user fields
// (no credentials) together with the derived total score. Callers append their
// own WHERE / ORDER BY clause.
const userSelectCols = `
	SELECT u.id, u.username, u.role::text, u.active, COALESCE(v.total_score, 0) AS total_score
	FROM "user" u
	LEFT JOIN user_total_score v ON v.id = u.id`

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// scanUser reads one row produced by userSelectCols into an entity.User.
// pgx.Rows also satisfies pgx.Row, so this works both inside a Query loop and
// for a single QueryRow result.
func scanUser(row pgx.Row) (entity.User, error) {
	var u entity.User
	var role string
	if err := row.Scan(&u.Id, &u.Username, &role, &u.Active, &u.TotalScore); err != nil {
		return entity.User{}, err
	}
	u.Role = entity.Role(role)
	return u, nil
}

func (r *Repository) GetAll(ctx context.Context, active bool) ([]entity.User, error) {
	rows, err := r.db.With(ctx).Query(ctx, userSelectCols+`
		WHERE u.active = $1
		ORDER BY u.id`, active)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *Repository) GetById(ctx context.Context, id int) (entity.User, error) {
	return scanUser(r.db.With(ctx).QueryRow(ctx, userSelectCols+` WHERE u.id = $1`, id))
}

// GetByUsername fetches a user with full credentials for authentication.
// Joins the user_total_score view for a consistent score calculation.
func (r *Repository) GetByUsername(ctx context.Context, username string) (entity.User, error) {
	var id int
	var totalScore float64
	var uname, role, password, hashsalt string
	var forcePasswordChange, active bool
	err := r.db.With(ctx).QueryRow(ctx, `
		SELECT u.id, u.username, u.role::text, u.password, u.hashsalt,
		       u.force_password_change, u.active,
		       COALESCE(v.total_score, 0) AS total_score
		FROM "user" u
		LEFT JOIN user_total_score v ON v.id = u.id
		WHERE u.username = $1`,
		username).
		Scan(&id, &uname, &role, &password, &hashsalt, &forcePasswordChange, &active, &totalScore)
	if err != nil {
		return entity.User{}, err
	}
	return entity.NewUser(id, uname, entity.Role(role), password, hashsalt, totalScore, forcePasswordChange, active), nil
}

// GetByIdWithCredentials fetches a user with full credentials by ID.
// Used by the change-password flow to verify the current password.
func (r *Repository) GetByIdWithCredentials(ctx context.Context, id int) (entity.User, error) {
	var uid int
	var username, role, password, hashsalt string
	var forcePasswordChange, active bool
	err := r.db.With(ctx).QueryRow(ctx, `
		SELECT id, username, role::text, password, hashsalt, force_password_change, active
		FROM "user"
		WHERE id = $1`, id).
		Scan(&uid, &username, &role, &password, &hashsalt, &forcePasswordChange, &active)
	if err != nil {
		return entity.User{}, err
	}
	return entity.NewUser(uid, username, entity.Role(role), password, hashsalt, 0, forcePasswordChange, active), nil
}

// Create inserts a new user and returns the created record.
// Set forcePasswordChange = true for admin-created accounts so the user is
// prompted to set their own password on first login.
// Set active = true for admin-created accounts; false for self-registration
// (requires admin activation before the account can be used).
func (r *Repository) Create(ctx context.Context, username, password string, forcePasswordChange, active bool) (entity.User, error) {
	hashed, err := hash.HashPassword(password)
	if err != nil {
		return entity.User{}, err
	}

	var id int
	err = r.db.With(ctx).QueryRow(ctx,
		// hashsalt is unused for bcrypt (salt is embedded in the hash) but the
		// column is NOT NULL so we store an empty string as a sentinel.
		`INSERT INTO "user" (username, password, hashsalt, force_password_change, active) VALUES ($1, $2, '', $3, $4) RETURNING id`,
		username, hashed, forcePasswordChange, active).Scan(&id)
	if err != nil {
		return entity.User{}, err
	}

	u, err := r.GetById(ctx, id)
	if err != nil {
		return entity.User{}, err
	}
	// GetById does not include force_password_change; carry it forward from
	// the known insert value.
	u.ForcePasswordChange = forcePasswordChange
	return u, nil
}

// Activate sets the user's active flag to true, allowing them to log in.
func (r *Repository) Activate(ctx context.Context, id int) error {
	_, err := r.db.With(ctx).Exec(ctx,
		`UPDATE "user" SET active = TRUE WHERE id = $1`, id)
	return err
}

// UpdatePassword replaces the user's hashed password and clears the
// force_password_change flag. The hashsalt column is set to '' because bcrypt
// embeds its own salt in the hash string.
func (r *Repository) UpdatePassword(ctx context.Context, id int, newPassword string) error {
	hashed, err := hash.HashPassword(newPassword)
	if err != nil {
		return err
	}
	_, err = r.db.With(ctx).Exec(ctx,
		`UPDATE "user" SET password = $1, hashsalt = '', force_password_change = FALSE WHERE id = $2`,
		hashed, id)
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

// UpdateRole changes the role of an existing user and returns the updated record.
func (r *Repository) UpdateRole(ctx context.Context, id int, role entity.Role) (entity.User, error) {
	_, err := r.db.With(ctx).Exec(ctx,
		`UPDATE "user" SET role = $1::user_role WHERE id = $2`, string(role), id)
	if err != nil {
		return entity.User{}, err
	}
	return r.GetById(ctx, id)
}
