// chatapp/internal/api/handlers/conversationHandler

package handlers

import (
	"chat-app/internal/services"
	"chat-app/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ConversationHandler struct {
	ConversationService *services.ConversationService
}

func NewConversationHandler(service *services.ConversationService) *ConversationHandler {
	return &ConversationHandler{ConversationService: service}
}

// Create a new conversation
func (h *ConversationHandler) CreateConversationHandler(c *gin.Context) {
	// Extract userID from context
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return // Response already written in util
	}

	// Define expected JSON input
	var input struct {
		Title string `json:"title"`
	}

	// Try to parse JSON body
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Set default title if not provided
	title := input.Title
	if title == "" {
		title = "New Conversation"
	}

	conversationID, err := h.ConversationService.CreateOrFetchConversation(userID, title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Create successfully conversationID": conversationID})
}

// DeleteConversationHandler handles the deletion of a conversation
func (h *ConversationHandler) DeleteConversationHandler(c *gin.Context) {
	// Extract userID from context
	_, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return // Response already written in util
	}

	// Get conversationID from query parameters
	conversationID := c.Param("id")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing conversationID or userID"})
		return
	}

	err := h.ConversationService.DeleteConversation(conversationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation and associated messages deleted successfully"})
}

func (h *ConversationHandler) UpdateConversationHandler(c *gin.Context) {
	// Extract userID from context
	_, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return // Response already written in util
	}

	// Get conversation ID from URL
	conversationID := c.Param("id")

	// Parse JSON body
	var input struct {
		Title string `json:"title" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call the service to update the title
	err := h.ConversationService.UpdateConversationTitle(conversationID, input.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation updated successfully"})
}
