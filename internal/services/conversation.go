// chatapp/internal/services/conversationService.go

package services

import (
	"time"

	"chat-ai-backend/internal/models"
	"chat-ai-backend/internal/repositories"
	"chat-ai-backend/utils"
)

type ConversationService struct {
	Repo *repositories.ConversationRepository
}

func NewConversationService(repo *repositories.ConversationRepository) *ConversationService {
	return &ConversationService{Repo: repo}
}

// CreateOrFetchConversation handles conversation creation or retrieval
func (s *ConversationService) CreateOrFetchConversation(userID, title string) (string, error) {
	conversation := models.Conversation{
		UserID:    userID,
		Title:     title,
		CreatedAt: time.Now(),
	}
	conversationID, err := s.Repo.SaveConversation(conversation)
	if err != nil {
		utils.Logger.Error("Failed to save conversation: %v\n", err)
		return "", err
	}

	return conversationID, nil
}

// DeleteConversation deletes MongoDB conversation and its associated messages
func (s *ConversationService) DeleteConversation(conversationID string) error {
	err := s.Repo.DeleteConversation(conversationID)
	if err != nil {
		utils.Logger.Error("Failed to delete conversation and messages: %v\n", err)
		return err
	}
	utils.Logger.Info("Successfully deleted conversation and messages in mongo: %s", conversationID)
	return nil
}

// UpdateConversationTitle updates the title of a conversation
func (s *ConversationService) UpdateConversationTitle(conversationID, title string) error {
	err := s.Repo.UpdateConversationTitle(conversationID, title)
	if err != nil {
		utils.Logger.Error("Failed to update conversation title: %v\n", err)
		return err
	}
	utils.Logger.Info("Successfully updated conversation title in mongo: %s", conversationID)
	return nil
}
