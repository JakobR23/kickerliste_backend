package database

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"

	"bierliste_backend/env"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Querier is satisfied by both *pgx.Conn and pgx.Tx, allowing repositories
// to work transparently with either a plain connection or an active transaction.
type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type txKey struct{}

// DB wraps a pgx connection and provides context-aware query access.
type DB struct {
	conn *pgx.Conn
}

// New wraps an open connection.
func New(conn *pgx.Conn) *DB {
	return &DB{conn: conn}
}

// connStr builds the base postgres:// connection string from environment variables.
func connStr() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		env.DatabaseUser.GetValue(),
		env.DatabasePassword.GetValue(),
		env.DatabaseHost.GetValue(),
		env.DatabasePort.GetValue(),
		env.DatabaseName.GetValue())
}

// InitializeConnection opens and returns a new database connection.
func InitializeConnection() *pgx.Conn {
	conn, err := pgx.Connect(context.Background(), connStr())
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	return conn
}

// RunMigrations applies all pending up-migrations embedded in migrationsFS.
//
// Recovery strategy when a migration is found dirty (previously started but
// not completed):
//
//  1. Non-destructive retry (all environments): reset schema_migrations to the
//     last clean version via direct SQL and re-run Up(). If the failed migration
//     ran inside a PostgreSQL transaction the schema is unchanged, so this simply
//     re-applies the failed migration without touching any existing data.
//
//  2. Full schema repair (development only, APP_ENV=development): if the
//     non-destructive retry fails — meaning schema objects from the partial run
//     are still present — drop every application table and type with CASCADE and
//     re-run all migrations from scratch. This is destructive and must never run
//     against a database with data worth keeping.
//
//  In production a dirty migration that cannot be non-destructively recovered
//  causes an immediate fatal, forcing a human operator to resolve it manually.
func RunMigrations(migrationsFS fs.FS) {
	dbURL := fmt.Sprintf("pgx5://%s:%s@%s:%s/%s",
		env.DatabaseUser.GetValue(),
		env.DatabasePassword.GetValue(),
		env.DatabaseHost.GetValue(),
		env.DatabasePort.GetValue(),
		env.DatabaseName.GetValue())

	newMigrate := func() *migrate.Migrate {
		d, err := iofs.New(migrationsFS, ".")
		if err != nil {
			log.Fatalf("migration source error: %v", err)
		}
		m, err := migrate.NewWithSourceInstance("iofs", d, dbURL)
		if err != nil {
			log.Fatalf("migration init error: %v", err)
		}
		return m
	}

	m := newMigrate()
	defer m.Close()

	if err := m.Up(); err == nil || errors.Is(err, migrate.ErrNoChange) {
		return
	} else {
		var dirtyErr migrate.ErrDirty
		if !errors.As(err, &dirtyErr) {
			log.Fatalf("migration failed: %v", err)
		}

		// Stage 1 — non-destructive: reset schema_migrations to the previous clean
		// version and retry. Works when the failed migration was transactional and
		// the schema is still intact.
		log.Printf("dirty migration at version %d — attempting non-destructive recovery", dirtyErr.Version)
		if err := resetToPreviousVersion(dirtyErr.Version); err != nil {
			log.Fatalf("migration: could not reset dirty state: %v", err)
		}

		m2 := newMigrate()
		defer m2.Close()

		if err := m2.Up(); err == nil || errors.Is(err, migrate.ErrNoChange) {
			return
		} else {
			// Stage 2 — destructive: the migration left partial schema objects
			// behind (non-transactional failure). Only permitted in development.
			if !env.IsDevelopment() {
				log.Fatalf(
					"migration failed after non-destructive recovery and full schema "+
						"repair is disabled outside of development (APP_ENV=%q). "+
						"Resolve manually by inspecting schema_migrations and the schema: %v",
					env.AppEnv.GetValue(), err)
			}

			log.Printf("non-destructive recovery failed (%v) — falling back to full schema repair (development only)", err)
			if err := repairMigrations(); err != nil {
				log.Fatalf("migration repair failed: %v", err)
			}

			m3 := newMigrate()
			defer m3.Close()

			if err := m3.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				log.Fatalf("migration failed after full repair: %v", err)
			}
		}
	}
}

