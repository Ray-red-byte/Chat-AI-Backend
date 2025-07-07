package database

import (
	"chat-ai-backend/config"
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/go-redis/redis/v8"
)

var (
	ctx         = context.Background()
	RedisUserDB *redis.Client
	RedisChatDB *redis.Client
	redisOnce   sync.Once // Ensures Redis is initialized only once
)

// InitRedis initializes the Redis connection
func InitRedis() {
	redisOnce.Do(func() { // Ensures initialization happens only once
		// Construct Redis address
		redisAddr := fmt.Sprintf("%s:%s", config.AppConfig.RedisHost, config.AppConfig.RedisPort)

		// Initialize Redis client
		RedisUserDB = redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Password: config.AppConfig.RedisPassword, // Use empty password if not set
			DB:       config.AppConfig.RedisDB,
		})

		// Chat Redis DB (for storing messages)
		RedisChatDB = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", config.AppConfig.RedisHost, config.AppConfig.RedisPort),
			Password: config.AppConfig.RedisPassword,
			DB:       config.AppConfig.RedisChatDB, // Use separate DB for chat
		})

		// Test Redis connections
		if _, err := RedisUserDB.Ping(ctx).Result(); err != nil {
			log.Fatalf("Failed to connect to Redis (User DB): %v", err)
		}
		if _, err := RedisChatDB.Ping(ctx).Result(); err != nil {
			log.Fatalf("Failed to connect to Redis (Chat DB): %v", err)
		}

		log.Println("Connected to Redis successfully!")
	})
}

// CloseRedis closes the Redis connections
func CloseRedis() {
	if RedisUserDB != nil {
		if err := RedisUserDB.Close(); err != nil {
			log.Printf("Failed to close Redis User DB connection: %v", err)
		} else {
			log.Println("Redis User DB connection closed successfully!")
		}
	}

	if RedisChatDB != nil {
		if err := RedisChatDB.Close(); err != nil {
			log.Printf("Failed to close Redis Chat DB connection: %v", err)
		} else {
			log.Println("Redis Chat DB connection closed successfully!")
		}
	}
}

func GetRedisClient() *redis.Client {
	if RedisUserDB == nil {
		log.Fatal("Redis client is not initialized. Call InitRedis first.")
	}
	return RedisUserDB
}
