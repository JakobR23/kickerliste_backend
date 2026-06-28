package auth

import (
	"net/http"

	"bierliste_backend/internal/httputil"

	"github.com/gin-gonic/gin"
)

type resource struct {
	service Service
}

// RegisterPublicHandlers mounts the unauthenticated auth routes (login, register).
func RegisterPublicHandlers(rg *gin.RouterGroup, service Service) {
	r := resource{service}
	rg.POST("/auth/login", r.login)
	rg.POST("/auth/register", r.register)
}

// RegisterProtectedHandlers mounts auth routes that require a valid token but
// must NOT be behind the PasswordChangedMiddleware (e.g. change-password).
func RegisterProtectedHandlers(rg *gin.RouterGroup, service Service) {
	r := resource{service}
	rg.POST("/auth/change-password", r.changePassword)
}

// login handles POST /auth/login
func (r resource) login(c *gin.Context) {
	var req LoginRequest
	if !httputil.BindJSON(c, &req) {
		return
	}

	token, err := r.service.Login(c.Request.Context(), req)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// register handles POST /auth/register
// Returns 201 with a message on success — no token is issued until an admin
// activates the account.
func (r resource) register(c *gin.Context) {
	var req RegisterRequest
	if !httputil.BindJSON(c, &req) {
		return
	}

	if err := r.service.Register(c.Request.Context(), req); err != nil {
		httputil.HandleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "account created, pending admin activation"})
}

// changePassword handles POST /auth/change-password
func (r resource) changePassword(c *gin.Context) {
	userId := UserID(c)

	var req ChangePasswordRequest
	if !httputil.BindJSON(c, &req) {
		return
	}

	token, err := r.service.ChangePassword(c.Request.Context(), userId, req)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
