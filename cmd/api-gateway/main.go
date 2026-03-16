package main

import (
	"common-auth/internal/gateway"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := gateway.LoadConfig()

	r := gin.Default()

	// CORS - allow frontend origin
	frontendURL := os.Getenv("APP_BASE_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendURL},
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
