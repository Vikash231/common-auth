package models

import (
	"time"
)

// User represents an application user.
type User struct {
	ID            string    `gorm:"primaryKey" json:"id"`
	Email         string    `gorm:"uniqueIndex;not null" json:"email"`
	Name          string    `gorm:"not null" json:"name"`
	PasswordHash  string    `json:"-"`
	IsAdmin       bool      `gorm:"default:false" json:"is_admin"`
	IsActive      bool      `gorm:"default:false" json:"is_active"`
	EmailVerified bool      `gorm:"default:false" json:"email_verified"`
	SAMLID        string    `gorm:"index" json:"-"`
	OIDCID        string    `gorm:"index" json:"-"`
	GroupID       *string   `json:"group_id"`
	Group         *PermissionGroup `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PermissionGroup defines what APIs a user can access.
type PermissionGroup struct {
	ID                 string    `gorm:"primaryKey" json:"id"`
	Name               string    `gorm:"uniqueIndex;not null" json:"name"`
	CanUseWeatherAPI   bool      `gorm:"default:true" json:"can_use_weather_api"`
	CanUseDistanceAPI  bool      `gorm:"default:true" json:"can_use_distance_api"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// EmailToken is used for signup verification and admin invites.
type EmailToken struct {
	ID        string    `gorm:"primaryKey"`
	UserID    string    `gorm:"not null;index"`
	Token     string    `gorm:"uniqueIndex;not null"`
	TokenType string    `gorm:"not null"` // "signup" | "invite" | "password_reset"
	ExpiresAt time.Time `gorm:"not null"`
	Used      bool      `gorm:"default:false"`
	CreatedAt time.Time
}

// Permissions returns the slice of permission strings for a user based on their group.
func (u *User) Permissions() []string {
	perms := []string{}
	if u.Group == nil {
		return perms
	}
	if u.Group.CanUseWeatherAPI {
		perms = append(perms, "weather:read")
	}
	if u.Group.CanUseDistanceAPI {
		perms = append(perms, "distance:read")
	}
	return perms
}
