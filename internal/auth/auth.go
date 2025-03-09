package auth

import (
	"errors"

	"github.com/menezmethod/ref_go/internal/domain"
)

// Errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)

// Service defines the authentication service interface
type Service interface {
	ValidateToken(token string) (string, error)
	IsAdmin(userID string) bool
	Authenticate(email, password string) (*domain.User, error)
	GenerateToken(userID string) (string, error)
}
