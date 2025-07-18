package main

import (
	"chat-ai-backend/config"
	"chat-ai-backend/internal/api"
	"chat-ai-backend/internal/repositories"
	"chat-ai-backend/internal/services"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chat-ai-backend/pkg/database"
	"chat-ai-backend/utils"
)

func main() {
	utils.Logger.Info("Starting server initialization...")

	// Load configuration
	config.LoadConfig()

	// Initialize Milvus
	// database.InitMilvus()
	// defer database.CloseMilvus()

	// Initialize MongoDB
	database.InitMongo(config.AppConfig.MongoURI)

	// Initialize Redis
	database.InitRedis()

	mongoClient := database.GetMongoClient()
	redisClient := database.GetRedisClient()

	// Initialize repo and shutdown service
	msgRepo := repositories.NewRedisMessageRepository(
		database.MessageCollection,
		database.ConversationCollection,
		database.RedisChatDB,
	)
	shutdownService := services.NewShutdownService(msgRepo)

	// Setup router
	r := api.SetupRouter(mongoClient, redisClient)
	// Run server in a goroutine
	go func() {
		if err := r.Run(":8000"); err != nil {
			utils.Logger.Error("Failed to run server: %v", err)
		}
	}()

	// Listen for shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit // wait for shutdown signal
	utils.Logger.Info("Shutdown signal received. Starting graceful shutdown...")

	// Perform graceful shutdown tasks
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Flush Redis data to MongoDB
	shutdownService.GracefulShutdown()

	// Close Mongo & Redis if needed
	database.CloseMongo()
	database.CloseRedis()

	utils.Logger.Info("Server shutdown complete.")
}
