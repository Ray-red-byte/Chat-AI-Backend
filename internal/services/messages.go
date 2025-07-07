// chatapp/internal/services/messagesService.go

package services

import (
	"time"

	"chat-ai-backend/internal/models"
	"chat-ai-backend/internal/repositories"
	"chat-ai-backend/utils"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MessageService struct {
	Repo *repositories.MessageRepository
}

// NewMessageService creates a new MessageService
func NewMessageService(repo *repositories.MessageRepository) *MessageService {
	return &MessageService{Repo: repo}
}

// ValidateConversationOwnership checks if the conversationID belongs to the userID
func (s *MessageService) ValidateConversationOwnership(userID, conversationID string) (bool, error) {
	conversation, err := s.Repo.GetConversationByID(conversationID)
	if err != nil {
		utils.Logger.Printf("Error retrieving conversation: %v\n", err)
		return false, err
	}
	return conversation.UserID == userID, nil
}

// WriteClientMessage sends a message to the WebSocket client
func (s *MessageService) WriteClientMessage(conn *websocket.Conn, messageType int, message string) error {
	err := conn.WriteMessage(messageType, []byte(message))
	if err != nil {
		utils.Logger.Printf("Error writing WebSocket message: %v\n", err)
		return err
	}
	return nil
}

// StoreMessage stores the user question and AI response in the database
func (s *MessageService) StoreMessage(userID, conversationID, question, answer string) {
	msg := models.Message{
		MessageID:      uuid.New().String(),
		UserID:         userID,
		ConversationID: conversationID,
		Question:       question,
		Answer:         answer,
		ThumbUp:        0,
		Feedback:       nil,
		InputURL:       "",
		OutputURL:      "",
		CreatedAt:      time.Now(),
	}
	err := s.Repo.SaveMessage(msg)
	if err != nil {
		utils.Logger.Printf("Failed to save message: %v\n", err)
	} else {
		utils.Logger.Printf("Message saved successfully for conversation %s", conversationID)
	}
}
