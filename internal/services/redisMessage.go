package services

import (
	"chat-ai-backend/internal/models"
	"chat-ai-backend/internal/repositories"
	"chat-ai-backend/utils"
	"time"
)

type RedisMessageService struct {
	Repo *repositories.RedisMessageRepository
}

// Constructor
func NewRedisMessageService(repo *repositories.RedisMessageRepository) *RedisMessageService {
	return &RedisMessageService{Repo: repo}
}

// LoadMessagesIntoRedis ensures messages for a conversation are loaded into Redis from MongoDB
func (s *RedisMessageService) LoadMsgIntoRedis(conversationID string) ([]models.Message, error) {
	messages, err := s.Repo.LoadMessagesIntoRedis(conversationID)
	if err != nil {
		utils.Logger.Printf("Error loading messages into Redis: %v", err)
		return nil, err
	}
	return messages, nil
}

// StoreOneMessageInRedis saves a single message to Redis
func (s *RedisMessageService) StoreOneMsgInRedis(userID, conversationID, question, answer string) error {
	return s.Repo.StoreOneMessageInRedis(userID, conversationID, question, answer)
}

// MoveConversationToMongo migrates messages from Redis to MongoDB after 30 minutes
func (s *RedisMessageService) MoveConversationToMongo(conversationID string) error {
	err := s.Repo.MoveConvToMongo(conversationID)
	if err != nil {
		utils.Logger.Printf("Error moving conversation %s to MongoDB: %v", conversationID, err)
		return err
	}
	return nil
}

// ScheduleRedisToMongoMigration schedules a migration of conversation data from Redis to MongoDB
func (s *RedisMessageService) ScheduleRedisToMongoMigration(conversationID string) {
	go func() {
		time.Sleep(30 * time.Minute) // Wait for 30 minutes
		err := s.MoveConversationToMongo(conversationID)
		if err != nil {
			utils.Logger.Printf("Failed to move conversation %s: %v", conversationID, err)
		}
		utils.Logger.Printf("Scheduled migration of conversation %s to MongoDB", conversationID)
	}()
}
