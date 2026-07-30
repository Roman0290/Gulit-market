package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/romina/pocket-market-api/internal/users"
)

const (
	ContextUserIDKey = "user_id"
	ContextRoleKey   = "role"
)

func (s *Service) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or malformed authorization header"})
			return
		}

		claims, err := s.ParseToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// Re-checked on every request (not just at login) so a suspension
		// takes effect immediately instead of waiting for the JWT to expire.
		u, err := s.userRepo.GetByID(c.Request.Context(), claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		if u.Status == users.StatusSuspended {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": ErrAccountSuspended.Error()})
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextRoleKey, claims.Role)
		c.Next()
	}
}

func RequireRole(roles ...users.Role) gin.HandlerFunc {
	allowed := make(map[users.Role]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(c *gin.Context) {
		role, _ := c.Get(ContextRoleKey)
		r, _ := role.(users.Role)

		if !allowed[r] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}
