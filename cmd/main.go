package main

import (
	"fmt"

	"bierliste_backend/env"
	"bierliste_backend/internal/auth"
	"bierliste_backend/internal/database"
	fixtureRepo "bierliste_backend/internal/fixture"
	"bierliste_backend/internal/logger"
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
	logger.Setup()

	database.RunMigrations(migrations.FS)

	conn := database.InitializeConnection()
	defer conn.Close()

	db := database.New(conn)

	users := userRepo.NewRepository(db)
	teams := teamRepo.NewRepository(db)
	tms := teammember.NewRepository(db)
	fixtures := fixtureRepo.NewRepository(db)
	adjustments := adjustmentRepo.NewRepository(db)

	jwtSecret := env.JWTSecret.GetValue()
	authMiddleware := auth.Middleware(jwtSecret)
	adminMiddleware := auth.AdminMiddleware()
	passwordChangedMiddleware := auth.PasswordChangedMiddleware()

	authService := auth.NewService(users, jwtSecret)
	userService := userRepo.NewService(users)
	teamService := teamRepo.NewService(teams, tms, users)
	fixtureService := fixtureRepo.NewService(fixtures)
	adjustmentService := adjustmentRepo.NewService(adjustments)

	r := router.New(
		func(rg *gin.RouterGroup) {
			// Public: login and self-registration.
			auth.RegisterPublicHandlers(rg, authService)

			// Authenticated: change-password is intentionally NOT behind
			// passwordChangedMiddleware so users can always reach it.
			protected := rg.Group("", authMiddleware)
			auth.RegisterProtectedHandlers(protected, authService)

			// Authenticated + password already changed (normal usage).
			normal := protected.Group("", passwordChangedMiddleware)
			userRepo.RegisterHandlers(normal, userService)
			teamRepo.RegisterHandlers(normal, teamService)
			fixtureRepo.RegisterHandlers(normal, fixtureService)

			// Authenticated + password changed + admin only.
			admin := normal.Group("", adminMiddleware)
			userRepo.RegisterAdminHandlers(admin, userService)
			teamRepo.RegisterAdminHandlers(admin, teamService)
			fixtureRepo.RegisterAdminHandlers(admin, fixtureService)
			adjustmentRepo.RegisterHandlers(admin, adjustmentService)
		},
	)

	host := fmt.Sprintf("%s:%s", env.Host.GetValue(), env.Port.GetValue())
	r.Run(host)
}
