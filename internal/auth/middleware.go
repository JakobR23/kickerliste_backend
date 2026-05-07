package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// UserIDKey is the gin.Context key under which the authenticated user's ID is stored.
const UserIDKey = "userId"

// UserRoleKey is the gin.Context key under which the authenticated user's role is stored.
const UserRoleKey = "userRole"

// Middleware returns a Gin handler that validates the Bearer JWT on every request.
// On success it stores the user ID in the context under UserIDKey and calls Next.
// On failure it aborts with 401.
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
		c.Next()
	}
}
