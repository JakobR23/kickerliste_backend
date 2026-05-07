package user

import (
	"net/http"

	"bierliste_backend/internal/httputil"

	"github.com/gin-gonic/gin"
)

type resource struct {
	service Service
}

// RegisterHandlers mounts the user routes onto the given router group.
func RegisterHandlers(rg *gin.RouterGroup, service Service) {
	r := resource{service}

	users := rg.Group("/users")
	users.GET("", r.list)
	users.POST("", r.create)
	users.GET("/:id", r.get)
	users.PUT("/:id", r.update)
	users.DELETE("/:id", r.delete)
}

// list handles GET /users
func (r resource) list(c *gin.Context) {
	users, err := r.service.GetAll(c.Request.Context())
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, users)
}

// get handles GET /users/:id
func (r resource) get(c *gin.Context) {
	id, ok := httputil.ParseID(c)
	if !ok {
		return
	}
	u, err := r.service.GetById(c.Request.Context(), id)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, u)
}

// create handles POST /users
func (r resource) create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	u, err := r.service.Create(c.Request.Context(), req)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, u)
}

// update handles PUT /users/:id
func (r resource) update(c *gin.Context) {
	id, ok := httputil.ParseID(c)
	if !ok {
		return
	}
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	u, err := r.service.Update(c.Request.Context(), id, req)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, u)
}

// delete handles DELETE /users/:id
func (r resource) delete(c *gin.Context) {
	id, ok := httputil.ParseID(c)
	if !ok {
		return
	}
	if err := r.service.Delete(c.Request.Context(), id); err != nil {
		httputil.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
