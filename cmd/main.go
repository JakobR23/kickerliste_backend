package main

import (
	"context"
	"fmt"

	"bierliste_backend/env"
	"bierliste_backend/internal/database"
	"bierliste_backend/internal/fixture"
	"bierliste_backend/internal/router"
	"bierliste_backend/internal/team"
	"bierliste_backend/internal/teammember"
	"bierliste_backend/internal/user"
)

func main() {
	env.LoadConfig()

	conn := database.InitializeConnection()
	defer conn.Close(context.Background())

	db := database.New(conn)

	_ = user.NewRepository(db)
	_ = team.NewRepository(db)
	_ = teammember.NewRepository(db)
	_ = fixture.NewRepository(db)

	r := router.New()
	host := fmt.Sprintf("%s:%s", env.Host.GetValue(), env.Port.GetValue())
	r.Run(host)
}
