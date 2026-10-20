package scorehistory

import (
	"net/http"

	"bierliste_backend/internal/auth"
	"bierliste_backend/internal/entity"
	"bierliste_backend/internal/httputil"

	"github.com/gin-gonic/gin"
)

type resource struct {
	service Service
}

// RegisterHandlers mounts the score-history route onto rg. rg must already have
// the auth + password-changed middleware applied. The route is available to all
// authenticated users; the self-or-admin rule is enforced per request in the
// handler (admins may view anyone's history).
func RegisterHandlers(rg *gin.RouterGroup, service Service) {
	r := resource{service}

	users := rg.Group("/users/:id")
	users.GET("/score-history", r.list)
}

// list handles GET /users/:id/score-history
func (r resource) list(c *gin.Context) {
	id, ok := httputil.ParseID(c)
	if !ok {
		return
	}

	// Self-or-admin. Authorise before any existence check so a non-admin cannot
	// probe which user IDs exist.
	if auth.UserID(c) != id && auth.UserRole(c) != entity.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"message": "you may only view your own score history"})
		return
	}

	entries, err := r.service.GetByUserId(c.Request.Context(), id)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, entries)
}
