package database

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"

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

		if err := repairMigrations(migrationsFS); err != nil {
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

// repairMigrations brings the database back to a clean baseline by running every
// down migration (highest version first) via raw SQL, then clearing schema_migrations.
// All down statements use IF EXISTS so the function is safe to call regardless of
// how much of the schema was actually applied.
func repairMigrations(migrationsFS fs.FS) error {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, connStr())
	if err != nil {
		return fmt.Errorf("repair: connect: %w", err)
	}
	defer conn.Close(ctx)

	entries, err := fs.ReadDir(migrationsFS, ".")
	if err != nil {
		return fmt.Errorf("repair: list migrations: %w", err)
	}

	// Collect down-migration filenames and sort highest-version first.
	var downFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".down.sql") {
			downFiles = append(downFiles, e.Name())
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(downFiles)))

	for _, name := range downFiles {
		sql, err := fs.ReadFile(migrationsFS, name)
		if err != nil {
			return fmt.Errorf("repair: read %s: %w", name, err)
		}
		if _, err := conn.Exec(ctx, string(sql)); err != nil {
			// Objects may not exist if the migration never completed — log and continue.
			log.Printf("repair: %s: %v", name, err)
		}
	}

	// Reset migration tracking. The table itself may not exist on a very first run,
	// so ignore errors here.
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
