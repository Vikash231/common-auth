package db

import (
	"common-auth/internal/auth/models"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// buildPostgresDSN returns a PostgreSQL connection string.
// Use DATABASE_URL for a full URL, or set DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME (default: commonauth), DB_SSLMODE (default: disable).
func buildPostgresDSN() string {
	if u := os.Getenv("DATABASE_URL"); u != "" {
		return u
	}
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "host.docker.internal"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}
	name := os.Getenv("DB_NAME")
	if name == "" {
		name = "commonauth"
	}
	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, name, sslmode)
}

func Init() {
	dsn := buildPostgresDSN()

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := DB.AutoMigrate(
		&models.User{},
		&models.PermissionGroup{},
		&models.EmailToken{},
	); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	seedDefaults()
}

func seedDefaults() {
	// Create default permission group
	defaultGroup := models.PermissionGroup{}
	if err := DB.Where("name = ?", "default").First(&defaultGroup).Error; err != nil {
		defaultGroup = models.PermissionGroup{
			ID:                uuid.NewString(),
			Name:              "default",
			CanUseWeatherAPI:  true,
			CanUseDistanceAPI: true,
		}
		DB.Create(&defaultGroup)
	}

	// Create restricted group (no permissions)
	restrictedGroup := models.PermissionGroup{}
	if err := DB.Where("name = ?", "restricted").First(&restrictedGroup).Error; err != nil {
		restrictedGroup = models.PermissionGroup{
			ID:                uuid.NewString(),
			Name:              "restricted",
			CanUseWeatherAPI:  false,
			CanUseDistanceAPI: false,
		}
		DB.Create(&restrictedGroup)
	}

	// Create admin user
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@example.com"
	}
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "Admin@12345"
	}

	admin := models.User{}
	if err := DB.Where("email = ?", adminEmail).First(&admin).Error; err != nil {
		hash, _ := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		admin = models.User{
			ID:            uuid.NewString(),
			Email:         adminEmail,
			Name:          "Admin",
			PasswordHash:  string(hash),
			IsAdmin:       true,
			IsActive:      true,
			EmailVerified: true,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		DB.Create(&admin)
		log.Printf("Admin user created: %s / %s", adminEmail, adminPassword)
	}
}
