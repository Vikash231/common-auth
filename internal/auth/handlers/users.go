package handlers

import (
	"common-auth/internal/auth/db"
	"common-auth/internal/auth/email"
	"common-auth/internal/auth/models"
	"common-auth/internal/common"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GET /auth/users  - list all users with their groups (public read, admin edit)
func ListUsers(c *gin.Context) {
	var users []models.User
	db.DB.Preload("Group").Find(&users)

	type userDTO struct {
		ID       string               `json:"id"`
		Email    string               `json:"email"`
		Name     string               `json:"name"`
		IsAdmin  bool                 `json:"is_admin"`
		IsActive bool                 `json:"is_active"`
		Group    *models.PermissionGroup `json:"group"`
	}

	result := make([]userDTO, len(users))
	for i, u := range users {
		result[i] = userDTO{
			ID:       u.ID,
			Email:    u.Email,
			Name:     u.Name,
			IsAdmin:  u.IsAdmin,
			IsActive: u.IsActive,
			Group:    u.Group,
		}
	}

	common.OK(c, result)
}

// GET /auth/groups - list all permission groups
func ListGroups(c *gin.Context) {
	var groups []models.PermissionGroup
	db.DB.Find(&groups)
	common.OK(c, groups)
}

// PUT /auth/users/:id/group  (admin only)
// Body: { "group_id": "..." }
func UpdateUserGroup(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		GroupID string `json:"group_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, err.Error())
		return
	}

	// Verify group exists
	var group models.PermissionGroup
	if err := db.DB.First(&group, "id = ?", req.GroupID).Error; err != nil {
		common.BadRequest(c, "group not found")
		return
	}

	// Cannot change admin's group
	var targetUser models.User
	if err := db.DB.First(&targetUser, "id = ?", userID).Error; err != nil {
		common.BadRequest(c, "user not found")
		return
	}
	if targetUser.IsAdmin {
		common.Forbidden(c, "cannot modify admin user's permissions")
		return
	}

	if err := db.DB.Model(&models.User{}).Where("id = ?", userID).
		Update("group_id", req.GroupID).Error; err != nil {
		common.InternalError(c, "failed to update user group")
		return
	}

	common.OK(c, gin.H{"message": "user group updated"})
}

// POST /auth/admin/invite  (admin only)
// Body: { "email": "...", "name": "..." }
func AdminInviteUser(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Name  string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, err.Error())
		return
	}

	// Check if already exists
	var existing models.User
	if db.DB.Where("email = ?", req.Email).First(&existing).Error == nil {
		common.BadRequest(c, "email already registered")
		return
	}

	// Find default group
	var group models.PermissionGroup
	if err := db.DB.Where("name = ?", "default").First(&group).Error; err != nil {
		common.InternalError(c, "default group not found")
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
	db.DB.Create(&user)

	// Get admin name from context
	adminUser, _ := c.Get("user")
	adminName := "Admin"
	if au, ok := adminUser.(*models.User); ok {
		adminName = au.Name
	}

	token := uuid.NewString()
	emailToken := models.EmailToken{
		ID:        uuid.NewString(),
		UserID:    userID,
		Token:     token,
		TokenType: "invite",
		ExpiresAt: time.Now().Add(72 * time.Hour),
		CreatedAt: time.Now(),
	}
	db.DB.Create(&emailToken)

	if err := email.SendInviteEmail(req.Email, adminName, token); err != nil {
		_ = err
	}

	common.Created(c, gin.H{"message": "invitation sent", "email": req.Email})
}

// DELETE /auth/users/:id  (admin only)
func DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := db.DB.First(&user, "id = ?", userID).Error; err != nil {
		common.BadRequest(c, "user not found")
		return
	}
	if user.IsAdmin {
		common.Forbidden(c, "cannot delete admin user")
		return
	}

	db.DB.Delete(&user)
	common.OK(c, gin.H{"message": "user deleted"})
}

// POST /auth/groups  (admin only)
// Body: { "name": "...", "can_use_weather_api": true, "can_use_distance_api": true }
func CreateGroup(c *gin.Context) {
	var req struct {
		Name              string `json:"name" binding:"required"`
		CanUseWeatherAPI  bool   `json:"can_use_weather_api"`
		CanUseDistanceAPI bool   `json:"can_use_distance_api"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, err.Error())
		return
	}

	group := models.PermissionGroup{
		ID:                uuid.NewString(),
		Name:              req.Name,
		CanUseWeatherAPI:  req.CanUseWeatherAPI,
		CanUseDistanceAPI: req.CanUseDistanceAPI,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	if err := db.DB.Create(&group).Error; err != nil {
		common.BadRequest(c, "group name already exists")
		return
	}

	common.Created(c, group)
}

// PUT /auth/groups/:id  (admin only)
func UpdateGroup(c *gin.Context) {
	groupID := c.Param("id")

	var req struct {
		CanUseWeatherAPI  *bool `json:"can_use_weather_api"`
		CanUseDistanceAPI *bool `json:"can_use_distance_api"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.BadRequest(c, err.Error())
		return
	}

	updates := map[string]interface{}{"updated_at": time.Now()}
	if req.CanUseWeatherAPI != nil {
		updates["can_use_weather_api"] = *req.CanUseWeatherAPI
	}
	if req.CanUseDistanceAPI != nil {
		updates["can_use_distance_api"] = *req.CanUseDistanceAPI
	}

	if err := db.DB.Model(&models.PermissionGroup{}).Where("id = ?", groupID).Updates(updates).Error; err != nil {
		common.InternalError(c, "failed to update group")
		return
	}

	common.OK(c, gin.H{"message": "group updated"})
}
