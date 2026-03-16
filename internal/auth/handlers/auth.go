package handlers

import (
	"common-auth/internal/auth/db"
	"common-auth/internal/auth/email"
	"common-auth/internal/auth/models"
	"common-auth/internal/common"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// POST /auth/signup
// Body: { "email": "...", "name": "..." }
func Signup(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Name  string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, err.Error())
		return
	}

	// Check if user already exists
	var existing models.User
	if db.DB.Where("email = ?", req.Email).First(&existing).Error == nil {
		common.BadRequest(c, "email already registered")
		return
	}

	// Find default permission group
	var group models.PermissionGroup
	if err := db.DB.Where("name = ?", "default").First(&group).Error; err != nil {
		common.InternalError(c, "permission group not found")
		return
	}

	userID := uuid.NewString()
	user := models.User{
		ID:        userID,
		Email:     req.Email,
		Name:      req.Name,
		IsActive:  false,
		GroupID:   &group.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.DB.Create(&user).Error; err != nil {
		common.InternalError(c, "failed to create user")
		return
	}

	// Create email verification token
	token := uuid.NewString()
	emailToken := models.EmailToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		Token:     token,
		TokenType: "signup",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
	db.DB.Create(&emailToken)

	if err := email.SendVerificationEmail(req.Email, req.Name, token); err != nil {
		// Non-fatal: log and continue
		_ = err
	}

	common.Created(c, gin.H{"message": "verification email sent", "email": req.Email})
}

// POST /auth/set-password
// Body: { "token": "...", "password": "..." }
func SetPassword(c *gin.Context) {
	var req struct {
		Token    string `json:"token" binding:"required"`
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, err.Error())
		return
	}

	var emailToken models.EmailToken
	if err := db.DB.Where("token = ? AND used = ? AND expires_at > ?", req.Token, false, time.Now()).
		First(&emailToken).Error; err != nil {
		common.BadRequest(c, "invalid or expired token")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		common.InternalError(c, "failed to hash password")
		return
	}

	// Activate the user
	if err := db.DB.Model(&models.User{}).Where("id = ?", emailToken.UserID).Updates(map[string]interface{}{
		"password_hash":  string(hash),
		"is_active":     true,
		"email_verified": true,
		"updated_at":    time.Now(),
	}).Error; err != nil {
		common.InternalError(c, "failed to set password")
		return
	}

	// Mark token as used
	db.DB.Model(&emailToken).Update("used", true)

	common.OK(c, gin.H{"message": "password set successfully, you can now login"})
}

// POST /auth/login
// Body: { "email": "...", "password": "..." }
func Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, err.Error())
		return
	}

	var user models.User
	if err := db.DB.Preload("Group").Where("email = ?", req.Email).First(&user).Error; err != nil {
		common.Unauthorized(c, "invalid email or password")
		return
	}

	if !user.IsActive || !user.EmailVerified {
		common.Unauthorized(c, "account not activated, check your email")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		common.Unauthorized(c, "invalid email or password")
		return
	}

	tokenStr, err := issueToken(&user)
	if err != nil {
		common.InternalError(c, "failed to issue token")
		return
	}

	// Set cookie
	c.SetCookie("auth_token", tokenStr, 3600*24, "/", "", false, true)

	common.OK(c, gin.H{
		"token": tokenStr,
		"user": gin.H{
			"id":       user.ID,
			"email":    user.Email,
			"name":     user.Name,
			"is_admin": user.IsAdmin,
		},
	})
}

// POST /auth/logout
func Logout(c *gin.Context) {
	c.SetCookie("auth_token", "", -1, "/", "", false, true)
	common.OK(c, gin.H{"message": "logged out"})
}

// GET /auth/me
func Me(c *gin.Context) {
	u, _ := c.Get("user")
	user := u.(*models.User)
	common.OK(c, gin.H{
		"id":          user.ID,
		"email":       user.Email,
		"name":        user.Name,
		"is_admin":    user.IsAdmin,
		"permissions": user.Permissions(),
		"group":       user.Group,
	})
}

// POST /auth/validate  - called by API Gateway
// Body: { "token": "...", "required_permission": "..." }
func Validate(c *gin.Context) {
	var req struct {
		Token              string `json:"token" binding:"required"`
		RequiredPermission string `json:"required_permission"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, err.Error())
		return
	}

	claims, err := common.ValidateToken(req.Token, publicKeyPath())
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token", "valid": false})
		return
	}

	// Load user from DB to get live permissions
	var user models.User
	if err := db.DB.Preload("Group").Where("id = ? AND is_active = ?", claims.UserID, true).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found or inactive", "valid": false})
		return
	}

	// Check permission if specified
	if req.RequiredPermission != "" {
		hasPermission := false
		for _, p := range user.Permissions() {
			if p == req.RequiredPermission {
				hasPermission = true
				break
			}
		}
		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "insufficient permissions",
				"valid":   true,
				"allowed": false,
			})
			return
		}
	}

	common.OK(c, gin.H{
		"valid":       true,
		"allowed":     true,
		"user_id":     user.ID,
		"email":       user.Email,
		"name":        user.Name,
		"is_admin":    user.IsAdmin,
		"permissions": user.Permissions(),
	})
}

func issueToken(user *models.User) (string, error) {
	return common.GenerateToken(
		user.ID,
		user.Email,
		user.Name,
		user.IsAdmin,
		user.Permissions(),
		privateKeyPath(),
		24*time.Hour,
	)
}

func privateKeyPath() string {
	if p := os.Getenv("PRIVATE_KEY_PATH"); p != "" {
		return p
	}
	return "./keys/private.pem"
}

func publicKeyPath() string {
	if p := os.Getenv("PUBLIC_KEY_PATH"); p != "" {
		return p
	}
	return "./keys/public.pem"
}
