// chatapp/internal/api/handlers/messagesHandler

package handlers

import (
	"chat-ai-backend/internal/services"
	"chat-ai-backend/utils"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type MessageHandler struct {
	MessageService      *services.MessageService
	RedisMessageService *services.RedisMessageService
	ConversationService *services.ConversationService
	OpenAIService       *services.OpenAIService
}

func NewMessageHandler(messageSvc *services.MessageService, convoSvc *services.ConversationService, redisMsgSvc *services.RedisMessageService, mainLLMSvc *services.OpenAIService) *MessageHandler {
	return &MessageHandler{
		MessageService:      messageSvc,
		RedisMessageService: redisMsgSvc,
		ConversationService: convoSvc,
		OpenAIService:       mainLLMSvc,
	}
}

func (h *MessageHandler) Messages(c *gin.Context) {
	// Extract userID from context
	userID, ok := utils.GetUserIDFromContext(c)
	if !ok {
		return // Response already written in util
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		utils.Logger.Error("WebSocket upgrade failed: %v\n", err)
		return
	}
	defer conn.Close()
	utils.Logger.Info("User %s connected", userID)

	// Create or fetch the active conversation
	conversationID := c.Query("conversationID")
	if conversationID == "" {
		// If no conversationID is provided, create a new conversation

		conversationID, err = h.ConversationService.CreateOrFetchConversation(userID, "New Conversation")
		if err != nil {
			utils.Logger.Error("Failed to create conversation for user %s: %v\n", userID, err)
			c.JSON(500, gin.H{"error": "Failed to create conversation"})
			return
		}
		utils.Logger.Info("New conversation %s created for user %s", conversationID, userID)
	} else {
		// Validate if the provided conversationID belongs to the user
		valid, err := h.MessageService.ValidateConversationOwnership(userID, conversationID)
		if err != nil || !valid {
			utils.Logger.Error("Invalid or unauthorized conversationID: %s for user %s", conversationID, userID)
			c.JSON(403, gin.H{"error": "Unauthorized access to conversation"})
			return
		}
		utils.Logger.Warn("Existing conversation %s accessed by user %s", conversationID, userID)
	}

	// Load into Redis from MongoDB
	messages, err := h.RedisMessageService.LoadMsgIntoRedis(conversationID)
	if err != nil {
		utils.Logger.Error("Failed to load messages into Redis: %v\n", err)
		c.JSON(500, gin.H{"error": "Failed to load messages"})
		return
	}

	// Send previous messages to the WebSocket client as JSON
	for _, msg := range messages {
		msgBytes, err := json.Marshal(msg)
		if err != nil {
			utils.Logger.Error("Error marshalling message: %v\n", err)
			continue
		}

		err = conn.WriteMessage(websocket.TextMessage, msgBytes)
		if err != nil {
			utils.Logger.Error("Error sending message to client: %v\n", err)
			return
		}
	}

	// Schedule the deletion of messages from Redis
	h.RedisMessageService.ScheduleRedisToMongoMigration(conversationID)

	// Move data from redis to MongDB if connection close
	defer func() {
		err := h.RedisMessageService.MoveConversationToMongo(conversationID)
		if err != nil {
			utils.Logger.Error("Failed to move conversation %s to MongoDB: %v", conversationID, err)
		}
	}()

	for {
		// Read message from client
		_, message, err := conn.ReadMessage()
		if err != nil {
			utils.Logger.Error("Error reading message: %v\n", err)
			break
		}

		utils.Logger.Info("Message from user %s: %s", userID, message)

		// Process the message and stream the response
		responseChan := make(chan string)
		go h.OpenAIService.GenerateAIResponse(string(message), conversationID, responseChan)

		// Stream the response to the WebSocket
		var aiResponse string
		for response := range responseChan {
			err = conn.WriteMessage(websocket.TextMessage, []byte(response))
			if err != nil {
				utils.Logger.Error("Error writing message: %v\n", err)
				break
			}
			aiResponse += response
		}
		utils.Logger.Info("AI response for user %s: %s", userID, aiResponse)

		// Store the message in Redis
		h.RedisMessageService.StoreOneMsgInRedis(userID, conversationID, string(message), aiResponse)
	}
}
