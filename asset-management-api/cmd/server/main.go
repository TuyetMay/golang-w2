package main

import (
	"asset-management-api/internal/cache"
	redisCache "asset-management-api/internal/cache/redis"
	"asset-management-api/internal/config"
	"asset-management-api/internal/database"
	"asset-management-api/internal/events/kafka"
	"asset-management-api/internal/handler"
	"asset-management-api/internal/middleware"
	"asset-management-api/internal/repository/postgres"
	"asset-management-api/internal/service"
	"asset-management-api/internal/utils"
	"asset-management-api/pkg/eventbus"
	cacheInterface "asset-management-api/pkg/cache"

	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	
	middleware.LogInfo("Database connected successfully", map[string]interface{}{
		"host": cfg.Database.Host,
		"port": cfg.Database.Port,
		"name": cfg.Database.DBName,
	})

	// Initialize JWT utility
	jwtUtil := utils.NewJWTUtil(cfg.JWT.SecretKey, cfg.JWT.ExpirationTime)

	// NEW: Initialize Redis cache if enabled
	var cacheService cacheInterface.CacheService
	if cfg.Redis.Enabled {
		cacheService, err = initializeRedisCache(&cfg.Redis)
		if err != nil {
			log.Printf("Failed to initialize Redis cache: %v, continuing without cache", err)
			cacheService = &noOpCacheService{} // Fallback to no-op implementation
		} else {
			middleware.LogInfo("Redis cache initialized successfully", map[string]interface{}{
				"host": cfg.Redis.Host,
				"port": cfg.Redis.Port,
				"database": cfg.Redis.Database,
			})
		}
	} else {
		log.Println("Redis cache disabled, using no-op cache service")
		cacheService = &noOpCacheService{}
	}

	// Initialize Kafka event bus if enabled
	var eventBus eventbus.EventBus
	var cacheEventHandler *cache.CacheEventHandler
	if cfg.Kafka.Enabled {
		eventBus, err = initializeKafka(cfg)
		if err != nil {
			log.Printf("Failed to initialize Kafka: %v, continuing without event bus", err)
			eventBus = &noOpEventBus{} // Fallback to no-op implementation
		} else {
			middleware.LogInfo("Kafka initialized successfully", map[string]interface{}{
				"brokers": cfg.Kafka.Brokers,
				"group_id": cfg.Kafka.ConsumerGroupID,
			})
			
			// NEW: Initialize cache event handler and subscribe to events
			cacheEventHandler = cache.NewCacheEventHandler(cacheService)
			if err := subscribeToEvents(eventBus, cacheEventHandler); err != nil {
				log.Printf("Failed to subscribe to events: %v", err)
			}
		}
	} else {
		log.Println("Kafka disabled, using no-op event bus")
		eventBus = &noOpEventBus{}
	}

	// Initialize repositories
	folderRepo := postgres.NewFolderRepository(db)
	noteRepo := postgres.NewNoteRepository(db)
	shareRepo := postgres.NewShareRepository(db)
	userRepo := postgres.NewUserRepository(db)
	teamRepo := postgres.NewTeamRepository(db)

	// Initialize services with event bus and cache
	folderService := service.NewFolderService(folderRepo, shareRepo, eventBus)
	noteService := service.NewNoteService(noteRepo, folderRepo, shareRepo, eventBus)
	shareService := service.NewShareService(shareRepo, folderRepo, noteRepo, userRepo, eventBus)
	managerService := service.NewManagerService(userRepo, teamRepo, folderRepo, noteRepo, shareRepo)
	teamService := service.NewTeamService(teamRepo, userRepo, eventBus)

	// Initialize handlers
	folderHandler := handler.NewFolderHandler(folderService)
	noteHandler := handler.NewNoteHandler(noteService)
	shareHandler := handler.NewShareHandler(shareService)
	managerHandler := handler.NewManagerHandler(managerService)
	teamHandler := handler.NewTeamHandler(teamService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtUtil)

	// Setup Gin router
	router := setupRouter(folderHandler, noteHandler, shareHandler, managerHandler, teamHandler, authMiddleware, jwtUtil, cacheService)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server in a goroutine
	go func() {
		middleware.LogInfo("Server starting", map[string]interface{}{
			"port":        cfg.Server.Port,
			"environment": gin.Mode(),
			"version":     "1.0.0",
			"kafka_enabled": cfg.Kafka.Enabled,
			"redis_enabled": cfg.Redis.Enabled,
		})

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			middleware.LogError(err, map[string]interface{}{
				"component": "http_server",
				"action":    "start",
			})
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-ctx.Done()
	
	// Graceful shutdown
	log.Println("Shutting down server...")
	
	// Create a deadline for shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Close event bus
	if eventBus != nil {
		if err := eventBus.Close(); err != nil {
			log.Printf("Error closing event bus: %v", err)
		}
	}

	// NEW: Close cache service
	if cacheService != nil {
		if err := cacheService.Close(); err != nil {
			log.Printf("Error closing cache service: %v", err)
		}
	}

	log.Println("Server exited")
}

// NEW: Initialize Redis cache
func initializeRedisCache(cfg *config.RedisConfig) (cacheInterface.CacheService, error) {
	// Convert config to Redis config
	redisConfig := &redisCache.RedisConfig{
		Host:               cfg.Host,
		Port:               cfg.Port,
		Password:           cfg.Password,
		Database:           cfg.Database,
		PoolSize:           cfg.PoolSize,
		MinIdleConns:       cfg.MinIdleConns,
		MaxRetries:         cfg.MaxRetries,
		RetryDelay:         cfg.RetryDelay,
		PoolTimeout:        cfg.PoolTimeout,
		IdleTimeout:        cfg.IdleTimeout,
		IdleCheckFrequency: cfg.IdleCheckFrequency,
		MaxConnAge:         cfg.MaxConnAge,
		ReadTimeout:        cfg.ReadTimeout,
		WriteTimeout:       cfg.WriteTimeout,
		DialTimeout:        cfg.DialTimeout,
	}

	// Create Redis client
	redisClient, err := redisCache.NewRedisClient(redisConfig)
	if err != nil {
		return nil, err
	}

	// Create cache service
	return redisCache.NewRedisCacheService(redisClient), nil
}

// NEW: Subscribe to Kafka events for cache invalidation
func subscribeToEvents(eventBus eventbus.EventBus, handler *cache.CacheEventHandler) error {
	ctx := context.Background()
	
	// Subscribe to team events
	if err := eventBus.Subscribe(ctx, "team.activity", handler.HandleTeamEvent); err != nil {
		return fmt.Errorf("failed to subscribe to team events: %w", err)
	}
	
	// Subscribe to asset events
	if err := eventBus.Subscribe(ctx, "asset.changes", handler.HandleAssetEvent); err != nil {
		return fmt.Errorf("failed to subscribe to asset events: %w", err)
	}
	
	log.Println("Successfully subscribed to Kafka events for cache invalidation")
	return nil
}

// Initialize Kafka event bus
func initializeKafka(cfg *config.Config) (eventbus.EventBus, error) {
	// Create Kafka configuration
	kafkaConfig := &kafka.KafkaConfig{
		Brokers: cfg.Kafka.Brokers,
		ProducerConfig: kafka.ProducerConfig{
			RetryMax:         cfg.Kafka.ProducerRetryMax,
			RequiredAcks:     cfg.Kafka.ProducerRequiredAcks,
			FlushTimeout:     cfg.Kafka.ProducerFlushTimeout,
			FlushFrequency:   100 * time.Millisecond,
			FlushMessages:    100,
			CompressionType:  "snappy",
			IdempotentWrites: true,
		},
		ConsumerConfig: kafka.ConsumerConfig{
			GroupID:            cfg.Kafka.ConsumerGroupID,
			SessionTimeout:     cfg.Kafka.ConsumerSessionTimeout,
			HeartbeatInterval:  3 * time.Second,
			RebalanceTimeout:   60 * time.Second,
			AutoCommit:         true,
			AutoCommitInterval: cfg.Kafka.AutoCommitInterval,
		},
	}

	// Create producer
	producer := kafka.NewKafkaProducer(kafkaConfig)
	
	// Test connectivity by creating a dummy writer
	testCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Try to publish a test message to validate connectivity
	if err := producer.Publish(testCtx, "test.connectivity", map[string]string{
		"test": "connectivity",
		"timestamp": time.Now().Format(time.RFC3339),
	}); err != nil {
		return nil, err
	}

	log.Println("Kafka connectivity test successful")
	return producer, nil
}

// NEW: No-op cache service for fallback
type noOpCacheService struct{}

func (n *noOpCacheService) CacheTeamMembers(ctx context.Context, teamID uuid.UUID, members []uuid.UUID) error { return nil }
func (n *noOpCacheService) GetTeamMembers(ctx context.Context, teamID uuid.UUID) ([]uuid.UUID, error) { return nil, nil }
func (n *noOpCacheService) AddTeamMember(ctx context.Context, teamID, memberID uuid.UUID) error { return nil }
func (n *noOpCacheService) RemoveTeamMember(ctx context.Context, teamID, memberID uuid.UUID) error { return nil }
func (n *noOpCacheService) InvalidateTeamMembers(ctx context.Context, teamID uuid.UUID) error { return nil }
func (n *noOpCacheService) CacheFolderMetadata(ctx context.Context, folder *models.Folder) error { return nil }
func (n *noOpCacheService) GetFolderMetadata(ctx context.Context, folderID uuid.UUID) (*models.Folder, error) { return nil, nil }
func (n *noOpCacheService) CacheNoteMetadata(ctx context.Context, note *models.Note) error { return nil }
func (n *noOpCacheService) GetNoteMetadata(ctx context.Context, noteID uuid.UUID) (*models.Note, error) { return nil, nil }
func (n *noOpCacheService) InvalidateFolderMetadata(ctx context.Context, folderID uuid.UUID) error { return nil }
func (n *noOpCacheService) InvalidateNoteMetadata(ctx context.Context, noteID uuid.UUID) error { return nil }
func (n *noOpCacheService) CacheAssetACL(ctx context.Context, assetID uuid.UUID, acl map[string]string) error { return nil }
func (n *noOpCacheService) GetAssetACL(ctx context.Context, assetID uuid.UUID) (map[string]string, error) { return nil, nil }
func (n *noOpCacheService) UpdateAssetACL(ctx context.Context, assetID, userID uuid.UUID, accessLevel string) error { return nil }
func (n *noOpCacheService) RemoveAssetACL(ctx context.Context, assetID, userID uuid.UUID) error { return nil }
func (n *noOpCacheService) InvalidateAssetACL(ctx context.Context, assetID uuid.UUID) error { return nil }
func (n *noOpCacheService) HealthCheck() map[string]interface{} { return map[string]interface{}{"status": "disabled"} }
func (n *noOpCacheService) Close() error { return nil }

// No-op event bus for fallback
type noOpEventBus struct{}

func (n *noOpEventBus) Publish(ctx context.Context, topic string, event interface{}) error {
	log.Printf("No-op event bus: would publish to topic %s: %+v", topic, event)
	return nil
}

func (n *noOpEventBus) Subscribe(ctx context.Context, topic string, handler eventbus.EventHandler) error {
	log.Printf("No-op event bus: would subscribe to topic %s", topic)
	return nil
}

func (n *noOpEventBus) Close() error {
	return nil
}

func setupRouter(
	folderHandler *handler.FolderHandler,
	noteHandler *handler.NoteHandler,
	shareHandler *handler.ShareHandler,
	managerHandler *handler.ManagerHandler,
	teamHandler *handler.TeamHandler,
	authMiddleware *middleware.AuthMiddleware,
	jwtUtil *utils.JWTUtil,
	cacheService cacheInterface.CacheService, // NEW: Added cache service
) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Global middleware - Order matters!
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.StructuredLoggingMiddleware())
	router.Use(middleware.RequestResponseLoggingMiddleware())
	router.Use(middleware.PrometheusMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.SecurityMiddleware())

	// Metrics endpoint for Prometheus
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check endpoint with enhanced monitoring
	router.GET("/health", func(c *gin.Context) {
		healthData := gin.H{
			"timestamp": time.Now().UTC(),
			"service":   "asset-management-api",
			"version":   "1.0.0",
			"status":    "healthy",
			"cache":     cacheService.HealthCheck(), // NEW: Include cache health
		}

		middleware.LogInfo("Health check performed", map[string]interface{}{
			"endpoint":  "/health",
			"client_ip": c.ClientIP(),
		})

		utils.SuccessResponse(c, http.StatusOK, "Server is healthy", healthData)
	})

	// Test login endpoint for debugging (REMOVE IN PRODUCTION)
	router.POST("/test/login", func(c *gin.Context) {
		testUserID := uuid.New()
		token, err := jwtUtil.GenerateToken(testUserID, "test@example.com", "manager", "testuser")
		if err != nil {
			middleware.LogError(err, map[string]interface{}{
				"component": "jwt",
				"action":    "generate_test_token",
				"user_id":   testUserID,
			})
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token", err.Error())
			return
		}

		// Record JWT generation
		middleware.RecordJWTGenerated()
		
		middleware.LogBusinessEvent("test_login", map[string]interface{}{
			"user_id":  testUserID,
			"username": "testuser",
			"role":     "manager",
		})

		utils.SuccessResponse(c, http.StatusOK, "Test token generated", gin.H{
			"token":      token,
			"user_id":    testUserID,
			"expires_in": "24h",
		})
	})

	// API v1 routes with authentication
	v1 := router.Group("/api/v1")
	v1.Use(authMiddleware.RequireAuth())
	{
		// Folder management routes
		folders := v1.Group("/folders")
		{
			folders.POST("", enhanceHandler(folderHandler.CreateFolder, "create_folder"))
			folders.GET("/:folderId", enhanceHandler(folderHandler.GetFolder, "get_folder"))
			folders.PUT("/:folderId", enhanceHandler(folderHandler.UpdateFolder, "update_folder"))
			folders.DELETE("/:folderId", enhanceHandler(folderHandler.DeleteFolder, "delete_folder"))
			folders.GET("", enhanceHandler(folderHandler.GetUserFolders, "get_user_folders"))

			// Notes in folder
			folders.POST("/:folderId/notes", enhanceHandler(noteHandler.CreateNote, "create_note"))
			folders.GET("/:folderId/notes", enhanceHandler(noteHandler.GetNotesByFolder, "get_folder_notes"))

			// Folder sharing
			folders.POST("/:folderId/share", enhanceHandler(shareHandler.ShareFolder, "share_folder"))
			folders.DELETE("/:folderId/share/:userId", enhanceHandler(shareHandler.UnshareFolder, "unshare_folder"))
			folders.GET("/:folderId/shares", enhanceHandler(shareHandler.GetFolderShares, "get_folder_shares"))
		}

		// Note management routes
		notes := v1.Group("/notes")
		{
			notes.GET("/:noteId", enhanceHandler(noteHandler.GetNote, "get_note"))
			notes.PUT("/:noteId", enhanceHandler(noteHandler.UpdateNote, "update_note"))
			notes.DELETE("/:noteId", enhanceHandler(noteHandler.DeleteNote, "delete_note"))
			notes.GET("", enhanceHandler(noteHandler.GetUserNotes, "get_user_notes"))

			// Note sharing
			notes.POST("/:noteId/share", enhanceHandler(shareHandler.ShareNote, "share_note"))
			notes.DELETE("/:noteId/share/:userId", enhanceHandler(shareHandler.UnshareNote, "unshare_note"))
			notes.GET("/:noteId/shares", enhanceHandler(shareHandler.GetNoteShares, "get_note_shares"))
		}

		// Team management routes
		teams := v1.Group("/teams")
		{
			teams.POST("", enhanceHandler(teamHandler.CreateTeam, "create_team"))
			teams.GET("/:teamId", enhanceHandler(teamHandler.GetTeam, "get_team"))
			teams.GET("", enhanceHandler(teamHandler.GetUserTeams, "get_user_teams"))

			// Team member management
			teams.POST("/:teamId/members", enhanceHandler(teamHandler.AddMember, "add_team_member"))
			teams.DELETE("/:teamId/members/:memberId", enhanceHandler(teamHandler.RemoveMember, "remove_team_member"))

			// Team manager management
			teams.POST("/:teamId/managers", enhanceHandler(teamHandler.AddManager, "add_team_manager"))
			teams.DELETE("/:teamId/managers/:managerId", enhanceHandler(teamHandler.RemoveManager, "remove_team_manager"))
		}

		// Manager-only routes
		manager := v1.Group("/")
		manager.Use(authMiddleware.RequireManagerRole())
		{
			manager.GET("/teams/:teamId/assets", enhanceHandler(managerHandler.GetTeamAssets, "get_team_assets"))
			manager.GET("/users/:userId/assets", enhanceHandler(managerHandler.GetUserAssets, "get_user_assets"))
		}
	}

	// 404 handler with logging
	router.NoRoute(func(c *gin.Context) {
		middleware.LogInfo("404 Not Found", map[string]interface{}{
			"path":      c.Request.URL.Path,
			"method":    c.Request.Method,
			"client_ip": c.ClientIP(),
		})
		utils.ErrorResponse(c, http.StatusNotFound, "Endpoint not found", "The requested endpoint does not exist")
	})

	return router
}

