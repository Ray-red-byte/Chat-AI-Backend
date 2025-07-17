// chatapp/internal/repositories/updateMessagesRepository

package repositories

import (
	"chat-ai-backend/internal/models"
	"chat-ai-backend/utils"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MessageUpdateRepository struct {
	MongoMsgCol   *mongo.Collection
	MongoConvoCol *mongo.Collection
	RedisChatDB   *redis.Client
}

func NewMessageUpdateRepository(
	messageCollection *mongo.Collection,
	conversationCollection *mongo.Collection,
	redisClient *redis.Client,
) *MessageUpdateRepository {
	return &MessageUpdateRepository{
		MongoMsgCol:   messageCollection,
		MongoConvoCol: conversationCollection,
		RedisChatDB:   redisClient,
	}
}

// UpdateMessageByConversationID updates the feedback and thumbup fields of a message document
func (r *MessageUpdateRepository) UpdateMessageByMessageID(messageID string, feedback string, thumbUp int) error {
	// Ensure the collection is initialized
	if r.MongoMsgCol == nil {
		utils.Logger.Error("Error: Message collection is not initialized")
		return errors.New("message collection is not initialized")
	}

	// Prepare the update filter and update document
	filter := bson.M{"message_id": messageID}
	update := bson.M{
		"$set": bson.M{
			"feedback": feedback,
			"thumbup":  thumbUp,
		},
	}

	// Set a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Update in Mongo
	result, err := r.MongoMsgCol.UpdateOne(ctx, filter, update)
	if err != nil {
		utils.Logger.Error("Failed to update message: %v\n", err)
		return err
	}

	// Update in Redis
	err = r.updateMessageInRedis(messageID, feedback, thumbUp)
	if err != nil {
		utils.Logger.Error("Failed to update message in Redis: %v\n", err)
		return err
	}

	utils.Logger.Info("Matched %d document(s) and modified %d document(s)", result.MatchedCount, result.ModifiedCount)
	return nil
}

// updateMessageInRedis finds and updates a message in Redis
func (r *MessageUpdateRepository) updateMessageInRedis(messageID string, feedback string, thumbUp int) error {
	ctx := context.Background()

	// Scan all keys that store messages (messages:<conversationID>)
	pattern := "messages:*"
	keys, err := r.RedisChatDB.Keys(ctx, pattern).Result()
	if err != nil {
		utils.Logger.Error("Failed to scan Redis keys: %v", err)
		return err
	}

	for _, redisKey := range keys {
		// Get all messages in the Redis list
		messagesJSON, err := r.RedisChatDB.LRange(ctx, redisKey, 0, -1).Result()
		if err != nil {
			utils.Logger.Error("Failed to fetch messages from Redis: %v", err)
			continue
		}

		// Iterate over messages and update the target message
		var updatedMessages []string
		found := false

		for _, msgJSON := range messagesJSON {
			var msg models.Message
			if err := json.Unmarshal([]byte(msgJSON), &msg); err != nil {
				utils.Logger.Error("Error decoding Redis message: %v", err)
				continue
			}

			// If the message matches, update the fields
			if msg.MessageID == messageID {
				msg.Feedback = &feedback
				msg.ThumbUp = thumbUp
				found = true
			}

			// Convert message back to JSON and store in the new list
			updatedMsgJSON, err := json.Marshal(msg)
			if err != nil {
				utils.Logger.Error("Error encoding updated message to JSON: %v", err)
				continue
			}
			updatedMessages = append(updatedMessages, string(updatedMsgJSON))
		}

		// If message was found and updated, replace the list in Redis
		if found {
			// Remove old list
			err = r.RedisChatDB.Del(ctx, redisKey).Err()
			if err != nil {
				utils.Logger.Error("Failed to delete old message list in Redis: %v", err)
				return err
			}

			// Push updated messages back into Redis
			for _, msg := range updatedMessages {
				err = r.RedisChatDB.RPush(ctx, redisKey, msg).Err()
				if err != nil {
					utils.Logger.Error("Failed to push updated message back into Redis: %v", err)
					return err
				}
			}

			utils.Logger.Info("Updated message %s in Redis (key: %s)", messageID, redisKey)
			return nil
		}
	}

	utils.Logger.Warn("Message %s not found in Redis", messageID)
	return nil
}
