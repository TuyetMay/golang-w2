package main

import (
	"asset-management-api/internal/config"
	"asset-management-api/internal/database"
	"asset-management-api/internal/handler"
	"asset-management-api/internal/middleware"
	"asset-management-api/internal/repository/postgres"
	"asset-management-api/internal/service"
	"asset-management-api/internal/utils"

	"log"
	"net/http"
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

	// Initialize repositories
	folderRepo := postgres.NewFolderRepository(db)
	noteRepo := postgres.NewNoteRepository(db)
	shareRepo := postgres.NewShareRepository(db)
	userRepo := postgres.NewUserRepository(db)
	teamRepo := postgres.NewTeamRepository(db)

	// Initialize services
	folderService := service.NewFolderService(folderRepo, shareRepo)
	noteService := service.NewNoteService(noteRepo, folderRepo, shareRepo)
	shareService := service.NewShareService(shareRepo, folderRepo, noteRepo, userRepo)
	managerService := service.NewManagerService(userRepo, teamRepo, folderRepo, noteRepo, shareRepo)

	// Initialize handlers
	folderHandler := handler.NewFolderHandler(folderService)
	noteHandler := handler.NewNoteHandler(noteService)
	shareHandler := handler.NewShareHandler(shareService)
	managerHandler := handler.NewManagerHandler(managerService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtUtil)

	// Setup Gin router
	router := setupRouter(folderHandler, noteHandler, shareHandler, managerHandler, authMiddleware, jwtUtil)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	middleware.LogInfo("Server starting", map[string]interface{}{
		"port":        cfg.Server.Port,
		"environment": gin.Mode(),
		"version":     "1.0.0",
	})

	// Start server
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		middleware.LogError(err, map[string]interface{}{
			"component": "http_server",
			"action":    "start",
		})
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter(
	folderHandler *handler.FolderHandler,
	noteHandler *handler.NoteHandler,
	shareHandler *handler.ShareHandler,
	managerHandler *handler.ManagerHandler,
	authMiddleware *middleware.AuthMiddleware,
	jwtUtil *utils.JWTUtil,
) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Global middleware - Order matters!
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.StructuredLoggingMiddleware()) // Replace default Gin logger
	router.Use(middleware.RequestResponseLoggingMiddleware()) // Detailed logging
	router.Use(middleware.PrometheusMiddleware()) // Metrics collection
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
				middleware.RecordShareCreated("folder", "unknown") // You can extract access level from request
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