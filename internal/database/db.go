package database

import (
	"context"
	"fmt"
	"log"

	"bierliste_backend/env"

	"github.com/jackc/pgx/v5"
)

func InitializeConnection() *pgx.Conn {
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s",
		env.DatabaseUser.GetValue(),
		env.DatabasePassword.GetValue(),
		env.DatabaseHost.GetValue(),
		env.DatabasePort.GetValue())
	conn, err := pgx.Connect(context.Background(), connectionString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())
	return conn
}
