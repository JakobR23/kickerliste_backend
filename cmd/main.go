package main

import (
	"context"
	"fmt"

	"bierliste_backend/env"
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

	userService := userRepo.NewService(users)
	teamService := teamRepo.NewService(teams, tms, users)
	fixtureService := fixtureRepo.NewService(fixtures)

	r := router.New(
		func(rg *gin.RouterGroup) { userRepo.RegisterHandlers(rg, userService) },
		func(rg *gin.RouterGroup) { teamRepo.RegisterHandlers(rg, teamService) },
		func(rg *gin.RouterGroup) { fixtureRepo.RegisterHandlers(rg, fixtureService) },
	)

	host := fmt.Sprintf("%s:%s", env.Host.GetValue(), env.Port.GetValue())
	r.Run(host)
}
