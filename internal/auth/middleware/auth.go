package middleware

import (
	"common-auth/internal/auth/db"
	"common-auth/internal/auth/models"
	"common-auth/internal/common"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// RequireAuth validates the JWT and loads the user into the context.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			common.Unauthorized(c, "missing authorization token")
			c.Abort()
			return
		}

		claims, err := common.ValidateToken(token, publicKeyPath())
		if err != nil {
			common.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		// Load full user from DB to get up-to-date permissions
		var user models.User
		if err := db.DB.Preload("Group").Where("id = ? AND is_active = ?", claims.UserID, true).First(&user).Error; err != nil {
			common.Unauthorized(c, "user not found or inactive")
			c.Abort()
			return
		}

		c.Set("user", &user)
		c.Set("claims", claims)
		c.Next()
	}
}

// RequireAdmin ensures the authenticated user is an admin.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		u, exists := c.Get("user")
		if !exists {
			common.Unauthorized(c, "authentication required")
			c.Abort()
			return
		}
		user := u.(*models.User)
		if !user.IsAdmin {
			common.Forbidden(c, "admin access required")
			c.Abort()
			return
		}
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	// Also check cookie
	if cookie, err := c.Cookie("auth_token"); err == nil {
		return cookie
	}
	return ""
}

func publicKeyPath() string {
	path := os.Getenv("PUBLIC_KEY_PATH")
	if path == "" {
		return "./keys/public.pem"
	}
	return path
}
