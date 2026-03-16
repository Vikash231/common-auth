package main

import (
	"common-auth/internal/auth/db"
	"common-auth/internal/auth/email"
	"common-auth/internal/auth/handlers"
	"common-auth/internal/auth/middleware"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize subsystems
	db.Init()
	email.Init()
	handlers.InitOIDC()

	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{getEnv("APP_BASE_URL", "http://localhost:5173")},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Public routes
	auth := r.Group("/auth")
	{
		auth.POST("/signup", handlers.Signup)
		auth.POST("/set-password", handlers.SetPassword)
		auth.POST("/login", handlers.Login)
		auth.POST("/logout", handlers.Logout)
		auth.POST("/validate", handlers.Validate)

		// OIDC (Google)
		auth.GET("/oidc/login", handlers.OIDCLogin)
		auth.GET("/oidc/callback", handlers.OIDCCallback)

		// SAML (Okta)
		auth.GET("/saml/login", handlers.SAMLLogin)
		auth.POST("/saml/acs", handlers.SAMLCallback)
		auth.GET("/saml/metadata", handlers.SAMLMetadata)
	}

	// Authenticated routes
	protected := r.Group("/auth")
	protected.Use(middleware.RequireAuth())
	{
		protected.GET("/me", handlers.Me)
		protected.GET("/users", handlers.ListUsers)
		protected.GET("/groups", handlers.ListGroups)
	}

	// Admin-only routes
	admin := r.Group("/auth/admin")
	admin.Use(middleware.RequireAuth(), middleware.RequireAdmin())
	{
		admin.POST("/invite", handlers.AdminInviteUser)
		admin.PUT("/users/:id/group", handlers.UpdateUserGroup)
		admin.DELETE("/users/:id", handlers.DeleteUser)
		admin.POST("/groups", handlers.CreateGroup)
		admin.PUT("/groups/:id", handlers.UpdateGroup)
	}

	port := getEnv("PORT", "8083")
	log.Printf("Auth service starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
