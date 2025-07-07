// internal/api/router.go

package api

import (
	"chat-ai-backend/config"
	"chat-ai-backend/internal/api/handlers"
	"chat-ai-backend/internal/repositories"
	"chat-ai-backend/internal/services"
	"chat-ai-backend/middleware"
	"chat-ai-backend/pkg/database"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gin-gonic/gin"
)

func SetupRouter(mongo *mongo.Client, redis *redis.Client) *gin.Engine {
	// Init dependencies here (local to api package)
	userRepo := repositories.NewUserRepository(
		database.UserCollection,
		database.RedisUserDB,
	)

	convoRepo := repositories.NewConversationRepository(
		database.ConversationCollection,
		database.MessageCollection,
		database.RedisChatDB,
	)

	messageRepo := repositories.NewMessageRepository(
		database.MessageCollection,
		database.ConversationCollection,
	)

	messageUpdateRepo := repositories.NewMessageUpdateRepository(
		database.MessageCollection,
		database.ConversationCollection,
		database.RedisChatDB,
	)

	redisMessageRepo := repositories.NewRedisMessageRepository(
		database.MessageCollection,
		database.ConversationCollection,
		database.RedisChatDB,
	)

	// Services
	authService := services.NewAuthService(userRepo)
	convoService := services.NewConversationService(convoRepo)
	messageService := services.NewMessageService(messageRepo)
	messageUpdateService := services.NewUpdateMessageService(messageUpdateRepo)
	redisMessageService := services.NewRedisMessageService(redisMessageRepo)
	OpenAIService := services.NewOpenAIService(config.AppConfig.OpenAIUrl, config.AppConfig.OpenAIKey)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	convoHandler := handlers.NewConversationHandler(convoService)
	messageHandler := handlers.NewMessageHandler(messageService, convoService, redisMessageService, OpenAIService)
	updateMessageHandler := handlers.NewUpdateMessageHandler(messageUpdateService)

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)
	// Create a new Gin engine instance
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true, // Allow cookies (HttpOnly refresh token)
		MaxAge:           12 * time.Hour,
	}))

	// API Version 1
	v1 := r.Group("/api/v1")
	{
		// Auth routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)       // Login route
			auth.POST("/register", authHandler.Register) // Register route
			auth.GET("/refresh-token", authHandler.RefreshTokenHandler)
			auth.POST("/logout", authHandler.LogoutHandler)
		}

		// WebSocket route
		messages := v1.Group("/messages")
		messages.Use(authMiddleware.AuthMiddleware())
		{
			messages.GET("/ws", messageHandler.Messages)
			messages.PUT("/:id", updateMessageHandler.UpdateMessage)
		}

		// Conversation routes
		conversations := v1.Group("/conversations")
		conversations.Use(authMiddleware.AuthMiddleware())
		{
			conversations.POST("/", convoHandler.CreateConversationHandler)      // Create a conversation
			conversations.DELETE("/:id", convoHandler.DeleteConversationHandler) // Delete a conversation
			conversations.PATCH("/:id", convoHandler.UpdateConversationHandler)  // Update a conversation
		}
	}

	return r
}
