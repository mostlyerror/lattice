package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/benjaminpoon/brain-api/internal/db"
	"github.com/benjaminpoon/brain-api/internal/handlers"
	"github.com/benjaminpoon/brain-api/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize database
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	// Run database migrations
	migrationsPath := filepath.Join("internal", "db", "migrations")
	if err := db.RunMigrations(migrationsPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Set up Gin router
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.CORSMiddleware())

	// API routes
	api := router.Group("/api")
	{
		// Concept routes
		concepts := api.Group("/concepts")
		{
			concepts.GET("", handlers.GetConcepts)
			concepts.GET("/:id", handlers.GetConcept)
			concepts.POST("", handlers.CreateConcept)
			concepts.PATCH("/:id", handlers.UpdateConcept)
			concepts.DELETE("/:id", handlers.DeleteConcept)
		}

		// Health check endpoint
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "ok",
				"message": "Brain API is running",
			})
		})
	}

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Starting Brain API server on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
