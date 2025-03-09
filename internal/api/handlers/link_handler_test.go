package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"

	"github.com/menezmethod/ref_go/internal/api/handlers"
	"github.com/menezmethod/ref_go/internal/config"
	"github.com/menezmethod/ref_go/internal/domain"
	"github.com/menezmethod/ref_go/internal/service"
)

func TestHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Handlers Suite")
}

var _ = Describe("LinkHandler", func() {
	var (
		router   *gin.Engine
		linkSvc  *MockLinkService
		handler  *handlers.MockLinkHandler
		recorder *httptest.ResponseRecorder
		cfg      *config.Config
		logger   *zap.Logger
	)

	BeforeEach(func() {
		// Set up gin in test mode
		gin.SetMode(gin.TestMode)
		router = gin.New()

		// Create mock services
		linkSvc = &MockLinkService{}

		// Create test logger
		logger, _ = zap.NewDevelopment()

		// Create test config
		cfg = &config.Config{
			Server: config.ServerConfig{
				BaseURL: "http://localhost:8081",
			},
		}

		// Create handler with mock services
		handler = handlers.NewMockLinkHandler(cfg, logger, linkSvc)

		// Set up test recorder
		recorder = httptest.NewRecorder()
	})

	Describe("CreateLink", func() {
		Context("when the request is valid", func() {
			It("creates a link and returns 201 Created", func() {
				// Setup mock service response
				linkSvc.CreateLinkFunc = func(req service.CreateLinkRequest) (*domain.Link, error) {
					return &domain.Link{
						ID:          "link-123",
						UserID:      "user-123",
						OriginalURL: "https://example.com",
						ShortURL:    "abc123",
					}, nil
				}

				// Create test request
				reqBody := map[string]interface{}{
					"original_url": "https://example.com",
					"custom_alias": "abc123",
				}
				jsonBody, _ := json.Marshal(reqBody)
				req, _ := http.NewRequest(http.MethodPost, "/api/links", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")

				// Set up authentication context (in a real app, middleware would do this)
				ctx, _ := gin.CreateTestContext(recorder)
				ctx.Request = req
				ctx.Set("user_id", "user-123")

				// Register and execute the handler
				router.POST("/api/links", func(c *gin.Context) {
					// Copy authentication data from test context
					c.Set("user_id", "user-123")
					handler.CreateLinkForTest(c)
				})
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusCreated))

				// Parse response body
				var respBody map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &respBody)
				Expect(err).NotTo(HaveOccurred())

				// Check response data
				Expect(respBody["id"]).To(Equal("link-123"))
				Expect(respBody["short_url"]).To(Equal("abc123"))
				Expect(respBody["original_url"]).To(Equal("https://example.com"))
			})
		})

		Context("when the request is invalid", func() {
			It("returns 400 Bad Request for missing original URL", func() {
				// Create test request with missing original_url
				reqBody := map[string]interface{}{
					"custom_alias": "abc123",
				}
				jsonBody, _ := json.Marshal(reqBody)
				req, _ := http.NewRequest(http.MethodPost, "/api/links", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")

				// Register and execute the handler
				router.POST("/api/links", func(c *gin.Context) {
					c.Set("user_id", "user-123")
					handler.CreateLinkForTest(c)
				})
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("when the service returns an error", func() {
			It("returns 409 Conflict when the alias is already taken", func() {
				// Setup mock service to return conflict error
				linkSvc.CreateLinkFunc = func(req service.CreateLinkRequest) (*domain.Link, error) {
					return nil, domain.ErrConflict
				}

				// Create test request
				reqBody := map[string]interface{}{
					"original_url": "https://example.com",
					"custom_alias": "abc123",
				}
				jsonBody, _ := json.Marshal(reqBody)
				req, _ := http.NewRequest(http.MethodPost, "/api/links", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")

				// Register and execute the handler
				router.POST("/api/links", func(c *gin.Context) {
					c.Set("user_id", "user-123")
					handler.CreateLinkForTest(c)
				})
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusConflict))
			})

			It("returns 500 Internal Server Error for other errors", func() {
				// Setup mock service to return an internal error
				linkSvc.CreateLinkFunc = func(req service.CreateLinkRequest) (*domain.Link, error) {
					return nil, errors.New("internal error")
				}

				// Create test request
				reqBody := map[string]interface{}{
					"original_url": "https://example.com",
					"custom_alias": "abc123",
				}
				jsonBody, _ := json.Marshal(reqBody)
				req, _ := http.NewRequest(http.MethodPost, "/api/links", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")

				// Register and execute the handler
				router.POST("/api/links", func(c *gin.Context) {
					c.Set("user_id", "user-123")
					handler.CreateLinkForTest(c)
				})
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})
	})

	Describe("GetLink", func() {
		Context("when the link exists", func() {
			It("returns the link details", func() {
				// Setup mock service
				linkSvc.GetLinkFunc = func(id string) (*domain.Link, error) {
					return &domain.Link{
						ID:          "link-123",
						UserID:      "user-123",
						OriginalURL: "https://example.com",
						ShortURL:    "abc123",
						Visits:      10,
					}, nil
				}

				// Create test request
				req, _ := http.NewRequest(http.MethodGet, "/api/links/link-123", nil)

				// Set up route with path parameter
				router.GET("/api/links/:id", func(c *gin.Context) {
					c.Set("user_id", "user-123")
					handler.GetLinkForTest(c)
				})
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusOK))

				// Parse response body
				var respBody map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &respBody)
				Expect(err).NotTo(HaveOccurred())

				// Check response data
				Expect(respBody["id"]).To(Equal("link-123"))
				Expect(respBody["original_url"]).To(Equal("https://example.com"))
				Expect(respBody["short_url"]).To(Equal("abc123"))
				Expect(respBody["visits"]).To(Equal(float64(10)))
			})
		})

		Context("when the link doesn't exist", func() {
			It("returns 404 Not Found", func() {
				// Setup mock service
				linkSvc.GetLinkFunc = func(id string) (*domain.Link, error) {
					return nil, domain.ErrNotFound
				}

				// Create test request
				req, _ := http.NewRequest(http.MethodGet, "/api/links/non-existent", nil)

				// Set up route with path parameter
				router.GET("/api/links/:id", func(c *gin.Context) {
					c.Set("user_id", "user-123")
					handler.GetLinkForTest(c)
				})
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusNotFound))
			})
		})

		Context("when the user is not authorized", func() {
			It("returns 403 Forbidden for another user's link", func() {
				// Setup mock service to return a link owned by a different user
				linkSvc.GetLinkFunc = func(id string) (*domain.Link, error) {
					return &domain.Link{
						ID:          "link-123",
						UserID:      "other-user", // Different from the authenticated user
						OriginalURL: "https://example.com",
						ShortURL:    "abc123",
					}, nil
				}

				// Create test request
				req, _ := http.NewRequest(http.MethodGet, "/api/links/link-123", nil)

				// Set up route with path parameter
				router.GET("/api/links/:id", func(c *gin.Context) {
					c.Set("user_id", "user-123") // Authenticated user
					handler.GetLinkForTest(c)
				})
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusForbidden))
			})
		})
	})
})

