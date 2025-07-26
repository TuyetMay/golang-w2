package main

import (
	"log"
	"team-service/internal/config"
	"team-service/internal/database"
	"team-service/internal/handlers"
	"team-service/internal/middleware"
	"team-service/internal/repositories"
	"team-service/internal/services"
	"team-service/pkg/logger"
	
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize logger
	logger.Init()
	
	// Load configuration
	cfg := config.LoadConfig()
	
	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	// Initialize repositories
	teamRepo := repositories.NewTeamRepository(db)
	
	// Initialize services
	teamService := services.NewTeamService(teamRepo)
	
	// Initialize handlers
	teamHandler := handlers.NewTeamHandler(teamService)
	
	// Setup routes
	router := setupRoutes(teamHandler)
	
	// Start server
	logger.Log.Infof("Starting team service on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func setupRoutes(teamHandler *handlers.TeamHandler) *gin.Engine {
	router := gin.Default()
	
	// Add middleware
	router.Use(middleware.CORSMiddleware())
	router.Use(gin.Recovery())
	
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"service": "team-service",
		})
	})
	
	// Protected routes
	api := router.Group("/api/v1")
	api.Use(middleware.AuthMiddleware())
	{
		// Team management
		teams := api.Group("/teams")
		{
			teams.POST("", teamHandler.CreateTeam)
			teams.GET("", teamHandler.GetUserTeams)
			teams.GET("/:teamId", teamHandler.GetTeam)
			
			// Member management
			teams.POST("/:teamId/members", teamHandler.AddMember)
			teams.DELETE("/:teamId/members/:memberId", teamHandler.RemoveMember)
			
			// Manager management
			teams.POST("/:teamId/managers", teamHandler.AddManager)
			teams.DELETE("/:teamId/managers/:managerId", teamHandler.RemoveManager)
		}
	}
	
	return router
}