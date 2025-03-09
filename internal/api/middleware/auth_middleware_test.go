package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"

	"github.com/menezmethod/ref_go/internal/api/middleware"
	"github.com/menezmethod/ref_go/internal/auth"
	"github.com/menezmethod/ref_go/internal/config"
	"github.com/menezmethod/ref_go/internal/domain"
)

func TestMiddleware(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Middleware Suite")
}

var _ = Describe("AuthMiddleware", func() {
	var (
		router   *gin.Engine
		recorder *httptest.ResponseRecorder
		authSvc  *MockAuthService
		cfg      *config.Config
		logger   *zap.Logger
		mw       *middleware.AuthMiddleware
	)

	BeforeEach(func() {
		// Set up gin in test mode
		gin.SetMode(gin.TestMode)
		router = gin.New()

		// Create mock services
		authSvc = &MockAuthService{}

		// Create test logger
		logger, _ = zap.NewDevelopment()

		// Create test config
		cfg = &config.Config{
			Auth: config.AuthConfig{
				JWTSecret: "test_secret",
				JWTExpiry: 0, // Not needed for tests
			},
		}

		// Create middleware
		mw = middleware.NewAuthMiddleware(cfg, logger, authSvc)

		// Set up test recorder
		recorder = httptest.NewRecorder()
	})

	Describe("RequireAuth", func() {
		Context("when a valid JWT token is provided", func() {
			It("sets the user ID in the context and allows the request", func() {
				// Set up a mock auth service that validates the token
				authSvc.ValidateTokenFunc = func(token string) (string, error) {
					return "user-123", nil // Return a user ID for valid token
				}

				// Set up a test endpoint that requires authentication
				router.GET("/protected", mw.RequireAuth(), func(c *gin.Context) {
					userID, exists := c.Get("user_id")
					if !exists {
						c.String(http.StatusInternalServerError, "No user ID in context")
						return
					}
					c.String(http.StatusOK, "UserID: %s", userID)
				})

				// Create test request with JWT token
				req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
				req.Header.Set("Authorization", "Bearer valid_token")

				// Execute the request
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(ContainSubstring("UserID: user-123"))
			})
		})

		Context("when no token is provided", func() {
			It("returns 401 Unauthorized", func() {
				// Set up a test endpoint that requires authentication
				router.GET("/protected", mw.RequireAuth(), func(c *gin.Context) {
					c.String(http.StatusOK, "This should not be reached")
				})

				// Create test request without token
				req, _ := http.NewRequest(http.MethodGet, "/protected", nil)

				// Execute the request
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when an invalid token is provided", func() {
			It("returns 401 Unauthorized", func() {
				// Set up a mock auth service that rejects the token
				authSvc.ValidateTokenFunc = func(token string) (string, error) {
					return "", auth.ErrInvalidToken
				}

				// Set up a test endpoint that requires authentication
				router.GET("/protected", mw.RequireAuth(), func(c *gin.Context) {
					c.String(http.StatusOK, "This should not be reached")
				})

				// Create test request with invalid token
				req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
				req.Header.Set("Authorization", "Bearer invalid_token")

				// Execute the request
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when the token has an invalid format", func() {
			It("returns 401 Unauthorized for non-Bearer tokens", func() {
				// Set up a test endpoint that requires authentication
				router.GET("/protected", mw.RequireAuth(), func(c *gin.Context) {
					c.String(http.StatusOK, "This should not be reached")
				})

				// Create test request with invalid token format
				req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
				req.Header.Set("Authorization", "NotBearer token")

				// Execute the request
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("AdminOnly", func() {
		Context("when a valid admin user is authenticated", func() {
			It("allows the request", func() {
				// Set up a mock auth service that validates the admin user
				authSvc.IsAdminFunc = func(userID string) bool {
					return true // This user is an admin
				}

				// Set up a test endpoint that requires admin access
				router.GET("/admin", mw.RequireAuth(), mw.AdminOnly(), func(c *gin.Context) {
					c.String(http.StatusOK, "Admin access granted")
				})

				// Create test request with JWT token
				req, _ := http.NewRequest(http.MethodGet, "/admin", nil)
				req.Header.Set("Authorization", "Bearer admin_token")

				// Mock token validation
				authSvc.ValidateTokenFunc = func(token string) (string, error) {
					return "admin-user", nil
				}

				// Execute the request
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(ContainSubstring("Admin access granted"))
			})
		})

		Context("when a non-admin user tries to access", func() {
			It("returns 403 Forbidden", func() {
				// Set up a mock auth service that validates a non-admin user
				authSvc.IsAdminFunc = func(userID string) bool {
					return false // This user is not an admin
				}

				// Set up a test endpoint that requires admin access
				router.GET("/admin", mw.RequireAuth(), mw.AdminOnly(), func(c *gin.Context) {
					c.String(http.StatusOK, "This should not be reached")
				})

				// Create test request with JWT token
				req, _ := http.NewRequest(http.MethodGet, "/admin", nil)
				req.Header.Set("Authorization", "Bearer user_token")

				// Mock token validation
				authSvc.ValidateTokenFunc = func(token string) (string, error) {
					return "regular-user", nil
				}

				// Execute the request
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusForbidden))
			})
		})

		Context("when no user is authenticated", func() {
			It("redirects to authentication first", func() {
				// Set up a test endpoint that requires admin access
				router.GET("/admin", mw.RequireAuth(), mw.AdminOnly(), func(c *gin.Context) {
					c.String(http.StatusOK, "This should not be reached")
				})

				// Create test request without token
				req, _ := http.NewRequest(http.MethodGet, "/admin", nil)

				// Execute the request
				router.ServeHTTP(recorder, req)

				// Check response - should be Unauthorized, not Forbidden
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})
})

// Mock services
type MockAuthService struct {
	ValidateTokenFunc func(token string) (string, error)
	IsAdminFunc       func(userID string) bool
	AuthenticateFunc  func(email, password string) (*domain.User, error)
	GenerateTokenFunc func(userID string) (string, error)
}

func (m *MockAuthService) ValidateToken(token string) (string, error) {
	if m.ValidateTokenFunc != nil {
		return m.ValidateTokenFunc(token)
	}
	return "", nil
}

func (m *MockAuthService) IsAdmin(userID string) bool {
	if m.IsAdminFunc != nil {
		return m.IsAdminFunc(userID)
	}
	return false
}

func (m *MockAuthService) Authenticate(email, password string) (*domain.User, error) {
	if m.AuthenticateFunc != nil {
		return m.AuthenticateFunc(email, password)
	}
	return nil, nil
}

func (m *MockAuthService) GenerateToken(userID string) (string, error) {
	if m.GenerateTokenFunc != nil {
		return m.GenerateTokenFunc(userID)
	}
	return "", nil
}
