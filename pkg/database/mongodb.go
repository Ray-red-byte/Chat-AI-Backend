// chatapp/pkg/database/mongodb.go

package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MongoClient            *mongo.Client
	UserCollection         *mongo.Collection
	ConversationCollection *mongo.Collection
	MessageCollection      *mongo.Collection
)

// createIndexes creates indexes for the provided collection
func createIndexes(collection *mongo.Collection, indexes []mongo.IndexModel) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, index := range indexes {
		_, err := collection.Indexes().CreateOne(ctx, index)
		if err != nil {
			log.Printf("Failed to create index for collection %s: %v\n", collection.Name(), err)
			continue
		}
		log.Printf("Created index on collection %s: %+v\n", collection.Name(), index)
	}
}

// Init initializes the MongoDB connection and collections
func InitMongo(mongoURI string) {
	clientOptions := options.Client().
		ApplyURI(mongoURI).
		SetMaxPoolSize(100).                        // Maximum number of connections in the pool
		SetMinPoolSize(10).                         // Minimum number of idle connections
		SetMaxConnIdleTime(30 * time.Minute).       // Maximum idle time for a connection
		SetSocketTimeout(10 * time.Second).         // Socket timeout duration
		SetServerSelectionTimeout(10 * time.Second) // Timeout for server selection

	// Connect to MongoDB
	var err error
	MongoClient, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v\n", err)
	}

	// Check the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := MongoClient.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB connection error: %v\n", err)
	}

	log.Println("Connected to MongoDB!")

	// Initialize database
	db := MongoClient.Database("chatapp")

	// Initialize collections
	UserCollection = db.Collection("users")
	ConversationCollection = db.Collection("conversations")
	MessageCollection = db.Collection("messages")

	// Create indexes for collections
	log.Println("Creating indexes for collections...")
	createIndexes(UserCollection, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}}, // Unique index on "email"
			Options: options.Index().SetUnique(true),
		},
	})
	createIndexes(ConversationCollection, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
	})
	createIndexes(MessageCollection, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "message_id", Value: 1}}, // Index on "conversation_id" for quick lookup
		},
	})
	log.Println("Collections and indexes initialized successfully!")
}

func CloseMongo() {
	if MongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := MongoClient.Disconnect(ctx); err != nil {
			log.Printf("Failed to disconnect from MongoDB: %v\n", err)
		} else {
			log.Println("Disconnected from MongoDB successfully!")
		}
	}
}

func GetMongoClient() *mongo.Client {
	if MongoClient == nil {
		log.Fatal("MongoDB client is not initialized. Call InitMongo first.")
	}
	return MongoClient
}