// enhanceHandler wraps handlers with additional monitoring and business logic
func enhanceHandler(handler gin.HandlerFunc, operation string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Get user context
		userID, _ := middleware.GetUserIDFromContext(c)
		userRole, _ := middleware.GetUserRoleFromContext(c)
		
		// Log business operation start
		middleware.LogBusinessEvent(operation+"_started", map[string]interface{}{
			"user_id":   userID,
			"user_role": userRole,
			"operation": operation,
		})

		// Execute handler
		handler(c)

		duration := time.Since(start)
		
		// Log business operation completion
		status := "success"
		if c.Writer.Status() >= 400 {
			status = "error"
		}
		
		middleware.LogBusinessEvent(operation+"_completed", map[string]interface{}{
			"user_id":     userID,
			"user_role":   userRole,
			"operation":   operation,
			"status":      status,
			"duration_ms": duration.Milliseconds(),
			"http_status": c.Writer.Status(),
		})

		// Record business metrics
		switch operation {
		case "create_folder":
			if c.Writer.Status() < 400 {
				middleware.RecordFolderCreated(userRole)
			}
		case "create_note":
			if c.Writer.Status() < 400 {
				middleware.RecordNoteCreated(userRole)
			}
		case "share_folder":
			if c.Writer.Status() < 400 {
				middleware.RecordShareCreated("folder", "unknown")
			}
		case "share_note":
			if c.Writer.Status() < 400 {
				middleware.RecordShareCreated("note", "unknown")
			}
		}

		// Log performance if slow
		if duration > 1*time.Second {
			middleware.LogPerformance(operation, duration, map[string]interface{}{
				"user_id":     userID,
				"user_role":   userRole,
				"http_status": c.Writer.Status(),
				"endpoint":    c.FullPath(),
			})
		}
	}
}