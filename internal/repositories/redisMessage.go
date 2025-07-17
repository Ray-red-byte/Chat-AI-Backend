package repositories

import (
	"chat-ai-backend/internal/models"
	"chat-ai-backend/utils"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RedisMessageRepository struct {
	MongoMsgCol   *mongo.Collection
	MongoConvoCol *mongo.Collection
	RedisChatDB   *redis.Client
}

// Constructor
func NewRedisMessageRepository(
	messageCollection *mongo.Collection, // MongoDB collection for messages
	conversationCollection *mongo.Collection, // MongoDB collection for conversations
	redisClient *redis.Client, // Redis client instance
) *RedisMessageRepository {
	return &RedisMessageRepository{
		MongoMsgCol:   messageCollection,
		MongoConvoCol: conversationCollection,
		RedisChatDB:   redisClient,
	}
}

// LoadMessagesIntoRedis loads messages from MongoDB into Redis.
func (r *RedisMessageRepository) LoadMessagesIntoRedis(conversationID string) ([]models.Message, error) {
	ctx := context.Background()
	redisKey := fmt.Sprintf("messages:%s", conversationID)

	// Check if messages exist in Redis
	exists, err := r.RedisChatDB.Exists(ctx, redisKey).Result()
	if err != nil {
		utils.Logger.Error("Redis error: %v", err)
		return nil, err
	}

	if exists > 0 {
		utils.Logger.Warn("Messages for conversationID %s are already in Redis", conversationID)
		// Retrieve existing messages from Redis
		messages, err := r.ReadAllMessagesFromRedis(conversationID) // Correct method call from RedisMessageRepository
		if err != nil {
			utils.Logger.Error("Failed to retrieve messages from Redis: %v", err)
			return nil, err
		}
		return messages, nil
	}

	// Fetch from MongoDB
	messages, err := r.LoadMessagesFromMongo(conversationID)
	if err != nil {
		utils.Logger.Error("Failed to load messages from MongoDB: %v", err)
		return nil, err
	}

	// Store messages in Redis
	err = r.StoreMessagesInRedis(conversationID, messages)
	if err != nil {
		utils.Logger.Error("Failed to store messages in Redis: %v", err)
		return nil, err
	}

	return messages, nil
}

// StoreMessagesInRedis saves multiple messages in Redis.
func (r *RedisMessageRepository) StoreMessagesInRedis(conversationID string, messages []models.Message) error {
	ctx := context.Background()
	redisKey := fmt.Sprintf("messages:%s", conversationID)

	for _, msg := range messages {
		messageJSON, err := json.Marshal(msg)
		if err != nil {
			utils.Logger.Error("Error encoding message to JSON: %v", err)
			continue
		}
		if err := r.RedisChatDB.RPush(ctx, redisKey, messageJSON).Err(); err != nil {
			utils.Logger.Error("Failed to store message in Redis: %v", err)
			return err
		}
	}

	utils.Logger.Info("Stored %d messages in Redis for conversationID %s", len(messages), conversationID)
	return nil
}

// LoadMessagesFromMongo retrieves messages from MongoDB.
func (r *RedisMessageRepository) LoadMessagesFromMongo(conversationID string) ([]models.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"conversation_id": conversationID}
	cursor, err := r.MongoMsgCol.Find(ctx, filter)
	if err != nil {
		utils.Logger.Error("Failed to query messages from MongoDB: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []models.Message
	if err := cursor.All(ctx, &messages); err != nil {
		utils.Logger.Error("Failed to decode MongoDB messages: %v", err)
		return nil, err
	}

	utils.Logger.Info("Loaded %d messages from MongoDB for conversationID %s", len(messages), conversationID)
	return messages, nil
}

// StoreOneMessageInRedis saves a single message in Redis.
func (r *RedisMessageRepository) StoreOneMessageInRedis(userID, conversationID, question, answer string) error {
	ctx := context.Background()
	redisKey := fmt.Sprintf("messages:%s", conversationID)

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

	messageJSON, err := json.Marshal(msg)
	if err != nil {
		utils.Logger.Error("Error encoding message to JSON: %v", err)
		return err
	}

	if err := r.RedisChatDB.RPush(ctx, redisKey, messageJSON).Err(); err != nil {
		utils.Logger.Error("Failed to store message in Redis: %v", err)
		return err
	}

	utils.Logger.Info("Message stored in Redis for conversation %s", conversationID)
	return nil
}

// ReadMessagesFromRedis fetches all messages from Redis.
func (r *RedisMessageRepository) ReadAllMessagesFromRedis(conversationID string) ([]models.Message, error) {
	ctx := context.Background()
	redisKey := fmt.Sprintf("messages:%s", conversationID)

	messagesJSON, err := r.RedisChatDB.LRange(ctx, redisKey, 0, -1).Result()
	if err != nil {
		utils.Logger.Error("Failed to get messages from Redis: %v", err)
		return nil, err
	}

	var messages []models.Message
	for _, msg := range messagesJSON {
		var message models.Message
		if err := json.Unmarshal([]byte(msg), &message); err != nil {
			utils.Logger.Error("Error decoding Redis message: %v", err)
			continue
		}
		messages = append(messages, message)
	}

	utils.Logger.Info("Loaded %d messages from Redis for conversationID %s", len(messages), conversationID)
	return messages, nil
}

// MoveAllConversationsToMongo finds all conversations (Used in Gracefule Shutdown)
func (r *RedisMessageRepository) MoveAllConversationsToMongo() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find all Redis keys matching pattern
	keys, err := r.RedisChatDB.Keys(ctx, "messages:*").Result()
	if err != nil {
		utils.Logger.Error("Failed to list Redis keys: %v", err)
		return err
	}

	if len(keys) == 0 {
		utils.Logger.Warn("No conversations found in Redis")
		return nil
	}

	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = r.RedisChatDB.Scan(ctx, cursor, "messages:*", 100).Result()
		if err != nil {
			return err
		}

		for _, key := range keys {
			convoID := key[len("messages:"):]
			_ = r.MoveConvToMongo(convoID)
		}

		if cursor == 0 {
			break
		}
	}

	utils.Logger.Info("Migrated %d conversations from Redis to MongoDB", len(keys))
	return nil
}