// Mock implementations for service interfaces
type MockLinkService struct {
	CreateLinkFunc        func(req service.CreateLinkRequest) (*domain.Link, error)
	GetLinkFunc           func(id string) (*domain.Link, error)
	GetLinkByShortURLFunc func(shortURL string) (*domain.Link, error)
	UpdateLinkFunc        func(id string, req service.UpdateLinkRequest) (*domain.Link, error)
	DeleteLinkFunc        func(id string) error
	ListLinksFunc         func(userID string, page, perPage int) ([]*domain.Link, int, error)
	RecordClickFunc       func(linkID, userAgent, referer, ipAddress string) error
	GetClicksFunc         func(linkID string, page, perPage int) ([]*domain.Click, int, error)
}

func (m *MockLinkService) CreateLink(req service.CreateLinkRequest) (*domain.Link, error) {
	if m.CreateLinkFunc != nil {
		return m.CreateLinkFunc(req)
	}
	return nil, nil
}

func (m *MockLinkService) GetLink(id string) (*domain.Link, error) {
	if m.GetLinkFunc != nil {
		return m.GetLinkFunc(id)
	}
	return nil, nil
}

func (m *MockLinkService) GetLinkByShortURL(shortURL string) (*domain.Link, error) {
	if m.GetLinkByShortURLFunc != nil {
		return m.GetLinkByShortURLFunc(shortURL)
	}
	return nil, nil
}

func (m *MockLinkService) UpdateLink(id string, req service.UpdateLinkRequest) (*domain.Link, error) {
	if m.UpdateLinkFunc != nil {
		return m.UpdateLinkFunc(id, req)
	}
	return nil, nil
}

func (m *MockLinkService) DeleteLink(id string) error {
	if m.DeleteLinkFunc != nil {
		return m.DeleteLinkFunc(id)
	}
	return nil
}

func (m *MockLinkService) ListLinks(userID string, page, perPage int) ([]*domain.Link, int, error) {
	if m.ListLinksFunc != nil {
		return m.ListLinksFunc(userID, page, perPage)
	}
	return nil, 0, nil
}

func (m *MockLinkService) RecordClick(linkID, userAgent, referer, ipAddress string) error {
	if m.RecordClickFunc != nil {
		return m.RecordClickFunc(linkID, userAgent, referer, ipAddress)
	}
	return nil
}

func (m *MockLinkService) GetClicks(linkID string, page, perPage int) ([]*domain.Click, int, error) {
	if m.GetClicksFunc != nil {
		return m.GetClicksFunc(linkID, page, perPage)
	}
	return nil, 0, nil
}
