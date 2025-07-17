package repositories

import (
	"chat-ai-backend/internal/models"
	"chat-ai-backend/utils"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ConversationRepository struct {
	MongoConvoCol *mongo.Collection
	MongoMsgCol   *mongo.Collection
	RedisClient   *redis.Client
}

func NewConversationRepository(
	mongoConvoCol *mongo.Collection,
	mongoMsgCol *mongo.Collection,
	redisClient *redis.Client,
) *ConversationRepository {
	return &ConversationRepository{
		MongoConvoCol: mongoConvoCol,
		MongoMsgCol:   mongoMsgCol,
		RedisClient:   redisClient,
	}
}

// SaveConversation saves a new conversation to MongoDB.
func (r *ConversationRepository) SaveConversation(convo models.Conversation) (string, error) {
	if r.MongoConvoCol == nil {
		return "", errors.New("conversation collection is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := r.MongoConvoCol.InsertOne(ctx, convo)
	if err != nil {
		utils.Logger.Error("Failed to save conversation: %v", err)
		return "", err
	}

	objectID, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("failed to convert inserted ID to ObjectID")
	}
	convoID := objectID.Hex()
	utils.Logger.Info("Conversation created: %s", convoID)
	return convoID, nil
}

// DeleteConversation deletes a conversation and associated messages from MongoDB and Redis.
func (r *ConversationRepository) DeleteConversation(convoID string) error {
	if r.MongoConvoCol == nil || r.MongoMsgCol == nil {
		return errors.New("MongoDB collections are not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Step 1: Delete from Redis
	redisKey := fmt.Sprintf("messages:%s", convoID)
	if exists, _ := r.RedisClient.Exists(ctx, redisKey).Result(); exists > 0 {
		if _, err := r.RedisClient.Del(ctx, redisKey).Result(); err != nil {
			utils.Logger.Error("Failed to delete Redis key: %v", err)
			return err
		}
		utils.Logger.Warn("Deleted conversation %s from Redis", convoID)
	}

	// Step 2: Delete from MongoDB
	objectID, err := primitive.ObjectIDFromHex(convoID)
	if err != nil {
		return fmt.Errorf("invalid ObjectID: %w", err)
	}

	if _, err := r.MongoConvoCol.DeleteOne(ctx, bson.M{"_id": objectID}); err != nil {
		utils.Logger.Error("Failed to delete conversation: %v", err)
		return err
	}

	if _, err := r.MongoMsgCol.DeleteMany(ctx, bson.M{"conversation_id": convoID}); err != nil {
		utils.Logger.Error("Failed to delete messages: %v", err)
		return err
	}

	utils.Logger.Warn("Conversation %s fully deleted", convoID)
	return nil
}

// UpdateConversationTitle updates the title of a conversation.
func (r *ConversationRepository) UpdateConversationTitle(convoID, title string) error {
	if r.MongoConvoCol == nil {
		return errors.New("conversation collection is not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(convoID)
	if err != nil {
		return fmt.Errorf("invalid ObjectID: %w", err)
	}

	_, err = r.MongoConvoCol.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{"title": title}},
	)
	if err != nil {
		utils.Logger.Error("Failed to update title for %s: %v", convoID, err)
		return err
	}

	utils.Logger.Info("Updated title for conversation %s to %s", convoID, title)
	return nil
}
