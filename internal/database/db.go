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
// It is a no-op if the schema is already up to date.
func RunMigrations(migrationsFS fs.FS) {
	d, err := iofs.New(migrationsFS, ".")
	if err != nil {
		log.Fatalf("migration source error: %v", err)
	}

	// golang-migrate's pgx v5 driver uses the pgx5:// scheme.
	dbURL := fmt.Sprintf("pgx5://%s:%s@%s:%s/%s",
		env.DatabaseUser.GetValue(),
		env.DatabasePassword.GetValue(),
		env.DatabaseHost.GetValue(),
		env.DatabasePort.GetValue(),
		env.DatabaseName.GetValue())

	m, err := migrate.NewWithSourceInstance("iofs", d, dbURL)
	if err != nil {
		log.Fatalf("migration init error: %v", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		var dirtyErr migrate.ErrDirty
		if errors.As(err, &dirtyErr) {
			// A previous run failed mid-migration, leaving the version flagged dirty.
			// Force the version back to the last clean state so Up can re-apply it.
			log.Printf("dirty migration detected at version %d — resetting to force clean re-run", dirtyErr.Version)
			if err := m.Force(dirtyErr.Version - 1); err != nil {
				log.Fatalf("migration force failed: %v", err)
			}
			if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
				log.Fatalf("migration failed after dirty reset: %v", err)
			}
			return
		}
		log.Fatalf("migration failed: %v", err)
	}
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
