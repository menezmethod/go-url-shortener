package middleware_test

import (
	"net/http"
	"net/http/httptest"

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
				router.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(204))
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
