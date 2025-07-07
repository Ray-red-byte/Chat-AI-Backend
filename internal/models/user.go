// internal/models/user.go

package models

import "time"

type User struct {
	ID           string    `bson:"_id,omitempty"` // MongoDB generates a unique ObjectID
	Username     string    `bson:"username"`      // Ensure unique index on this field
	Password     string    `bson:"password"`      // Store hashed password
	Email        string    `bson:"email"`
	LastLogin    time.Time `bson:"last_login"`
	RegisteredAt time.Time `bson:"registered_at"`
}

type UserInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"` // Includes basic email format validation
}
