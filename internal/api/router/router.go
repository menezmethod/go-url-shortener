package router

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	"github.com/menezmethod/ref_go/internal/api/handlers"
	"github.com/menezmethod/ref_go/internal/api/middleware"
	"github.com/menezmethod/ref_go/internal/auth"
	"github.com/menezmethod/ref_go/internal/config"
	"github.com/menezmethod/ref_go/internal/db"
	"github.com/menezmethod/ref_go/internal/metrics"
	"github.com/menezmethod/ref_go/internal/repository/postgres"
	"github.com/menezmethod/ref_go/internal/service"
)

// New creates a new HTTP router with middleware
func New(cfg *config.Config, logger *zap.Logger, database *db.DB) http.Handler {
	// Set Gin to release mode in production
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create a new Gin router
	router := gin.New()

	// Initialize metrics
	metricsCollector := metrics.NewMetrics()

	// Create rate limiter
	rateLimiter := middleware.NewRateLimiter(cfg, logger)

	// Create repositories
	urlRepo := postgres.NewURLRepository(database)
	linkRepo := postgres.NewShortLinkRepository(database)
	clickRepo := postgres.NewLinkClickRepository(database)

	// Create services
	tokenService := auth.NewTokenService(cfg)
	shortenerService := service.NewURLShortenerService(
		urlRepo,
		linkRepo,
		clickRepo,
		logger,
		cfg.Server.BaseURL,
		cfg.ShortLink.DefaultExpiry,
	)

	// Create handlers
	authHandler := handlers.NewAuthHandler(tokenService)
	linkHandler := handlers.NewLinkHandler(shortenerService, cfg.Server.BaseURL, metricsCollector)

	// Apply global middleware
	router.Use(middleware.RequestID())
	router.Use(middleware.Logging(logger))
	router.Use(middleware.Recovery())
	router.Use(middleware.Metrics(metricsCollector))
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.CORS([]string{"*"})) // For development - change in production
	router.Use(middleware.Timeout(30 * time.Second))

	// Register health check and readiness endpoints (unprotected)
	router.GET("/api/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Add a specific Swagger health check endpoint
	router.GET("/api/swagger-health", func(c *gin.Context) {
		// Check if swagger.json exists
		swaggerPath := "./docs/swagger.json"
		if _, err := os.Stat(swaggerPath); os.IsNotExist(err) {
			logger.Error("Swagger JSON file not found", zap.String("path", swaggerPath), zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "error",
				"message": "Swagger documentation is not available",
				"error":   err.Error(),
			})
			return
		}

		// Return success
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"message":  "Swagger documentation is available",
			"docs_url": "/swagger/index.html",
		})
	})

	router.GET("/api/ready", func(c *gin.Context) {
		// Check database connectivity
		ctx := c.Request.Context()
		if err := database.HealthCheck(ctx); err != nil {
			logger.Error("Database health check failed", zap.Error(err))
			c.String(http.StatusServiceUnavailable, "Database connection error")
			return
		}

		// Check that migrations are up to date
		upToDate, err := db.CheckMigrations(database.DB)
		if err != nil {
			logger.Error("Migration check failed", zap.Error(err))
			c.String(http.StatusServiceUnavailable, "Database migration check failed")
			return
		}

		if !upToDate {
			logger.Error("Database migrations are not up to date")
			c.String(http.StatusServiceUnavailable, "Database migrations are not up to date")
			return
		}

		c.String(http.StatusOK, "Ready")
	})

	// Register Swagger endpoint
	router.GET("/swagger/*any", func(c *gin.Context) {
		path := c.Request.URL.Path
		logger.Info("Swagger request received",
			zap.String("path", path),
			zap.String("method", c.Request.Method),
			zap.String("client_ip", c.ClientIP()),
		)

		// Check if this is a request for the doc.json file
		if strings.HasSuffix(path, "doc.json") {
			logger.Info("Swagger doc.json request detected")

			// Try different potential paths for doc.json
			potentialPaths := []string{
				"./docs/swagger.json",
				"/app/docs/swagger.json", // Docker path
				"../docs/swagger.json",
				"../../docs/swagger.json",
				"docs/swagger.json",
			}

			var foundPath string
			for _, p := range potentialPaths {
				if _, err := os.Stat(p); err == nil {
					foundPath = p
					logger.Info("Found Swagger JSON file", zap.String("path", p))
					break
				}
			}

			// If found, serve the file directly
			if foundPath != "" {
				c.File(foundPath)
				return
			} else {
				logger.Error("Swagger JSON file not found in any of the potential locations")
			}
		}

		// Proceed with the standard handler
		ginSwagger.WrapHandler(swaggerFiles.Handler)(c)
	})

	// Register metrics endpoint (public)
	router.GET("/metrics", func(c *gin.Context) {
		// Update short link count before serving metrics
		count, err := linkRepo.Count(c.Request.Context())
		if err != nil {
			logger.Error("Failed to get short link count", zap.Error(err))
		} else {
			metricsCollector.SetShortLinkCount(int64(count))
		}

		metricsCollector.ServeHTTP(c.Writer, c.Request)
	})

	// Register auth routes
	router.POST("/api/auth/token", authHandler.GenerateToken)

	// Register redirect endpoint (unprotected)
	router.GET("/:code", linkHandler.RedirectLink)

	// Group protected API routes
	api := router.Group("/api/links")
	api.Use(middleware.Authentication(tokenService))
	api.Use(middleware.RateLimit(rateLimiter))
	{
		api.GET("", linkHandler.ListLinks)
		api.POST("", linkHandler.CreateLink)
		api.GET("/:code", linkHandler.GetLink)
		api.PUT("/:code", linkHandler.UpdateLink)
		api.DELETE("/:code", linkHandler.DeleteLink)
		api.GET("/:code/stats", linkHandler.GetLinkStats)
	}

	return router
}
