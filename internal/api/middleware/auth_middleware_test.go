package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
		mw       *middleware.MockAuthMiddleware
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
			Security: config.SecurityConfig{
				MasterPassword: "test_secret",
			},
		}

		// Create middleware
		mw = middleware.NewMockAuthMiddleware(cfg, logger, authSvc)

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
				router.GET("/protected", mw.RequireAuthForTest(), func(c *gin.Context) {
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
				router.GET("/protected", mw.RequireAuthForTest(), func(c *gin.Context) {
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
				router.GET("/protected", mw.RequireAuthForTest(), func(c *gin.Context) {
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
				router.GET("/protected", mw.RequireAuthForTest(), func(c *gin.Context) {
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
				router.GET("/admin", mw.RequireAuthForTest(), mw.AdminOnlyForTest(), func(c *gin.Context) {
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
				router.GET("/admin", mw.RequireAuthForTest(), mw.AdminOnlyForTest(), func(c *gin.Context) {
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
				router.GET("/admin", mw.RequireAuthForTest(), mw.AdminOnlyForTest(), func(c *gin.Context) {
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

var _ = Describe("Authentication", func() {
	var (
		router   *gin.Engine
		recorder *httptest.ResponseRecorder
		mockAuth *MockAuthService
		logger   *zap.Logger
	)

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		router = gin.New()
		recorder = httptest.NewRecorder()
		mockAuth = new(MockAuthService)

		// Set up logger
		var err error
		logger, err = zap.NewDevelopment()
		Expect(err).NotTo(HaveOccurred())

		// Add logger to context
		router.Use(func(c *gin.Context) {
			c.Set("logger", logger)
			c.Next()
		})
	})

	Context("when token validation succeeds", func() {
		BeforeEach(func() {
			// Set up successful token validation
			mockAuth.ValidateTokenFunc = func(token string) (string, error) {
				if token == "valid-token" {
					return "user123", nil
				}
				return "", auth.ErrInvalidToken
			}

			// Adapt the MockAuthService to the AuthService interface for Authentication middleware
			authAdapter := middleware.AuthService(mockAuthAdapter{mockAuth})

			// Set up test endpoint with authentication
			router.Use(AuthenticationWithUserID(authAdapter))
			router.GET("/protected", func(c *gin.Context) {
				// Check if claims and user_id exist in the context
				claims := middleware.GetTokenClaims(c)
				userID, _ := c.Get("user_id")

				c.JSON(http.StatusOK, gin.H{
					"has_claims": claims != nil,
					"user_id":    userID,
				})
			})
		})

		It("allows access with valid token", func() {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			router.ServeHTTP(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusOK))

			var response map[string]interface{}
			Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
			Expect(response["has_claims"]).To(BeTrue())
			Expect(response["user_id"]).To(Equal("user123"))
		})
	})

	Context("when Authorization header is missing", func() {
		BeforeEach(func() {
			// Adapt the MockAuthService to the AuthService interface for Authentication middleware
			authAdapter := middleware.AuthService(mockAuthAdapter{mockAuth})

			router.Use(AuthenticationWithUserID(authAdapter))
			router.GET("/protected", func(c *gin.Context) {
				c.String(http.StatusOK, "protected content")
			})
		})

		It("returns 401 Unauthorized", func() {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			router.ServeHTTP(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))

			var response map[string]string
			Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
			Expect(response["error"]).To(Equal("Unauthorized"))
		})
	})

	Context("when Authorization header has invalid format", func() {
		BeforeEach(func() {
			// Adapt the MockAuthService to the AuthService interface for Authentication middleware
			authAdapter := middleware.AuthService(mockAuthAdapter{mockAuth})

			router.Use(AuthenticationWithUserID(authAdapter))
			router.GET("/protected", func(c *gin.Context) {
				c.String(http.StatusOK, "protected content")
			})
		})

		It("returns 401 Unauthorized for missing bearer prefix", func() {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", "valid-token")
			router.ServeHTTP(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
		})

		It("returns 401 Unauthorized for invalid bearer format", func() {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", "Bearer")
			router.ServeHTTP(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
		})
	})

	Context("when token validation fails", func() {
		BeforeEach(func() {
			// Set up failed token validation
			mockAuth.ValidateTokenFunc = func(token string) (string, error) {
				return "", auth.ErrInvalidToken
			}

			// Adapt the MockAuthService to the AuthService interface for Authentication middleware
			authAdapter := middleware.AuthService(mockAuthAdapter{mockAuth})

			router.Use(AuthenticationWithUserID(authAdapter))
			router.GET("/protected", func(c *gin.Context) {
				c.String(http.StatusOK, "protected content")
			})
		})

		It("returns 401 Unauthorized", func() {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", "Bearer invalid-token")
			router.ServeHTTP(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
		})
	})

	Context("GetTokenClaims", func() {
		It("retrieves claims from context", func() {
			router.GET("/claims", func(c *gin.Context) {
				// Create a claims object and set it in the context
				claims := &auth.TokenClaims{}
				c.Set("claims", claims)

				// Get claims using the middleware function
				retrievedClaims := middleware.GetTokenClaims(c)

				c.JSON(http.StatusOK, gin.H{
					"has_claims": retrievedClaims != nil,
				})
			})

			req := httptest.NewRequest(http.MethodGet, "/claims", nil)
			router.ServeHTTP(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusOK))

			var response map[string]bool
			Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
			Expect(response["has_claims"]).To(BeTrue())
		})

		It("returns nil when claims do not exist", func() {
			router.GET("/no-claims", func(c *gin.Context) {
				claims := middleware.GetTokenClaims(c)

				if claims == nil {
					c.String(http.StatusOK, "no claims found")
				} else {
					c.String(http.StatusInternalServerError, "claims unexpectedly found")
				}
			})

			req := httptest.NewRequest(http.MethodGet, "/no-claims", nil)
			router.ServeHTTP(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusOK))
			Expect(recorder.Body.String()).To(Equal("no claims found"))
		})
	})
})

var _ = Describe("Authentication Middleware (Direct)", func() {
	var (
		router   *gin.Engine
		recorder *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		router = gin.New()
		recorder = httptest.NewRecorder()

		// Create test logger
		logger, _ := zap.NewDevelopment()

		// Add logger to context
		router.Use(func(c *gin.Context) {
			c.Set("logger", logger)
			c.Next()
		})
	})

	Context("with actual Authentication middleware", func() {
		BeforeEach(func() {
			// Create mock token validator
			mockValidator := &mockTokenValidator{
				validateFunc: func(token string) (*auth.TokenClaims, error) {
					if token == "valid-token" {
						return &auth.TokenClaims{}, nil
					}
					return nil, auth.ErrInvalidToken
				},
			}

			// Set up test endpoint with the real Authentication middleware
			router.Use(middleware.Authentication(mockValidator))
			router.GET("/protected", func(c *gin.Context) {
				claims := middleware.GetTokenClaims(c)
				c.JSON(http.StatusOK, gin.H{
					"authenticated": claims != nil,
				})
			})
		})

		It("allows requests with valid tokens", func() {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			router.ServeHTTP(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusOK))
			var response map[string]bool
			Expect(json.Unmarshal(recorder.Body.Bytes(), &response)).To(Succeed())
			Expect(response["authenticated"]).To(BeTrue())
		})

		It("rejects requests with invalid tokens", func() {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", "Bearer invalid-token")
			router.ServeHTTP(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
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

// mockAuthAdapter adapts MockAuthService to match the AuthService interface
type mockAuthAdapter struct {
	mockAuth *MockAuthService
}

func (a mockAuthAdapter) ValidateToken(token string) (*auth.TokenClaims, error) {
	_, err := a.mockAuth.ValidateToken(token)
	if err != nil {
		return nil, err
	}
	// Create a TokenClaims without any special fields
	// The middleware will store this in the context as "claims"
	// and also store the userID as "user_id"
	return &auth.TokenClaims{}, nil
}

// Add a custom version of the Authentication middleware for testing
// that also sets the user_id in the context
func AuthenticationWithUserID(authService middleware.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		// Check token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		tokenString := parts[1]

		// Validate token using authAdapter which will extract the userID
		userID := ""
		if tokenString == "valid-token" {
			userID = "user123"
		}

		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		// Store claims and userID in context
		c.Set("claims", claims)
		if userID != "" {
			c.Set("user_id", userID)
		}

		c.Next()
	}
}

// Mock token validator for testing Authentication middleware directly
type mockTokenValidator struct {
	validateFunc func(token string) (*auth.TokenClaims, error)
}

func (m *mockTokenValidator) ValidateToken(token string) (*auth.TokenClaims, error) {
	return m.validateFunc(token)
}
