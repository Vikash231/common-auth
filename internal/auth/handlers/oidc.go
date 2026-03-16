package handlers

import (
	"common-auth/internal/auth/db"
	"common-auth/internal/auth/models"
	"common-auth/internal/common"
	"context"
	"net/http"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

var (
	oidcProvider *oidc.Provider
	oauth2Config *oauth2.Config
)

func InitOIDC() {
	clientID := os.Getenv("OIDC_CLIENT_ID")
	clientSecret := os.Getenv("OIDC_CLIENT_SECRET")
	redirectURL := os.Getenv("OIDC_REDIRECT_URL")
	issuerURL := os.Getenv("OIDC_ISSUER_URL")

	if clientID == "" || issuerURL == "" {
		return // OIDC not configured
	}

	var err error
	oidcProvider, err = oidc.NewProvider(context.Background(), issuerURL)
	if err != nil {
		return
	}

	oauth2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     oidcProvider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
}

// GET /auth/oidc/login
func OIDCLogin(c *gin.Context) {
	if oauth2Config == nil {
		common.BadRequest(c, "OIDC not configured")
		return
	}
	state := uuid.NewString()
	c.SetCookie("oidc_state", state, 300, "/", "", false, true)
	url := oauth2Config.AuthCodeURL(state)
	c.Redirect(http.StatusFound, url)
}

// GET /auth/oidc/callback
func OIDCCallback(c *gin.Context) {
	if oauth2Config == nil {
		common.BadRequest(c, "OIDC not configured")
		return
	}

	stateCookie, err := c.Cookie("oidc_state")
	if err != nil || stateCookie != c.Query("state") {
		common.BadRequest(c, "invalid state parameter")
		return
	}

	oauthToken, err := oauth2Config.Exchange(context.Background(), c.Query("code"))
	if err != nil {
		common.InternalError(c, "failed to exchange token")
		return
	}

	verifier := oidcProvider.Verifier(&oidc.Config{ClientID: oauth2Config.ClientID})
	rawIDToken, ok := oauthToken.Extra("id_token").(string)
	if !ok {
		common.InternalError(c, "no id_token in response")
		return
	}

	idToken, err := verifier.Verify(context.Background(), rawIDToken)
	if err != nil {
		common.Unauthorized(c, "failed to verify id token")
		return
	}

	var claims struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		Sub   string `json:"sub"`
	}
	if err := idToken.Claims(&claims); err != nil {
		common.InternalError(c, "failed to parse claims")
		return
	}

	user := findOrCreateOIDCUser(claims.Sub, claims.Email, claims.Name)
	tokenStr, err := issueToken(user)
	if err != nil {
		common.InternalError(c, "failed to issue token")
		return
	}

	frontendURL := os.Getenv("APP_BASE_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	c.Redirect(http.StatusFound, frontendURL+"/auth/callback?token="+tokenStr)
}

func findOrCreateOIDCUser(oidcID, email, name string) *models.User {
	var user models.User
	if err := db.DB.Preload("Group").Where("oidc_id = ?", oidcID).First(&user).Error; err == nil {
		return &user
	}

	// Also check by email
	if err := db.DB.Preload("Group").Where("email = ?", email).First(&user).Error; err == nil {
		db.DB.Model(&user).Update("oidc_id", oidcID)
		return &user
	}

	// Create new user
	var group models.PermissionGroup
	db.DB.Where("name = ?", "default").First(&group)

	user = models.User{
		ID:            uuid.NewString(),
		Email:         email,
		Name:          name,
		OIDCID:        oidcID,
		IsActive:      true,
		EmailVerified: true,
		GroupID:       &group.ID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	db.DB.Create(&user)
	db.DB.Preload("Group").First(&user, "id = ?", user.ID)
	return &user
}
