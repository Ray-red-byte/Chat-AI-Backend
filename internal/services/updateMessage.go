package services

import (
	"chat-ai-backend/internal/repositories"
	"chat-ai-backend/utils"
)

type UpdateMessageService struct {
	Repo *repositories.MessageUpdateRepository
}

func NewUpdateMessageService(repo *repositories.MessageUpdateRepository) *UpdateMessageService {
	return &UpdateMessageService{Repo: repo}
}

func (s *UpdateMessageService) UpdateMessage(messageID, feedback string, thumbUp int) error {
	err := s.Repo.UpdateMessageByMessageID(messageID, feedback, thumbUp)
	if err != nil {
		utils.Logger.Printf("Service failed to update message: %v\n", err)
		return err
	}
	return nil
}
