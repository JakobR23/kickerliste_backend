package router

import (
	"time"

	"bierliste_backend/env"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// RegisterFunc registers routes onto a router group.
// Each domain package provides one by calling its own RegisterHandlers.
type RegisterFunc func(rg *gin.RouterGroup)

// New builds and returns the application router.
// Pass one RegisterFunc per domain to mount its routes.
func New(registrars ...RegisterFunc) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     env.GetAllowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	for _, register := range registrars {
		register(&r.RouterGroup)
	}
	return r
}
