package main

import (
	"context"
	"fmt"

	"bierliste_backend/env"
	"bierliste_backend/internal/auth"
	"bierliste_backend/internal/database"
	fixtureRepo "bierliste_backend/internal/fixture"
	"bierliste_backend/internal/router"
	adjustmentRepo "bierliste_backend/internal/scoreadjustment"
	teamRepo "bierliste_backend/internal/team"
	"bierliste_backend/internal/teammember"
	userRepo "bierliste_backend/internal/user"
	"bierliste_backend/migrations"

	"github.com/gin-gonic/gin"
)

func main() {
	env.LoadConfig()

	database.RunMigrations(migrations.FS)

	conn := database.InitializeConnection()
	defer conn.Close(context.Background())

	db := database.New(conn)

	users := userRepo.NewRepository(db)
	teams := teamRepo.NewRepository(db)
	tms := teammember.NewRepository(db)
	fixtures := fixtureRepo.NewRepository(db)
	adjustments := adjustmentRepo.NewRepository(db)

	jwtSecret := env.JWTSecret.GetValue()
	authMiddleware := auth.Middleware(jwtSecret)
	adminMiddleware := auth.AdminMiddleware()

	authService := auth.NewService(users, jwtSecret)
	userService := userRepo.NewService(users)
	teamService := teamRepo.NewService(teams, tms, users)
	fixtureService := fixtureRepo.NewService(fixtures)
	adjustmentService := adjustmentRepo.NewService(adjustments)

	r := router.New(
		func(rg *gin.RouterGroup) { auth.RegisterHandlers(rg, authService) },
		func(rg *gin.RouterGroup) {
			protected := rg.Group("", authMiddleware)
			userRepo.RegisterHandlers(protected, userService)
			teamRepo.RegisterHandlers(protected, teamService)
			fixtureRepo.RegisterHandlers(protected, fixtureService)

			admin := protected.Group("", adminMiddleware)
			adjustmentRepo.RegisterHandlers(admin, adjustmentService)
		},
	)

	host := fmt.Sprintf("%s:%s", env.Host.GetValue(), env.Port.GetValue())
	r.Run(host)
}
