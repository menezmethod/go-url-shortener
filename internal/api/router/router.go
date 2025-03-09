package router

import (
	"net/http"
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
	linkHandler := handlers.NewLinkHandler(shortenerService, cfg.Server.BaseURL)

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
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Register metrics endpoint (protected)
	metricsGroup := router.Group("/metrics")
	metricsGroup.Use(middleware.Authentication(tokenService))
	{
		metricsGroup.GET("", func(c *gin.Context) {
			// Update short link count before serving metrics
			count, err := linkRepo.Count(c.Request.Context())
			if err != nil {
				logger.Error("Failed to get short link count", zap.Error(err))
			} else {
				metricsCollector.SetShortLinkCount(int64(count))
			}

			metricsCollector.ServeHTTP(c.Writer, c.Request)
		})
	}

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
