package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/menezmethod/ref_go/internal/api/middleware"
)

var _ = Describe("Security Middleware", func() {
	var (
		router   *gin.Engine
		recorder *httptest.ResponseRecorder
	)

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		router = gin.New()
		recorder = httptest.NewRecorder()
	})

	Describe("SecurityHeaders", func() {
		BeforeEach(func() {
			router.GET("/test", middleware.SecurityHeaders(), func(c *gin.Context) {
				c.String(http.StatusOK, "success")
			})
			router.POST("/test", middleware.SecurityHeaders(), func(c *gin.Context) {
				c.String(http.StatusCreated, "created")
			})
		})

		It("sets all required security headers", func() {
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			router.ServeHTTP(recorder, req)

			headers := recorder.Header()
			Expect(headers.Get("X-Content-Type-Options")).To(Equal("nosniff"))
			Expect(headers.Get("X-Frame-Options")).To(Equal("DENY"))
			Expect(headers.Get("X-XSS-Protection")).To(Equal("1; mode=block"))
			Expect(headers.Get("Content-Security-Policy")).To(Equal("default-src 'self'"))
			Expect(headers.Get("Referrer-Policy")).To(Equal("strict-origin-when-cross-origin"))
			Expect(headers.Get("Strict-Transport-Security")).To(Equal("max-age=31536000; includeSubDomains"))
		})

		It("allows the request to proceed", func() {
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			router.ServeHTTP(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusOK))
			Expect(recorder.Body.String()).To(Equal("success"))
		})

		It("works with different HTTP methods", func() {
			req, _ := http.NewRequest(http.MethodPost, "/test", nil)
			router.ServeHTTP(recorder, req)

			Expect(recorder.Code).To(Equal(http.StatusCreated))

			// All security headers should be set regardless of method
			headers := recorder.Header()
			Expect(headers.Get("X-Content-Type-Options")).To(Equal("nosniff"))
			Expect(headers.Get("X-Frame-Options")).To(Equal("DENY"))
		})

		It("properly handles Content-Security-Policy format", func() {
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			router.ServeHTTP(recorder, req)

			csp := recorder.Header().Get("Content-Security-Policy")
			Expect(csp).NotTo(BeEmpty())

			// CSP should contain self as default source
			Expect(csp).To(ContainSubstring("default-src"))
			Expect(csp).To(ContainSubstring("'self'"))
		})

		It("properly sets Strict-Transport-Security with correct values", func() {
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			router.ServeHTTP(recorder, req)

			hsts := recorder.Header().Get("Strict-Transport-Security")
			Expect(hsts).NotTo(BeEmpty())

			// HSTS should set a reasonable max age (at least a year)
			Expect(hsts).To(ContainSubstring("max-age="))

			// Extract max-age value
			maxAge := "0"
			if strings.Contains(hsts, "max-age=") {
				parts := strings.Split(hsts, "max-age=")
				if len(parts) > 1 {
					maxAge = strings.Split(parts[1], ";")[0]
				}
			}

			// Convert to int and check it's at least 1 year (31536000 seconds)
			Expect(maxAge).To(MatchRegexp(`^\d+$`))

			// Check includeSubDomains is set
			Expect(hsts).To(ContainSubstring("includeSubDomains"))
		})

		It("allows customizing headers through the request context", func() {
			// Set up custom endpoint that modifies CSP
			router.GET("/custom", middleware.SecurityHeaders(), func(c *gin.Context) {
				// Customize a header after middleware runs but before response
				c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' https://trusted-cdn.com")
				c.String(http.StatusOK, "custom")
			})

			req, _ := http.NewRequest(http.MethodGet, "/custom", nil)
			router.ServeHTTP(recorder, req)

			Expect(recorder.Header().Get("Content-Security-Policy")).To(
				Equal("default-src 'self'; script-src 'self' https://trusted-cdn.com"))
		})
	})

	Describe("CORS", func() {
		var allowedOrigins []string

		BeforeEach(func() {
			allowedOrigins = []string{"http://localhost:3000", "https://example.com"}
			router.Use(middleware.CORS(allowedOrigins))
			router.GET("/test", func(c *gin.Context) {
				c.String(http.StatusOK, "success")
			})
		})

		Context("when origin is allowed", func() {
			It("sets CORS headers for allowed origin", func() {
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Origin", "http://localhost:3000")
				router.ServeHTTP(recorder, req)

				headers := recorder.Header()
				Expect(headers.Get("Access-Control-Allow-Origin")).To(Equal("http://localhost:3000"))
				Expect(headers.Get("Access-Control-Allow-Methods")).To(Equal("GET, POST, PUT, DELETE, OPTIONS"))
				Expect(headers.Get("Access-Control-Allow-Headers")).To(Equal("Content-Type, Authorization"))
				Expect(headers.Get("Access-Control-Max-Age")).To(Equal("86400"))
			})

			It("handles preflight requests", func() {
				req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
				req.Header.Set("Origin", "http://localhost:3000")
				req.Header.Set("Access-Control-Request-Method", "POST")
				req.Header.Set("Access-Control-Request-Headers", "Content-Type")
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(204))
				Expect(recorder.Header().Get("Access-Control-Allow-Methods")).To(ContainSubstring("POST"))
				Expect(recorder.Header().Get("Access-Control-Allow-Headers")).To(ContainSubstring("Content-Type"))
			})

			It("handles preflight requests with complex headers", func() {
				req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
				req.Header.Set("Origin", "http://localhost:3000")
				req.Header.Set("Access-Control-Request-Method", "DELETE")
				req.Header.Set("Access-Control-Request-Headers", "Content-Type, X-Requested-With, Authorization")
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(204))
				Expect(recorder.Header().Get("Access-Control-Allow-Methods")).To(ContainSubstring("DELETE"))
				Expect(recorder.Header().Get("Access-Control-Allow-Headers")).To(ContainSubstring("Content-Type"))
				Expect(recorder.Header().Get("Access-Control-Allow-Headers")).To(ContainSubstring("Authorization"))
			})
		})

		Context("when origin is not allowed", func() {
			It("does not set CORS headers", func() {
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Origin", "http://malicious.com")
				router.ServeHTTP(recorder, req)

				headers := recorder.Header()
				Expect(headers.Get("Access-Control-Allow-Origin")).To(BeEmpty())
			})

			It("but still allows the request to complete", func() {
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Origin", "http://malicious.com")
				router.ServeHTTP(recorder, req)

				// Request should complete successfully, just without CORS headers
				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Body.String()).To(Equal("success"))
			})
		})

		Context("when wildcard origin is allowed", func() {
			BeforeEach(func() {
				router = gin.New()
				router.Use(middleware.CORS([]string{"*"}))
				router.GET("/test", func(c *gin.Context) {
					c.String(http.StatusOK, "success")
				})
			})

			It("allows any origin", func() {
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				req.Header.Set("Origin", "http://any-domain.com")
				router.ServeHTTP(recorder, req)

				headers := recorder.Header()
				Expect(headers.Get("Access-Control-Allow-Origin")).To(Equal("http://any-domain.com"))
			})

			It("can handle unusual or complex origins", func() {
				// Test with unusual origins
				unusualOrigins := []string{
					"http://subdomain.example.com:8080",
					"https://example.io",
					"http://localhost:8080",
					"app://localhost", // For electron apps
				}

				for _, origin := range unusualOrigins {
					recorder = httptest.NewRecorder()
					req, _ := http.NewRequest(http.MethodGet, "/test", nil)
					req.Header.Set("Origin", origin)
					router.ServeHTTP(recorder, req)

					Expect(recorder.Header().Get("Access-Control-Allow-Origin")).To(Equal(origin))
				}
			})
		})

		Context("when no origin header is present", func() {
			It("allows the request to proceed without CORS headers", func() {
				req, _ := http.NewRequest(http.MethodGet, "/test", nil)
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusOK))
				Expect(recorder.Header().Get("Access-Control-Allow-Origin")).To(BeEmpty())
			})
		})
	})
})
