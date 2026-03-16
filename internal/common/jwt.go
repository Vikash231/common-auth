package common

import (
	"crypto/rsa"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID      string   `json:"sub"`
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	IsAdmin     bool     `json:"is_admin"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// ValidateToken verifies a JWT token using the provided RSA public key file.
func ValidateToken(tokenString string, publicKeyPath string) (*Claims, error) {
	keyData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, errors.New("could not read public key")
	}
	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		return nil, errors.New("invalid public key")
	}
	return validateWithKey(tokenString, pubKey)
}

// ValidateTokenWithKey verifies a JWT using an already-loaded RSA public key.
func ValidateTokenWithKey(tokenString string, pubKey *rsa.PublicKey) (*Claims, error) {
	return validateWithKey(tokenString, pubKey)
}

func validateWithKey(tokenString string, pubKey *rsa.PublicKey) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return pubKey, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// GenerateToken creates a signed JWT using the RSA private key file.
func GenerateToken(userID, email, name string, isAdmin bool, permissions []string, privateKeyPath string, expiry time.Duration) (string, error) {
	keyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", errors.New("could not read private key")
	}
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return "", errors.New("invalid private key")
	}

	claims := Claims{
		UserID:      userID,
		Email:       email,
		Name:        name,
		IsAdmin:     isAdmin,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(privKey)
}
