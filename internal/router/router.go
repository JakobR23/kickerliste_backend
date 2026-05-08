package router

import (
	"github.com/gin-gonic/gin"
)

// RegisterFunc registers routes onto a router group.
// Each domain package provides one by calling its own RegisterHandlers.
type RegisterFunc func(rg *gin.RouterGroup)

// New builds and returns the application router.
// All routes are mounted under /api/v1.
// Pass one RegisterFunc per domain to mount its routes.
func New(registrars ...RegisterFunc) *gin.Engine {
	r := gin.Default()
	v1 := r.Group("/api/v1")
	for _, register := range registrars {
		register(v1)
	}
	return r
}
