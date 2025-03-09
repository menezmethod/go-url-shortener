package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/menezmethod/ref_go/internal/config"
)

// TokenClaims represents the custom JWT claims
type TokenClaims struct {
	jwt.RegisteredClaims
}

// TokenService handles JWT token generation and validation
type TokenService struct {
	config *config.Config
}

// NewTokenService creates a new token service
func NewTokenService(cfg *config.Config) *TokenService {
	return &TokenService{
		config: cfg,
	}
}

// GenerateToken creates a new JWT token
func (s *TokenService) GenerateToken() (string, error) {
	now := time.Now()
	expiresAt := now.Add(s.config.Security.TokenExpiry)

	claims := TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the master password
	tokenString, err := token.SignedString([]byte(s.config.Security.MasterPassword))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken verifies that a token is valid
func (s *TokenService) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Return the secret used for signing
		return []byte(s.config.Security.MasterPassword), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// ValidateMasterPassword checks if the provided password matches the master password
func (s *TokenService) ValidateMasterPassword(password string) bool {
	return password == s.config.Security.MasterPassword
}
