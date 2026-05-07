package router

import (
	"github.com/gin-gonic/gin"
)

// RegisterFunc registers routes onto a router group.
// Each domain package provides one by calling its own RegisterHandlers.
type RegisterFunc func(rg *gin.RouterGroup)

// New builds and returns the application router.
// Pass one RegisterFunc per domain to mount its routes.
func New(registrars ...RegisterFunc) *gin.Engine {
	r := gin.Default()
	for _, register := range registrars {
		register(&r.RouterGroup)
	}
	return r
}