// resetToPreviousVersion sets schema_migrations to (dirtyVersion-1, dirty=false)
// using direct SQL. This bypasses golang-migrate's Force() API, which has an
// internal uint conversion that maps -1 to 0 and causes subsequent Up() calls
// to fail. When dirtyVersion is 1 the row is simply deleted, leaving the table
// empty (no migrations applied).
func resetToPreviousVersion(dirtyVersion int) error {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, connStr())
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, `DELETE FROM schema_migrations`); err != nil {
		return fmt.Errorf("clear schema_migrations: %w", err)
	}
	if dirtyVersion > 1 {
		if _, err := conn.Exec(ctx,
			`INSERT INTO schema_migrations (version, dirty) VALUES ($1, false)`,
			dirtyVersion-1); err != nil {
			return fmt.Errorf("restore previous version: %w", err)
		}
	}
	return nil
}

// repairMigrations brings the database back to a clean baseline so that
// RunMigrations can re-apply every up-migration from scratch.
//
// Rather than running down-migration SQL files (which carry state assumptions
// about which earlier migrations have fully applied), we drop every known
// application-level object directly with DROP … IF EXISTS … CASCADE.
// CASCADE lets PostgreSQL resolve the dependency order automatically — views,
// foreign keys, and other dependents are removed without needing to be listed
// explicitly — and IF EXISTS makes each statement safe to run regardless of
// how much of the schema was actually applied before the failure.
//
// NOTE: this list must be updated whenever a migration adds a new top-level
// table or enum type.  Views and indexes do not need to be listed because they
// are dropped automatically via CASCADE when their parent table is dropped.
func repairMigrations() error {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, connStr())
	if err != nil {
		return fmt.Errorf("repair: connect: %w", err)
	}
	defer conn.Close(ctx)

	drops := []string{
		// Tables — drop in child-first order so FK constraints are never an obstacle,
		// though CASCADE would handle that anyway.
		`DROP TABLE IF EXISTS fixture     CASCADE`,
		`DROP TABLE IF EXISTS team_member CASCADE`,
		`DROP TABLE IF EXISTS team        CASCADE`,
		`DROP TABLE IF EXISTS "user"      CASCADE`,
		// Enum types — must come after tables that reference them.
		`DROP TYPE IF EXISTS fixture_status CASCADE`,
		`DROP TYPE IF EXISTS match_result   CASCADE`,
		`DROP TYPE IF EXISTS user_role      CASCADE`,
	}

	for _, stmt := range drops {
		if _, err := conn.Exec(ctx, stmt); err != nil {
			// Shouldn't happen with IF EXISTS, but log rather than abort.
			log.Printf("repair: warning: %v", err)
		}
	}

	// Reset migration tracking so Up() starts from a clean slate.
	// Ignore errors — the table may not exist on the very first run.
	if _, err := conn.Exec(ctx, `DELETE FROM schema_migrations`); err != nil {
		log.Printf("repair: clear schema_migrations: %v", err)
	}

	return nil
}

// With returns the active transaction stored in ctx, or the base connection
// if no transaction is in progress. Repositories always call this instead of
// using the connection directly, so they participate in transactions for free.
func (db *DB) With(ctx context.Context) Querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return db.conn
}

// Transactional runs fn inside a database transaction. The transaction is
// injected into ctx so any With(ctx) call inside fn automatically uses it.
// The transaction is rolled back if fn returns an error, committed otherwise.
func (db *DB) Transactional(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := db.conn.Begin(ctx)
	if err != nil {
		return err
	}
	ctx = context.WithValue(ctx, txKey{}, tx)
	if err := fn(ctx); err != nil {
		tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}
