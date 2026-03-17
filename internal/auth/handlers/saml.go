package handlers

import (
	"common-auth/internal/auth/db"
	"common-auth/internal/auth/models"
	"common-auth/internal/common"
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var samlMiddleware *samlsp.Middleware

// InitSAML initializes the SAML service provider middleware.
func InitSAML() {
	certFile := os.Getenv("SAML_CERT_FILE")
	keyFile := os.Getenv("SAML_KEY_FILE")
	idpMetadataURL := os.Getenv("SAML_IDP_METADATA_URL")
	spBaseURL := os.Getenv("SAML_SP_BASE_URL")

	if certFile == "" || keyFile == "" || idpMetadataURL == "" || spBaseURL == "" {
		log.Printf("SAML skipped: set SAML_CERT_FILE, SAML_KEY_FILE, SAML_IDP_METADATA_URL, SAML_SP_BASE_URL to enable")
		return
	}

	keyPair, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Printf("SAML init failed loading cert/key: %v", err)
		return
	}
	keyPair.Leaf, _ = x509.ParseCertificate(keyPair.Certificate[0])

	idpMetadataURLParsed, err := url.Parse(idpMetadataURL)
	if err != nil {
		log.Printf("SAML init failed parsing IdP metadata URL: %v", err)
		return
	}

	idpMetadata, err := samlsp.FetchMetadata(context.Background(), http.DefaultClient, *idpMetadataURLParsed)
	if err != nil {
		log.Printf("SAML init failed fetching IdP metadata: %v", err)
		return
	}

	rootURL, _ := url.Parse(spBaseURL)
	samlMiddleware, err = samlsp.New(samlsp.Options{
		URL:         *rootURL,
		Key:         keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate: keyPair.Leaf,
		IDPMetadata: idpMetadata,
	})
	if err != nil {
		log.Printf("SAML init failed creating SP: %v", err)
		return
	}
	log.Printf("SAML configured for SP base %s", spBaseURL)
}

// GET /auth/saml/login
func SAMLLogin(c *gin.Context) {
	if samlMiddleware == nil {
		common.BadRequest(c, "SAML not configured")
		return
	}
	samlMiddleware.HandleStartAuthFlow(c.Writer, c.Request)
}

// POST /auth/saml/acs  (Assertion Consumer Service)
func SAMLCallback(c *gin.Context) {
	if samlMiddleware == nil {
		common.BadRequest(c, "SAML not configured")
		return
	}

	if err := c.Request.ParseForm(); err != nil {
		common.BadRequest(c, "invalid form data")
		return
	}

	assertion, err := samlMiddleware.ServiceProvider.ParseResponse(c.Request, nil)
	if err != nil {
		common.Unauthorized(c, "SAML assertion validation failed")
		return
	}

	// Extract attributes
	samlID := assertion.Subject.NameID.Value
	email := ""
	name := ""

	for _, stmt := range assertion.AttributeStatements {
		for _, attr := range stmt.Attributes {
			switch attr.Name {
			case "email", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress":
				if len(attr.Values) > 0 {
					email = attr.Values[0].Value
				}
			case "displayName", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/displayname":
				if len(attr.Values) > 0 {
					name = attr.Values[0].Value
				}
			}
		}
	}

	if email == "" {
		common.BadRequest(c, "SAML assertion missing email attribute")
		return
	}

	user := findOrCreateSAMLUser(samlID, email, name)
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

// GET /auth/saml/metadata
func SAMLMetadata(c *gin.Context) {
	if samlMiddleware == nil {
		common.BadRequest(c, "SAML not configured")
		return
	}
	samlMiddleware.ServeMetadata(c.Writer, c.Request)
}

func findOrCreateSAMLUser(samlID, email, name string) *models.User {
	var user models.User
	if err := db.DB.Preload("Group").Where("saml_id = ?", samlID).First(&user).Error; err == nil {
		return &user
	}

	if err := db.DB.Preload("Group").Where("email = ?", email).First(&user).Error; err == nil {
		db.DB.Model(&user).Update("saml_id", samlID)
		return &user
	}

	var group models.PermissionGroup
	db.DB.Where("name = ?", "default").First(&group)

	user = models.User{
		ID:            uuid.NewString(),
		Email:         email,
		Name:          name,
		SAMLID:        samlID,
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
