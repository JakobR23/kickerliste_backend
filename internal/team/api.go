package team

import (
	"errors"
	"net/http"
	"strconv"

	"bierliste_backend/internal/httputil"

	"github.com/gin-gonic/gin"
)

type resource struct {
	service Service
}

// RegisterHandlers mounts the team and team-member routes available to all
// authenticated users onto the given router group.
func RegisterHandlers(rg *gin.RouterGroup, service Service) {
	r := resource{service}

	teams := rg.Group("/teams")
	teams.GET("", r.list)
	teams.GET("/with-members", r.listWithMembers)
	teams.POST("", r.create)
	teams.GET("/:id", r.get)
	teams.PUT("/:id", r.update)
	teams.GET("/:id/members", r.listMembers)
	teams.POST("/:id/members", r.addMember)
}

// RegisterAdminHandlers mounts the team and team-member routes that require
// admin privileges onto the given router group.
func RegisterAdminHandlers(rg *gin.RouterGroup, service Service) {
	r := resource{service}

	teams := rg.Group("/teams")
	teams.DELETE("/:id", r.delete)
	teams.DELETE("/:id/members/:userId", r.removeMember)
}

// listWithMembers handles GET /teams/with-members
func (r resource) listWithMembers(c *gin.Context) {
	teams, err := r.service.GetAllWithMembers(c.Request.Context())
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, teams)
}

// list handles GET /teams
func (r resource) list(c *gin.Context) {
	teams, err := r.service.GetAll(c.Request.Context())
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, teams)
}

// get handles GET /teams/:id
func (r resource) get(c *gin.Context) {
	id, ok := httputil.ParseID(c)
	if !ok {
		return
	}
	t, err := r.service.GetById(c.Request.Context(), id)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, t)
}

// create handles POST /teams
func (r resource) create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	t, err := r.service.Create(c.Request.Context(), req)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, t)
}

// update handles PUT /teams/:id
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
	t, err := r.service.Update(c.Request.Context(), id, req)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, t)
}

// delete handles DELETE /teams/:id
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

// listMembers handles GET /teams/:id/members
func (r resource) listMembers(c *gin.Context) {
	id, ok := httputil.ParseID(c)
	if !ok {
		return
	}
	members, err := r.service.GetMembers(c.Request.Context(), id)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, members)
}

// addMember handles POST /teams/:id/members
func (r resource) addMember(c *gin.Context) {
	id, ok := httputil.ParseID(c)
	if !ok {
		return
	}
	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	u, err := r.service.AddMember(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, ErrTeamFull) {
			c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
			return
		}
		httputil.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, u)
}

// removeMember handles DELETE /teams/:id/members/:userId
func (r resource) removeMember(c *gin.Context) {
	id, ok := httputil.ParseID(c)
	if !ok {
		return
	}
	userId, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid userId"})
		return
	}
	if err := r.service.RemoveMember(c.Request.Context(), id, userId); err != nil {
		httputil.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
