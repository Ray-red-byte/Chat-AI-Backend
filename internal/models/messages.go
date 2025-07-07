// internal/models/messages.go

package models

import "time"

// Conversation represents a chat session.
type Conversation struct {
	ID        string    `bson:"_id,omitempty"` // MongoDB auto-generates this field
	UserID    string    `bson:"user_id"`       // ID of the user owning the conversation
	Title     string    `bson:"title"`         // Conversation title
	CreatedAt time.Time `bson:"created_at"`    // When the conversation was created
}

// Message represents a single message in a conversation.
type Message struct {
	ID             string    `bson:"_id,omitempty"`      // MongoDB auto-generates this field
	MessageID      string    `bson:"message_id"`         // Unique ID for the message
	UserID         string    `bson:"user_id"`            // ID of the user owning the conversation
	ConversationID string    `bson:"conversation_id"`    // ID of the related conversation
	Title          string    `bson:"title"`              // Title of the conversation (optional)
	Question       string    `bson:"question"`           // User's question
	Answer         string    `bson:"answer"`             // AI's response
	Feedback       *string   `bson:"feedback,omitempty"` // Feedback provided by the user (default null)
	ThumbUp        int       `bson:"thumbup"`            // Thumb feedback (-1, 0, 1)
	InputURL       string    `bson:"input_url"`          // URL for input data (if any)
	OutputURL      string    `bson:"output_url"`         // URL for output data (if any)
	CreatedAt      time.Time `bson:"created_at"`         // When the message was created
}
