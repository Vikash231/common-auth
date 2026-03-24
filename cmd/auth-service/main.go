package main

import (
	"common-auth/internal/auth/db"
	"common-auth/internal/auth/email"
	"common-auth/internal/auth/handlers"
	"common-auth/internal/auth/middleware"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// loadDotEnv tries .env in cwd, then parent dirs (e.g. when debugging from cmd/auth-service).
func loadDotEnv() string {
	candidates := []string{
		".env",
		filepath.Join("..", ".env"),
		filepath.Join("..", "..", ".env"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err != nil {
			continue
		}
		if err := godotenv.Load(p); err != nil {
			log.Printf("godotenv: error loading %s: %v", p, err)
			continue
		}
		return p
	}
	return ""
}

func main() {
	cwd, _ := os.Getwd()
	loadedFrom := loadDotEnv()

	if loadedFrom == "" {
		log.Printf("godotenv: no .env found in ., .., or ../.. (cwd=%s)", cwd)
	}

	// Initialize subsystems
	db.Init()
	email.Init()
	handlers.InitOIDC()
	handlers.InitSAML()

	r := gin.Default()

	// CORS
	frontendURL := getEnv("APP_BASE_URL", "http://localhost:5173")
	frontendHost := ""
	if u, err := url.Parse(frontendURL); err == nil {
		frontendHost = u.Hostname()
	}
	idpHost := ""
	if idpURL := os.Getenv("SAML_IDP_METADATA_URL"); idpURL != "" {
		if u, err := url.Parse(idpURL); err == nil {
			idpHost = u.Hostname()
		}
	}
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{frontendURL},
		AllowOriginFunc: func(origin string) bool {
			if origin == "null" {
				return true
			}
			u, err := url.Parse(origin)
			if err != nil {
				return false
			}
			host := u.Hostname()
			if frontendHost != "" && host == frontendHost {
				return true
			}
			if host == "localhost" || host == "127.0.0.1" {
				return true
			}
			if idpHost != "" && host == idpHost {
				return true
			}
			if strings.HasSuffix(host, ".auth0.com") || strings.HasSuffix(host, ".okta.com") {
				return true
			}
			return false
		},
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
