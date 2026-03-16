package db

import (
	"common-auth/internal/auth/models"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() {
	dsn := os.Getenv("DB_PATH")
	if dsn == "" {
		dsn = "./auth.db"
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
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
