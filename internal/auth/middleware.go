package auth

import (
	"net/http"
	"strings"

	"bierliste_backend/internal/entity"

	"github.com/gin-gonic/gin"
)

// UserIDKey is the gin.Context key under which the authenticated user's ID is stored.
const UserIDKey = "userId"

// UserRoleKey is the gin.Context key under which the authenticated user's role is stored.
const UserRoleKey = "userRole"

// ForcePasswordChangeKey is the gin.Context key indicating whether the
// authenticated user must change their password before doing anything else.
const ForcePasswordChangeKey = "forcePasswordChange"

// Middleware returns a Gin handler that validates the Bearer JWT on every request.
// On success it stores the user ID, role, and force-password-change flag in the
// context and calls Next. On failure it aborts with 401.
func Middleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "authorization header must be: Bearer <token>"})
			return
		}

		claims, err := parseToken(parts[1], secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "invalid or expired token"})
			return
		}

		c.Set(UserIDKey, claims.UserId)
		c.Set(UserRoleKey, claims.Role)
		c.Set(ForcePasswordChangeKey, claims.ForcePasswordChange)
		c.Next()
	}
}

// AdminMiddleware aborts with 403 if the authenticated user does not have the
// admin role. Must be chained after Middleware, which sets UserRoleKey.
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(UserRoleKey)
		if role != entity.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "admin access required"})
			return
		}
		c.Next()
	}
}

// PasswordChangedMiddleware aborts with 403 when the authenticated user's token
// carries forcePasswordChange = true. Place this after Middleware in the chain
// for any route that should be unavailable until the user sets a new password.
// The change-password endpoint itself must NOT be behind this middleware.
func PasswordChangedMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		force, _ := c.Get(ForcePasswordChangeKey)
		if force == true {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "you must change your password before continuing",
			})
			return
		}
		c.Next()
	}
}
