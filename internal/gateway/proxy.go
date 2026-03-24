package gateway

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type Config struct {
	AuthServiceURL     string
	WeatherServiceURL  string
	DistanceServiceURL string
}

func LoadConfig() Config {
	return Config{
		AuthServiceURL:     getEnv("AUTH_SERVICE_URL", "http://localhost:8083"),
		WeatherServiceURL:  getEnv("WEATHER_SERVICE_URL", "http://localhost:8081"),
		DistanceServiceURL: getEnv("DISTANCE_SERVICE_URL", "http://localhost:8082"),
	}
}

type validateResponse struct {
	Valid   bool   `json:"valid"`
	Allowed bool   `json:"allowed"`
	UserID  string `json:"user_id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	IsAdmin bool   `json:"is_admin"`
	Error   string `json:"error"`
}

// Proxy creates the gateway handler that routes requests to the appropriate service.
func Proxy(cfg Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Auth service routes pass through directly (no token check for auth endpoints)
		if strings.HasPrefix(path, "/api/auth/") {
			targetPath := strings.TrimPrefix(path, "/api")
			proxyTo(c, cfg.AuthServiceURL, targetPath)
			return
		}

		// All other /api/* routes require auth
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization token"})
			return
		}

		var requiredPermission string
		var targetURL string
		var targetPath string

		switch {
		case strings.HasPrefix(path, "/api/weather"):
			requiredPermission = "weather:read"
			targetURL = cfg.WeatherServiceURL
			targetPath = strings.TrimPrefix(path, "/api")
		case strings.HasPrefix(path, "/api/distance"):
			requiredPermission = "distance:read"
			targetURL = cfg.DistanceServiceURL
			targetPath = strings.TrimPrefix(path, "/api")
		default:
			c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
			return
		}

		// Validate token and check permission via auth service
		valid, userInfo, statusCode := validateToken(cfg.AuthServiceURL, token, requiredPermission)
		if !valid {
			if statusCode == http.StatusForbidden {
				c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			}
			return
		}

		// Forward user info to downstream service via headers
		c.Request.Header.Set("X-User-ID", userInfo.UserID)
		c.Request.Header.Set("X-User-Email", userInfo.Email)
		c.Request.Header.Set("X-User-Name", userInfo.Name)
		if userInfo.IsAdmin {
			c.Request.Header.Set("X-User-Admin", "true")
		}

		proxyTo(c, targetURL, targetPath)
	}
}

func validateToken(authURL, token, permission string) (bool, validateResponse, int) {
	body, _ := json.Marshal(map[string]string{
		"token":               token,
		"required_permission": permission,
	})

	resp, err := http.Post(authURL+"/auth/validate", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Printf("auth service error: %v", err)
		return false, validateResponse{}, http.StatusUnauthorized
	}
	defer resp.Body.Close()

	var result validateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, result, http.StatusUnauthorized
	}

	if resp.StatusCode == http.StatusOK && result.Valid && result.Allowed {
		return true, result, http.StatusOK
	}

	return false, result, resp.StatusCode
}

func proxyTo(c *gin.Context, targetBase, targetPath string) {
	target, err := url.Parse(targetBase)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "invalid upstream URL"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy error: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(w, `{"error":"upstream service unavailable"}`)
	}

	// Preserve original external request details for upstream validation (important for SAML ACS).
	origPath := c.Request.URL.Path
	origHost := c.Request.Host
	origProto := "http"
	if c.Request.TLS != nil {
		origProto = "https"
	}
	if xfProto := c.GetHeader("X-Forwarded-Proto"); xfProto != "" {
		origProto = xfProto
	}
	c.Request.Header.Set("X-Forwarded-Host", origHost)
	c.Request.Header.Set("X-Forwarded-Proto", origProto)
	c.Request.Header.Set("X-Forwarded-Uri", origPath)

	// Rewrite the path
	c.Request.URL.Path = targetPath
	c.Request.URL.Host = target.Host
	c.Request.URL.Scheme = target.Scheme

	proxy.ServeHTTP(c.Writer, c.Request)
}

func extractToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	if cookie, err := c.Cookie("auth_token"); err == nil {
		return cookie
	}
	return ""
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
