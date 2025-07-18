// chatapp/internal/repositories/messageRepository.go

package repositories

import (
	"chat-ai-backend/internal/models"
	"chat-ai-backend/utils"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MessageRepository struct {
	MongoMsgCol   *mongo.Collection
	MongoConvoCol *mongo.Collection
}

func NewMessageRepository(
	mongoMsgCol *mongo.Collection,
	mongoConvoCol *mongo.Collection,
) *MessageRepository {
	return &MessageRepository{
		MongoMsgCol:   mongoMsgCol,
		MongoConvoCol: mongoConvoCol,
	}
}

// GetConversationByID retrieves a conversation by its ID
func (r *MessageRepository) GetConversationByID(conversationID string) (*models.Conversation, error) {
	// Convert string to ObjectId
	objectID, err := primitive.ObjectIDFromHex(conversationID)
	if err != nil {
		utils.Logger.Error("Invalid conversationID format: %s\n", conversationID)
		return nil, err
	}

	var conversation models.Conversation
	err = r.MongoConvoCol.FindOne(context.TODO(), bson.M{"_id": objectID}).Decode(&conversation)
	if err != nil {
		utils.Logger.Error("Failed to find conversation by ID %s: %v\n", conversationID, err)
		return nil, err
	}
	return &conversation, nil
}

// SaveMessage saves a message to MongoDB.
func (r *MessageRepository) SaveMessage(message models.Message) error {
	// Set default values for fields if not provided
	if message.Feedback == nil {
		defaultFeedback := "No feedback" // Default feedback message
		message.Feedback = &defaultFeedback
	}
	if message.ThumbUp == 0 {
		message.ThumbUp = 0 // Default value for ThumbUp
	}
	if message.InputURL == "" {
		message.InputURL = "" // Default empty string
	}
	if message.OutputURL == "" {
		message.OutputURL = "" // Default empty string
	}

	// Check if the MessageCollection is initialized
	if r.MongoMsgCol == nil {
		utils.Logger.Error("Error: Message collection is not initialized")
		return errors.New("message collection is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := r.MongoMsgCol.InsertOne(ctx, message)
	if err != nil {
		utils.Logger.Error("Failed to save message: %v\n", err)
		return err
	}

	utils.Logger.Info("Message saved: %+v", message)
	return nil
}