// MoveConversationToMongo moves messages from Redis to MongoDB
func (r *RedisMessageRepository) MoveConvToMongo(conversationID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	redisKey := fmt.Sprintf("messages:%s", conversationID)

	// Retrieve messages from Redis
	messagesJSON, err := r.RedisChatDB.LRange(ctx, redisKey, 0, -1).Result()
	if err != nil {
		utils.Logger.Error("Failed to fetch messages from Redis: %v", err)
		return err
	}

	// Convert JSON strings to Message objects
	var messages []models.Message
	for _, msg := range messagesJSON {
		var message models.Message
		if err := json.Unmarshal([]byte(msg), &message); err != nil {
			utils.Logger.Error("Error decoding Redis message: %v", err)
			continue
		}
		messages = append(messages, message)
	}

	// Ensure messages exist before saving
	if len(messages) == 0 {
		utils.Logger.Warn("No messages found in Redis for conversationID %s. Skipping migration.", conversationID)
		return nil
	}

	// Upsert each message in MongoDB
	for _, msg := range messages {
		filter := bson.M{"message_id": msg.MessageID} // Check if message already exists
		update := bson.M{
			"$set": bson.M{
				"user_id":         msg.UserID,
				"conversation_id": msg.ConversationID,
				"question":        msg.Question,
				"answer":          msg.Answer,
				"thumbup":         msg.ThumbUp,
				"feedback":        msg.Feedback,
				"input_url":       msg.InputURL,
				"output_url":      msg.OutputURL,
				"created_at":      msg.CreatedAt,
			},
		}

		// Perform upsert (update if exists, insert if not)
		_, err := r.MongoMsgCol.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
		if err != nil {
			utils.Logger.Error("Failed to upsert message to MongoDB: %v", err)
			return err
		}
	}

	// Delete from Redis after successful migration
	_, err = r.RedisChatDB.Del(ctx, redisKey).Result()
	if err != nil {
		utils.Logger.Error("Failed to delete Redis conversation: %v", err)
		return err
	}

	utils.Logger.Info("Moved conversation %s to MongoDB", conversationID)
	return nil
}
