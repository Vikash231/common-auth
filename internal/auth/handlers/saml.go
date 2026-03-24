package handlers

import (
	"common-auth/internal/auth/db"
	"common-auth/internal/auth/models"
	"common-auth/internal/common"
	"context"
	"encoding/base64"
	"encoding/xml"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/crewjam/saml/samlsp"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var samlMiddleware *samlsp.Middleware

type samlStatusOnly struct {
	XMLName xml.Name `xml:"Response"`
	Status  struct {
		StatusCode struct {
			Value string `xml:"Value,attr"`
		} `xml:"StatusCode"`
		StatusMessage string `xml:"StatusMessage"`
		StatusDetail  string `xml:"StatusDetail"`
	} `xml:"Status"`
}

// samlMetadataTransport sets headers so metadata GET matches what curl/browsers send; avoids some 403/400 from default Go UA.
type samlMetadataTransport struct{ base http.RoundTripper }

func (t samlMetadataTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r := req.Clone(req.Context())
	if r.Header.Get("User-Agent") == "" {
		r.Header.Set("User-Agent", "curl/8.0")
	}
	if r.Header.Get("Accept") == "" {
		r.Header.Set("Accept", "application/samlmetadata+xml, application/xml, text/xml, */*")
	}
	return t.base.RoundTrip(r)
}

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

	// Use a client with browser-like headers. Some IdPs/WAFs reject Go's default User-Agent (Go-http-client/1.1).
	metaClient := &http.Client{Transport: samlMetadataTransport{base: http.DefaultTransport}}
	log.Printf("SAML fetching IdP metadata from %s", idpMetadataURLParsed.String())

	idpMetadata, err := samlsp.FetchMetadata(context.Background(), metaClient, *idpMetadataURLParsed)
	if err != nil {
		log.Printf("SAML init failed fetching IdP metadata: %v", err)
		return
	}

	rootURL, err := url.Parse(spBaseURL)
	if err != nil {
		log.Printf("SAML init failed parsing SAML_SP_BASE_URL: %v", err)
		return
	}
	// crewjam/saml joins ACS as relative "saml/acs". Per net/url.ResolveReference, a base
	// without trailing slash makes .../api/auth + saml/acs become .../api/saml/acs (drops "auth").
	// Trailing slash yields .../api/auth/saml/acs.
	if rootURL.Path != "" && rootURL.Path != "/" && !strings.HasSuffix(rootURL.Path, "/") {
		rootURL.Path += "/"
	}
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

	// Behind api-gateway, preserve original external ACS URL context for SAML validation.
	// ParseResponse validates destination/recipient against request URL.
	if xfHost := c.GetHeader("X-Forwarded-Host"); xfHost != "" {
		c.Request.Host = xfHost
		c.Request.URL.Host = xfHost
	}
	if xfProto := c.GetHeader("X-Forwarded-Proto"); xfProto != "" {
		c.Request.URL.Scheme = xfProto
	}
	if xfURI := c.GetHeader("X-Forwarded-Uri"); xfURI != "" {
		c.Request.URL.Path = xfURI
	}
	// Fallback for direct/internal callbacks where URL scheme/host are not preserved.
	// Validation must compare against the public ACS derived from SAML_SP_BASE_URL.
	if c.Request.URL.Scheme == "" || c.Request.URL.Host == "" {
		if spBase := os.Getenv("SAML_SP_BASE_URL"); spBase != "" {
			if sp, err := url.Parse(spBase); err == nil {
				if c.Request.URL.Scheme == "" {
					c.Request.URL.Scheme = sp.Scheme
				}
				if c.Request.URL.Host == "" {
					c.Request.URL.Host = sp.Host
					c.Request.Host = sp.Host
				}
				acsPath := strings.TrimSuffix(sp.Path, "/") + "/saml/acs"
				c.Request.URL.Path = acsPath
			}
		}
	}

	possibleRequestIDs := []string{}
	if samlMiddleware.ServiceProvider.AllowIDPInitiated {
		possibleRequestIDs = append(possibleRequestIDs, "")
	}
	trackedRequests := samlMiddleware.RequestTracker.GetTrackedRequests(c.Request)
	for _, tr := range trackedRequests {
		possibleRequestIDs = append(possibleRequestIDs, tr.SAMLRequestID)
	}

	assertion, err := samlMiddleware.ServiceProvider.ParseResponse(c.Request, possibleRequestIDs)
	if err != nil {
		if raw := c.PostForm("SAMLResponse"); raw != "" {
			if decoded, decErr := base64.StdEncoding.DecodeString(raw); decErr == nil {
				var st samlStatusOnly
				if xmlErr := xml.Unmarshal(decoded, &st); xmlErr == nil {
					log.Printf("SAML response status: code=%q message=%q detail=%q", st.Status.StatusCode.Value, st.Status.StatusMessage, st.Status.StatusDetail)
				}
			}
		}
		log.Printf("SAML assertion validation failed: %v (req=%s://%s%s, tracked_ids=%d)", err, c.Request.URL.Scheme, c.Request.Host, c.Request.URL.Path, len(possibleRequestIDs))
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
