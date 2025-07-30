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
	"github.com/google/uuid" // Add this import
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
	log.Printf("Database connected successfully")

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

	// Setup Gin router - Pass jwtUtil to setupRouter
	router := setupRouter(folderHandler, noteHandler, shareHandler, managerHandler, authMiddleware, jwtUtil)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	log.Printf("Server starting on port %s", cfg.Server.Port)
	log.Printf("Environment: %s", gin.Mode())

	// Start server
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupRouter(
	folderHandler *handler.FolderHandler,
	noteHandler *handler.NoteHandler,
	shareHandler *handler.ShareHandler,
	managerHandler *handler.ManagerHandler,
	authMiddleware *middleware.AuthMiddleware,
	jwtUtil *utils.JWTUtil, // Add jwtUtil parameter
) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode) // Change to gin.DebugMode for development

	router := gin.New()

	// Global middleware
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.RequestLoggingMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.SecurityMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		utils.SuccessResponse(c, http.StatusOK, "Server is healthy", gin.H{
			"timestamp": time.Now().UTC(),
			"service":   "asset-management-api",
			"version":   "1.0.0",
		})
	})

	// Test login endpoint for debugging (REMOVE IN PRODUCTION)
	router.POST("/test/login", func(c *gin.Context) {
		testUserID := uuid.New()
		token, err := jwtUtil.GenerateToken(testUserID, "test@example.com", "manager", "testuser")
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token", err.Error())
			return
		}
		utils.SuccessResponse(c, http.StatusOK, "Test token generated", gin.H{
			"token": token,
			"user_id": testUserID,
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
			folders.POST("", folderHandler.CreateFolder)                    // Create folder
			folders.GET("/:folderId", folderHandler.GetFolder)              // Get folder by ID
			folders.PUT("/:folderId", folderHandler.UpdateFolder)           // Update folder
			folders.DELETE("/:folderId", folderHandler.DeleteFolder)        // Delete folder
			folders.GET("", folderHandler.GetUserFolders)                   // Get user's folders

			// Notes in folder
			folders.POST("/:folderId/notes", noteHandler.CreateNote)        // Create note in folder
			folders.GET("/:folderId/notes", noteHandler.GetNotesByFolder)   // Get notes in folder

			// Folder sharing
			folders.POST("/:folderId/share", shareHandler.ShareFolder)                    // Share folder
			folders.DELETE("/:folderId/share/:userId", shareHandler.UnshareFolder)        // Unshare folder
			folders.GET("/:folderId/shares", shareHandler.GetFolderShares)                // Get folder shares
		}

		// Note management routes
		notes := v1.Group("/notes")
		{
			notes.GET("/:noteId", noteHandler.GetNote)                      // Get note by ID
			notes.PUT("/:noteId", noteHandler.UpdateNote)                   // Update note
			notes.DELETE("/:noteId", noteHandler.DeleteNote)                // Delete note
			notes.GET("", noteHandler.GetUserNotes)                         // Get user's notes

			// Note sharing
			notes.POST("/:noteId/share", shareHandler.ShareNote)                         // Share note
			notes.DELETE("/:noteId/share/:userId", shareHandler.UnshareNote)             // Unshare note
			notes.GET("/:noteId/shares", shareHandler.GetNoteShares)                     // Get note shares
		}

		// Manager-only routes
		manager := v1.Group("/")
		manager.Use(authMiddleware.RequireManagerRole())
		{
			manager.GET("/teams/:teamId/assets", managerHandler.GetTeamAssets)    // Get team assets
			manager.GET("/users/:userId/assets", managerHandler.GetUserAssets)    // Get user assets
		}
	}

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		utils.ErrorResponse(c, http.StatusNotFound, "Endpoint not found", "The requested endpoint does not exist")
	})

	return router
}