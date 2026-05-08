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
// If the database is left in a dirty or otherwise corrupt state from a previous
// failed run, it repairs automatically by executing the down migrations directly
// via raw SQL and retrying.
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

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		var dirtyErr migrate.ErrDirty
		if errors.As(err, &dirtyErr) {
			log.Printf("dirty migration at version %d — repairing", dirtyErr.Version)
		} else {
			log.Printf("migration error (%v) — attempting repair", err)
		}

		if err := repairMigrations(); err != nil {
			log.Fatalf("migration repair failed: %v", err)
		}

		// Fresh instance so the internal source/DB state is clean.
		m2 := newMigrate()
		defer m2.Close()

		if err := m2.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("migration failed after repair: %v", err)
		}
	}
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
