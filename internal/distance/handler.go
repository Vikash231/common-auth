package distance

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

type DistanceData struct {
	From           string  `json:"from"`
	To             string  `json:"to"`
	DistanceKm     float64 `json:"distance_km"`
	DrivingTimeMin int     `json:"driving_time_minutes"`
	TransitTimeMin int     `json:"transit_time_minutes"`
	WalkingTimeMin int     `json:"walking_time_minutes"`
	FetchedAt      string  `json:"fetched_at"`
}

// Handler validates JWT and returns mock distance/travel time data.
func Handler(c *gin.Context) {
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

	hasPerm := false
	for _, p := range claims.Permissions {
		if p == "distance:read" {
			hasPerm = true
			break
		}
	}
	if !hasPerm {
		common.Forbidden(c, "no permission to access distance API")
		return
	}

	from := strings.TrimSpace(c.Query("from"))
	to := strings.TrimSpace(c.Query("to"))

	if from == "" || to == "" {
		common.BadRequest(c, "both 'from' and 'to' parameters are required")
		return
	}

	data := calculateDistance(from, to)
	common.OK(c, data)
}

func calculateDistance(from, to string) DistanceData {
	// Deterministic mock based on place names
	seed := int64(0)
	for _, ch := range from + to {
		seed += int64(ch)
	}
	r := rand.New(rand.NewSource(seed))

	distanceKm := 10.0 + r.Float64()*990.0
	distanceKm = math.Round(distanceKm*10) / 10

	// Approximate travel times
	avgDrivingSpeed := 60.0  // km/h
	avgTransitSpeed := 40.0  // km/h
	avgWalkingSpeed := 5.0   // km/h

	drivingMin := int(distanceKm / avgDrivingSpeed * 60)
	transitMin := int(distanceKm / avgTransitSpeed * 60)
	walkingMin := int(distanceKm / avgWalkingSpeed * 60)

	return DistanceData{
		From:           fmt.Sprintf("%s", from),
		To:             fmt.Sprintf("%s", to),
		DistanceKm:     distanceKm,
		DrivingTimeMin: drivingMin,
		TransitTimeMin: transitMin,
		WalkingTimeMin: walkingMin,
		FetchedAt:      time.Now().UTC().Format(time.RFC3339),
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
