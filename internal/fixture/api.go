package fixture

import (
	"errors"
	"net/http"
	"strconv"

	"bierliste_backend/internal/auth"
	"bierliste_backend/internal/httputil"

	"github.com/gin-gonic/gin"
)

type resource struct {
	service Service
}

// RegisterHandlers mounts the fixture routes onto the given router group.
func RegisterHandlers(rg *gin.RouterGroup, service Service) {
	r := resource{service}

	fixtures := rg.Group("/fixtures")
	fixtures.GET("", r.list)
	fixtures.POST("", r.create)
	fixtures.GET("/:id", r.get)
	fixtures.PUT("/:id", r.update)
	fixtures.DELETE("/:id", r.delete)
}

// list handles GET /fixtures?teamId=<int>
func (r resource) list(c *gin.Context) {
	var teamId *int
	if raw := c.Query("teamId"); raw != "" {
		id, err := strconv.Atoi(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid teamId"})
			return
		}
		teamId = &id
	}

	fixtures, err := r.service.GetAll(c.Request.Context(), teamId)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, fixtures)
}

// get handles GET /fixtures/:id
func (r resource) get(c *gin.Context) {
	id, ok := httputil.ParseID(c)
	if !ok {
		return
	}
	f, err := r.service.GetById(c.Request.Context(), id)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, f)
}

// create handles POST /fixtures
// The fixture is created with status 'pending' and attributed to the caller.
func (r resource) create(c *gin.Context) {
	submittedBy := c.MustGet(auth.UserIDKey).(int)

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	f, err := r.service.Create(c.Request.Context(), submittedBy, req)
	if err != nil {
		if errors.Is(err, ErrSameTeam) || errors.Is(err, ErrScoresInconsistent) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
			return
		}
		httputil.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, f)
}

// update handles PUT /fixtures/:id
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
	f, err := r.service.Update(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, ErrScoresInconsistent) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
			return
		}
		httputil.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, f)
}

// delete handles DELETE /fixtures/:id
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
