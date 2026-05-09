package router

import (
	"time"

	"bierliste_backend/env"
	"bierliste_backend/internal/logger"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// RegisterFunc registers routes onto a router group.
// Each domain package provides one by calling its own RegisterHandlers.
type RegisterFunc func(rg *gin.RouterGroup)

// New builds and returns the application router.
// All routes are mounted under /api/v1.
// Pass one RegisterFunc per domain to mount its routes.
func New(registrars ...RegisterFunc) *gin.Engine {
	r := gin.New()

	// Trust no proxies — use the direct connection's remote address as the
	// client IP. If a reverse proxy (nginx, Traefik, etc.) is added in front
	// of this service, set this to the proxy's IP or CIDR range instead so
	// that X-Forwarded-For headers are read correctly.
	r.SetTrustedProxies(nil) //nolint:errcheck

	r.Use(
		cors.New(cors.Config{
			AllowOrigins:     env.GetAllowedOrigins(),
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
			AllowHeaders:     []string{"Authorization", "Content-Type"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}),
		logger.RequestLogger(),
		logger.Recovery(),
	)

	v1 := r.Group("/api/v1")
	for _, register := range registrars {
		register(v1)
	}
	return r
}
