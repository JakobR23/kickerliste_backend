package main

import (
	"bierliste_backend/env"
	"bierliste_backend/internal/database"
	"bierliste_backend/internal/router"
	"fmt"
)

func main() {
	env.LoadConfig()
	database.InitializeConnection()
	router := router.New()
	host := fmt.Sprintf("%s:%s",
		env.Host.GetValue(),
		env.Port.GetValue())
	router.Run(host)
	router.Run("localhost:8080")
}
