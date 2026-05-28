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

func (r *Repository) GetAll(ctx context.Context, active bool) ([]entity.User, error) {
	rows, err := r.db.With(ctx).Query(ctx, `
		SELECT u.id, u.username, u.role::text, u.active, COALESCE(v.total_score, 0) AS total_score
		FROM "user" u
		LEFT JOIN user_total_score v ON v.id = u.id
		WHERE u.active = $1
		ORDER BY u.id`,
		active)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var u entity.User
		var role string
		if err := rows.Scan(&u.Id, &u.Username, &role, &u.Active, &u.TotalScore); err != nil {
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
	err := r.db.With(ctx).QueryRow(ctx, `
		SELECT u.id, u.username, u.role::text, u.active, COALESCE(v.total_score, 0) AS total_score
		FROM "user" u
		LEFT JOIN user_total_score v ON v.id = u.id
		WHERE u.id = $1`, id).
		Scan(&u.Id, &u.Username, &role, &u.Active, &u.TotalScore)
	if err != nil {
		return entity.User{}, err
	}
	u.Role = entity.Role(role)
	return u, nil
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
func (r *Repository) Create(ctx context.Context, username, password string, role entity.Role, forcePasswordChange, active bool) (entity.User, error) {
	salt := hash.GenerateSalt()
	hashed := hash.Password(password, salt)

	var id int
	err := r.db.With(ctx).QueryRow(ctx,
		`INSERT INTO "user" (username, password, hashsalt, role, force_password_change, active) VALUES ($1, $2, $3, $4::user_role, $5, $6) RETURNING id`,
		username, hashed, salt, string(role), forcePasswordChange, active).Scan(&id)
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
