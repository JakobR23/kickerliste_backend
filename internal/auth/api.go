package auth

import (
	"errors"
	"net/http"

	"bierliste_backend/internal/user"

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
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	token, err := r.service.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// register handles POST /auth/register
func (r resource) register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	token, err := r.service.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, user.ErrUsernameTaken) {
			c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"token": token})
}

// changePassword handles POST /auth/change-password
func (r resource) changePassword(c *gin.Context) {
	userId := c.MustGet(UserIDKey).(int)

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	token, err := r.service.ChangePassword(c.Request.Context(), userId, req)
	if err != nil {
		if errors.Is(err, ErrPasswordMismatch) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
