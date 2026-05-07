package main

import (
	"context"
	"fmt"

	"bierliste_backend/env"
	"bierliste_backend/internal/auth"
	"bierliste_backend/internal/database"
	fixtureRepo "bierliste_backend/internal/fixture"
	"bierliste_backend/internal/router"
	teamRepo "bierliste_backend/internal/team"
	"bierliste_backend/internal/teammember"
	userRepo "bierliste_backend/internal/user"

	"github.com/gin-gonic/gin"
)

func main() {
	env.LoadConfig()

	conn := database.InitializeConnection()
	defer conn.Close(context.Background())

	db := database.New(conn)

	users := userRepo.NewRepository(db)
	teams := teamRepo.NewRepository(db)
	tms := teammember.NewRepository(db)
	fixtures := fixtureRepo.NewRepository(db)

	jwtSecret := env.JWTSecret.GetValue()
	authMiddleware := auth.Middleware(jwtSecret)

	authService := auth.NewService(users, jwtSecret)
	userService := userRepo.NewService(users)
	teamService := teamRepo.NewService(teams, tms, users)
	fixtureService := fixtureRepo.NewService(fixtures)

	r := router.New(
		// Public: login endpoint — no token required
		func(rg *gin.RouterGroup) { auth.RegisterHandlers(rg, authService) },
		// Protected: all other endpoints require a valid Bearer token
		func(rg *gin.RouterGroup) {
			protected := rg.Group("", authMiddleware)
			userRepo.RegisterHandlers(protected, userService)
			teamRepo.RegisterHandlers(protected, teamService)
			fixtureRepo.RegisterHandlers(protected, fixtureService)
		},
	)

	host := fmt.Sprintf("%s:%s", env.Host.GetValue(), env.Port.GetValue())
	r.Run(host)
}
