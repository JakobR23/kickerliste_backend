package database

import (
	"context"
	"fmt"
	"log"

	"bierliste_backend/env"

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

// InitializeConnection opens and returns a new database connection.
func InitializeConnection() *pgx.Conn {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s",
		env.DatabaseUser.GetValue(),
		env.DatabasePassword.GetValue(),
		env.DatabaseHost.GetValue(),
		env.DatabasePort.GetValue())
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	return conn
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
