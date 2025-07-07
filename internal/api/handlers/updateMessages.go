package handlers

import (
	"chat-ai-backend/internal/services"
	"chat-ai-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MessageUpdateHandler struct {
	Service *services.UpdateMessageService
}

func NewUpdateMessageHandler(service *services.UpdateMessageService) *MessageUpdateHandler {
	return &MessageUpdateHandler{Service: service}
}

func (h *MessageUpdateHandler) UpdateMessage(c *gin.Context) {
	// Extract userID from context
	_, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return // Response already written in util
	}

	var input struct {
		Feedback string `json:"feedback" binding:"required"`
		ThumbUp  int    `json:"thumbup" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Logger.Printf("Invalid input: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	messageID := c.Param("id")
	if messageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message ID is required"})
		return
	}

	err := h.Service.UpdateMessage(messageID, input.Feedback, input.ThumbUp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message updated successfully"})
}
