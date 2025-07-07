// config/config.go

package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	MongoURI             string
	JWTSecretKey         string
	OpenAIUrl            string
	OpenAIKey            string
	RedisHost            string
	RedisPort            string
	RedisPassword        string
	RedisDB              int
	RedisChatDB          int
	MilvusHost           string
	MilvusPort           int
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

var AppConfig *Config

func LoadConfig() {
	// Load the .env file , if use Dockerfile then I cannot use this
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Printf("Error loading .env file: %v", err)
	// }

	// Parse Redis DB as an integer
	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		log.Printf("Invalid REDIS_DB value, must be an integer: %v", err)
	}

	// Parse Redis Chat DB as an integer
	redisChatDB, err := strconv.Atoi(getEnv("REDIS_CHAT_DB", "1"))
	if err != nil {
		log.Printf("Invalid REDIS_CHAT_DB value, must be an integer: %v", err)
	}

	milvusPort, err := strconv.Atoi(getEnv("MILVUS_PORT", "19530"))
	if err != nil {
		log.Printf("Invalid MILVUS_PORT value, must be an integer: %v", err)
	}

	// Initialize the AppConfig with values from the environment
	AppConfig = &Config{
		MongoURI:             getEnv("MONGO_URI", ""),
		JWTSecretKey:         getEnv("JWT_SECRET_KEY", "default_jwt_secret"),
		OpenAIUrl:            getEnv("OPENAI_URL", "http://localhost:8090"),
		OpenAIKey:            getEnv("OPENAI_API_KEY", ""),
		RedisHost:            getEnv("REDIS_HOST", "localhost"),
		RedisPort:            getEnv("REDIS_PORT", "6379"),
		RedisPassword:        getEnv("REDIS_PASSWORD", ""),
		RedisDB:              redisDB,
		RedisChatDB:          redisChatDB,
		MilvusHost:           getEnv("MILVUS_HOST", "127.0.0.1"),
		MilvusPort:           milvusPort,
		AccessTokenDuration:  600 * time.Second,
		RefreshTokenDuration: 7 * 24 * time.Hour,
	}
	log.Printf("Configuration loaded successfully!")
}

// Helper to get an environment variable or a default value
func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
