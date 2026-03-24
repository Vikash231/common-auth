package main

import (
	"common-auth/internal/gateway"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := gateway.LoadConfig()

	r := gin.Default()

	// CORS - allow frontend origin. SAML ACS POSTs come from the IdP page with Origin: https://tenant.auth0.com
	// (or Okta); without this, gin-contrib/cors rejects with 403 before the proxy runs.
	frontendURL := os.Getenv("APP_BASE_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
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
			// Some SAML form-post browser flows can send Origin: null.
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
			log.Printf("CORS rejected origin=%q host=%q", origin, host)
			return false
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "api-gateway"})
	})

	// All /api/* routes go through the proxy handler
	r.Any("/api/*path", gateway.Proxy(cfg))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("API Gateway starting on :%s", port)
	log.Printf("  Auth Service: %s", cfg.AuthServiceURL)
	log.Printf("  Weather Service: %s", cfg.WeatherServiceURL)
	log.Printf("  Distance Service: %s", cfg.DistanceServiceURL)

	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
