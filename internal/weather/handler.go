package weather

import (
	"common-auth/internal/common"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type WeatherData struct {
	Place       string  `json:"place"`
	Temperature float64 `json:"temperature_celsius"`
	Humidity    int     `json:"humidity_percent"`
	Condition   string  `json:"condition"`
	WindSpeed   float64 `json:"wind_speed_kmh"`
	Visibility  float64 `json:"visibility_km"`
	FetchedAt   string  `json:"fetched_at"`
}

var conditions = []string{
	"Sunny", "Partly Cloudy", "Overcast", "Light Rain",
	"Heavy Rain", "Thunderstorm", "Foggy", "Snowy", "Windy", "Clear",
}

// Handler validates JWT and returns mock weather data.
func Handler(c *gin.Context) {
	// Validate JWT
	token := extractToken(c)
	if token == "" {
		common.Unauthorized(c, "missing authorization token")
		return
	}

	pubKeyPath := os.Getenv("PUBLIC_KEY_PATH")
	if pubKeyPath == "" {
		pubKeyPath = "./keys/public.pem"
	}

	claims, err := common.ValidateToken(token, pubKeyPath)
	if err != nil {
		common.Unauthorized(c, "invalid or expired token")
		return
	}

	// Check weather permission
	hasPerm := false
	for _, p := range claims.Permissions {
		if p == "weather:read" {
			hasPerm = true
			break
		}
	}
	if !hasPerm {
		common.Forbidden(c, "no permission to access weather API")
		return
	}

	place := strings.TrimSpace(c.Query("place"))
	if place == "" {
		common.BadRequest(c, "place parameter is required")
		return
	}

	data := generateWeather(place)
	common.OK(c, data)
}

func generateWeather(place string) WeatherData {
	// Use place name as seed for deterministic-ish results
	seed := int64(0)
	for _, ch := range place {
		seed += int64(ch)
	}
	r := rand.New(rand.NewSource(seed + time.Now().Unix()/3600))

	temp := -10.0 + r.Float64()*50.0
	humidity := 20 + r.Intn(80)
	condition := conditions[r.Intn(len(conditions))]
	windSpeed := r.Float64() * 80
	visibility := 1.0 + r.Float64()*19.0

	return WeatherData{
		Place:       fmt.Sprintf("%s", strings.Title(strings.ToLower(place))),
		Temperature: math.Round(temp*10) / 10,
		Humidity:    humidity,
		Condition:   condition,
		WindSpeed:   math.Round(windSpeed*10) / 10,
		Visibility:  math.Round(visibility*10) / 10,
		FetchedAt:   time.Now().UTC().Format(time.RFC3339),
	}
}

func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	if cookie, err := c.Cookie("auth_token"); err == nil {
		return cookie
	}
	return ""
}
