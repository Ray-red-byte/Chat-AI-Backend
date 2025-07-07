// chatapp/internal/services/shutdown.go
package services

import (
	"chat-ai-backend/internal/repositories"
	"chat-ai-backend/utils"
)

type ShutdownService struct {
	Repo *repositories.RedisMessageRepository
}

// NewShutdownService creates a new ShutdownService
func NewShutdownService(repo *repositories.RedisMessageRepository) *ShutdownService {
	return &ShutdownService{Repo: repo}
}

// GracefulShutdown moves all Redis conversations to MongoDB before shutdown
func (s *ShutdownService) GracefulShutdown() {
	if err := s.Repo.MoveAllConversationsToMongo(); err != nil {
		utils.Logger.Printf("Error flushing Redis to MongoDB: %v", err)
	} else {
		utils.Logger.Println("All Redis conversations successfully migrated to MongoDB.")
	}
}
