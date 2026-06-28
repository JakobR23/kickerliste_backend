package scoreadjustment

import (
	"net/http"

	"bierliste_backend/internal/auth"
	"bierliste_backend/internal/httputil"

	"github.com/gin-gonic/gin"
)

type resource struct {
	service Service
}

// RegisterHandlers mounts the score adjustment routes onto rg.
// rg must already have the auth middleware and admin middleware applied.
func RegisterHandlers(rg *gin.RouterGroup, service Service) {
	r := resource{service}

	users := rg.Group("/users/:id")
	users.GET("/adjustments", r.list)
	users.POST("/adjustments", r.create)
}

// list handles GET /users/:id/adjustments
func (r resource) list(c *gin.Context) {
	userId, ok := httputil.ParseID(c)
	if !ok {
		return
	}
	adjustments, err := r.service.GetByUserId(c.Request.Context(), userId)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, adjustments)
}

// create handles POST /users/:id/adjustments
func (r resource) create(c *gin.Context) {
	targetUserId, ok := httputil.ParseID(c)
	if !ok {
		return
	}
	adminUserId := auth.UserID(c)

	var req CreateRequest
	if !httputil.BindJSON(c, &req) {
		return
	}
	a, err := r.service.Create(c.Request.Context(), adminUserId, targetUserId, req)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, a)
}
