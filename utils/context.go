package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetUserIDFromContext extracts userID from Gin context and handles error response if missing or invalid
func GetUserIDFromContext(c *gin.Context) (string, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		Logger.Println("userID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return "", false
	}

	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		Logger.Println("userID is not a valid string")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return "", false
	}

	return userIDStr, true
}
