package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
						UserID:      "other-user",
						OriginalURL: "https://example.com",
						ShortURL:    "abc123",
					}, nil
				}

				// Create test request
				req, _ := http.NewRequest(http.MethodGet, "/api/links/link-123", nil)

				// Set up route with path parameter
				router.GET("/api/links/:id", func(c *gin.Context) {
					c.Set("user_id", "user-123") // Different from link owner
					handler.GetLinkForTest(c)
				})
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusForbidden))
			})
		})
	})

	// Additional handler method tests
	Describe("UpdateLink", func() {
		Context("when the request is valid", func() {
			It("updates a link and returns 200 OK", func() {
				// Setup mock service responses
				linkSvc.GetLinkFunc = func(id string) (*domain.Link, error) {
					return &domain.Link{
						ID:          "link-123",
						UserID:      "user-123",
						OriginalURL: "https://example.com",
						ShortURL:    "abc123",
					}, nil
				}

				linkSvc.UpdateLinkFunc = func(id string, req service.UpdateLinkRequest) (*domain.Link, error) {
					return &domain.Link{
						ID:          "link-123",
						UserID:      "user-123",
						OriginalURL: "https://updated-example.com",
						ShortURL:    "abc123",
					}, nil
				}

				// Create test request
				reqBody := map[string]interface{}{
					"original_url": "https://updated-example.com",
				}
				jsonBody, _ := json.Marshal(reqBody)
				req, _ := http.NewRequest(http.MethodPut, "/api/links/link-123", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")

				// Register and execute the handler
				router.PUT("/api/links/:id", func(c *gin.Context) {
					c.Set("user_id", "user-123")
					handler.UpdateLinkForTest(c)
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
				Expect(respBody["original_url"]).To(Equal("https://updated-example.com"))
				Expect(respBody["short_url"]).To(Equal("abc123"))
			})
		})

		Context("when the link doesn't exist", func() {
			It("returns 404 Not Found", func() {
				// Setup mock service
				linkSvc.GetLinkFunc = func(id string) (*domain.Link, error) {
					return nil, domain.ErrNotFound
				}

				// Create test request
				reqBody := map[string]interface{}{
					"original_url": "https://updated-example.com",
				}
				jsonBody, _ := json.Marshal(reqBody)
				req, _ := http.NewRequest(http.MethodPut, "/api/links/non-existent", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")

				// Register and execute the handler
				router.PUT("/api/links/:id", func(c *gin.Context) {
					c.Set("user_id", "user-123")
					handler.UpdateLinkForTest(c)
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
				reqBody := map[string]interface{}{
					"original_url": "https://updated-example.com",
				}
				jsonBody, _ := json.Marshal(reqBody)
				req, _ := http.NewRequest(http.MethodPut, "/api/links/link-123", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")

				// Register and execute the handler
				router.PUT("/api/links/:id", func(c *gin.Context) {
					c.Set("user_id", "user-123") // Authenticated user
					handler.UpdateLinkForTest(c)
				})
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusForbidden))
			})
		})

		Context("when the update request is invalid", func() {
			It("returns 400 Bad Request for invalid JSON", func() {
				// Setup mock service
				linkSvc.GetLinkFunc = func(id string) (*domain.Link, error) {
					return &domain.Link{
						ID:          "link-123",
						UserID:      "user-123",
						OriginalURL: "https://example.com",
						ShortURL:    "abc123",
					}, nil
				}

				// Create test request with invalid JSON
				req, _ := http.NewRequest(http.MethodPut, "/api/links/link-123", bytes.NewBuffer([]byte("invalid json")))
				req.Header.Set("Content-Type", "application/json")

				// Register and execute the handler
				router.PUT("/api/links/:id", func(c *gin.Context) {
					c.Set("user_id", "user-123")
					handler.UpdateLinkForTest(c)
				})
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})

	Describe("DeleteLink", func() {
		Context("when the link exists and user is authorized", func() {
			It("deletes the link and returns 204 No Content", func() {
				// Setup mock service
				linkSvc.GetLinkFunc = func(id string) (*domain.Link, error) {
					return &domain.Link{
						ID:          "link-123",
						UserID:      "user-123",
						OriginalURL: "https://example.com",
						ShortURL:    "abc123",
					}, nil
				}

				linkSvc.DeleteLinkFunc = func(id string) error {
					return nil
				}

				// Create test request
				req, _ := http.NewRequest(http.MethodDelete, "/api/links/link-123", nil)

				// Register and execute the handler
				router.DELETE("/api/links/:id", func(c *gin.Context) {
					c.Set("user_id", "user-123")
					handler.DeleteLinkForTest(c)
				})
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusNoContent))
				Expect(recorder.Body.Len()).To(Equal(0))
			})
		})

		Context("when the link doesn't exist", func() {
			It("returns 404 Not Found", func() {
				// Setup mock service
				linkSvc.GetLinkFunc = func(id string) (*domain.Link, error) {
					return nil, domain.ErrNotFound
				}

				// Create test request
				req, _ := http.NewRequest(http.MethodDelete, "/api/links/non-existent", nil)

				// Register and execute the handler
				router.DELETE("/api/links/:id", func(c *gin.Context) {
					c.Set("user_id", "user-123")
					handler.DeleteLinkForTest(c)
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
				req, _ := http.NewRequest(http.MethodDelete, "/api/links/link-123", nil)

				// Register and execute the handler
				router.DELETE("/api/links/:id", func(c *gin.Context) {
					c.Set("user_id", "user-123") // Authenticated user
					handler.DeleteLinkForTest(c)
				})
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusForbidden))
			})
		})

		Context("when the service returns an error", func() {
			It("returns 500 Internal Server Error", func() {
				// Setup mock service
				linkSvc.GetLinkFunc = func(id string) (*domain.Link, error) {
					return &domain.Link{
						ID:          "link-123",
						UserID:      "user-123",
						OriginalURL: "https://example.com",
						ShortURL:    "abc123",
					}, nil
				}

				linkSvc.DeleteLinkFunc = func(id string) error {
					return errors.New("database error")
				}

				// Create test request
				req, _ := http.NewRequest(http.MethodDelete, "/api/links/link-123", nil)

				// Register and execute the handler
				router.DELETE("/api/links/:id", func(c *gin.Context) {
					c.Set("user_id", "user-123")
					handler.DeleteLinkForTest(c)
				})
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})
	})

	Describe("ListLinks", func() {
		Context("when the user is authenticated", func() {
			It("returns a list of links with pagination", func() {
				// Setup mock service
				linkSvc.ListLinksFunc = func(userID string, page, perPage int) ([]*domain.Link, int, error) {
					links := []*domain.Link{
						{
							ID:          "link-123",
							UserID:      "user-123",
							OriginalURL: "https://example.com",
							ShortURL:    "abc123",
							Visits:      10,
						},
						{
							ID:          "link-456",
							UserID:      "user-123",
							OriginalURL: "https://another-example.com",
							ShortURL:    "def456",
							Visits:      5,
						},
					}
					return links, 2, nil
				}

				// Create test request
				req, _ := http.NewRequest(http.MethodGet, "/api/links?page=1&per_page=10", nil)

				// Register and execute the handler
				router.GET("/api/links", func(c *gin.Context) {
					c.Set("user_id", "user-123")
					handler.ListLinksForTest(c)
				})
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusOK))

				// Parse response body
				var respBody map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &respBody)
				Expect(err).NotTo(HaveOccurred())

				// Check response data
				links, ok := respBody["links"].([]interface{})
				Expect(ok).To(BeTrue())
				Expect(len(links)).To(Equal(2))

				meta, ok := respBody["meta"].(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(meta["total"]).To(Equal(float64(2)))
				Expect(meta["page"]).To(Equal(float64(1)))
				Expect(meta["per_page"]).To(Equal(float64(10)))
			})
		})

		Context("when the service returns an error", func() {
			It("returns 500 Internal Server Error", func() {
				// Setup mock service
				linkSvc.ListLinksFunc = func(userID string, page, perPage int) ([]*domain.Link, int, error) {
					return nil, 0, errors.New("database error")
				}

				// Create test request
				req, _ := http.NewRequest(http.MethodGet, "/api/links", nil)

				// Register and execute the handler
				router.GET("/api/links", func(c *gin.Context) {
					c.Set("user_id", "user-123")
					handler.ListLinksForTest(c)
				})
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when the user is not authenticated", func() {
			It("returns 401 Unauthorized", func() {
				// Create test request
				req, _ := http.NewRequest(http.MethodGet, "/api/links", nil)

				// Register and execute the handler
				router.GET("/api/links", func(c *gin.Context) {
					// No user_id set in context
					handler.ListLinksForTest(c)
				})
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("GetLinkStats", func() {
		Context("when the link exists and user is authorized", func() {
			It("returns the link statistics", func() {
				// Setup mock service
				linkSvc.GetLinkFunc = func(id string) (*domain.Link, error) {
					return &domain.Link{
						ID:          "link-123",
						UserID:      "user-123",
						OriginalURL: "https://example.com",
						ShortURL:    "abc123",
					}, nil
				}

				linkSvc.GetClicksFunc = func(linkID string, page, perPage int) ([]*domain.Click, int, error) {
					createdAt, _ := time.Parse(time.RFC3339, "2023-01-01T12:00:00Z")
					clicks := []*domain.Click{
						{
							ID:        "click-1",
							LinkID:    "link-123",
							UserAgent: "Mozilla/5.0",
							Referer:   "https://google.com",
							IPAddress: "192.168.1.1",
							CreatedAt: createdAt,
						},
					}
					return clicks, 1, nil
				}

				// Create test request
				req, _ := http.NewRequest(http.MethodGet, "/api/links/link-123/stats", nil)

				// Register and execute the handler
				router.GET("/api/links/:id/stats", func(c *gin.Context) {
					c.Set("user_id", "user-123")
					handler.GetLinkStatsForTest(c)
				})
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusOK))

				// Parse response body
				var respBody map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &respBody)
				Expect(err).NotTo(HaveOccurred())

				// Check response data
				Expect(respBody["link_id"]).To(Equal("link-123"))

				clicks, ok := respBody["clicks"].([]interface{})
				Expect(ok).To(BeTrue())
				Expect(len(clicks)).To(Equal(1))

				meta, ok := respBody["meta"].(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(meta["total"]).To(Equal(float64(1)))
			})
		})

		Context("when the link doesn't exist", func() {
			It("returns 404 Not Found", func() {
				// Setup mock service
				linkSvc.GetLinkFunc = func(id string) (*domain.Link, error) {
					return nil, domain.ErrNotFound
				}

				// Create test request
				req, _ := http.NewRequest(http.MethodGet, "/api/links/non-existent/stats", nil)

				// Register and execute the handler
				router.GET("/api/links/:id/stats", func(c *gin.Context) {
					c.Set("user_id", "user-123")
					handler.GetLinkStatsForTest(c)
				})
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusNotFound))
			})
		})

		Context("when the user is not authorized", func() {
			It("returns 403 Forbidden for another user's link", func() {
				// Setup mock service
				linkSvc.GetLinkFunc = func(id string) (*domain.Link, error) {
					return &domain.Link{
						ID:          "link-123",
						UserID:      "other-user", // Different from the authenticated user
						OriginalURL: "https://example.com",
						ShortURL:    "abc123",
					}, nil
				}

				// Create test request
				req, _ := http.NewRequest(http.MethodGet, "/api/links/link-123/stats", nil)

				// Register and execute the handler
				router.GET("/api/links/:id/stats", func(c *gin.Context) {
					c.Set("user_id", "user-123") // Authenticated user
					handler.GetLinkStatsForTest(c)
				})
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusForbidden))
			})
		})
	})

	Describe("RedirectLink", func() {
		Context("when the link exists", func() {
			It("redirects to the original URL", func() {
				// Setup mock service
				linkSvc.GetLinkByShortURLFunc = func(shortURL string) (*domain.Link, error) {
					return &domain.Link{
						ID:          "link-123",
						UserID:      "user-123",
						OriginalURL: "https://example.com",
						ShortURL:    "abc123",
					}, nil
				}

				// Create test request
				req, _ := http.NewRequest(http.MethodGet, "/abc123", nil)

				// Register and execute the handler
				router.GET("/:short_url", func(c *gin.Context) {
					handler.RedirectLinkForTest(c)
				})
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusMovedPermanently))
				Expect(recorder.Header().Get("Location")).To(Equal("https://example.com"))
			})

			It("records the click asynchronously", func() {
				// Setup mock service
				linkSvc.GetLinkByShortURLFunc = func(shortURL string) (*domain.Link, error) {
					return &domain.Link{
						ID:          "link-123",
						UserID:      "user-123",
						OriginalURL: "https://example.com",
						ShortURL:    "abc123",
					}, nil
				}

				clickRecorded := false
				linkSvc.RecordClickFunc = func(linkID, userAgent, referer, ipAddress string) error {
					clickRecorded = true
					return nil
				}

				// Create test request
				req, _ := http.NewRequest(http.MethodGet, "/abc123", nil)
				req.Header.Set("User-Agent", "Test-Agent")
				req.Header.Set("Referer", "https://test-referer.com")

				// Register and execute the handler
				router.GET("/:short_url", func(c *gin.Context) {
					handler.RedirectLinkForTest(c)
				})
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusMovedPermanently))

				// Check click recording
				Eventually(func() bool {
					return clickRecorded
				}).Should(BeTrue(), "Click should be recorded")
			})
		})

		Context("when the link doesn't exist", func() {
			It("returns 404 Not Found", func() {
				// Setup mock service
				linkSvc.GetLinkByShortURLFunc = func(shortURL string) (*domain.Link, error) {
					return nil, domain.ErrNotFound
				}

				// Create test request
				req, _ := http.NewRequest(http.MethodGet, "/non-existent", nil)

				// Register and execute the handler
				router.GET("/:short_url", func(c *gin.Context) {
					handler.RedirectLinkForTest(c)
				})
				router.ServeHTTP(recorder, req)

				// Check response
				Expect(recorder.Code).To(Equal(http.StatusNotFound))
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
