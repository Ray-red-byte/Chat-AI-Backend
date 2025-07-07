// middleware/auth.go

package middleware

import (
	"chat-ai-backend/config"
	"chat-ai-backend/internal/services"
	"chat-ai-backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	AuthService *services.AuthService
}

// Constructor
func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{AuthService: authService}
}

// AuthMiddleware checks access and refresh token validity
func (m *AuthMiddleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Step 1: Try to get and validate access token
		accessToken, _ := c.Cookie("access_token")
		if accessToken != "" {
			if claims, err := utils.ValidateJWT(accessToken); err == nil {
				c.Set("userID", claims.ID)
				c.Next()
				return
			}
		}

		// Step 2: Try to fallback to refresh token
		refreshToken, err := c.Cookie("refresh_token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No valid access or refresh token"})
			return
		}

		claims, err := utils.ValidateJWT(refreshToken)
		if err != nil {
			// Step 3: refresh token invalid → force logout
			c.SetCookie("refresh_token", "", -1, "/", "", false, true)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Session expired, please login again"})
			return
		}

		// Step 4: check if refresh token is in Redis
		exist, _ := m.AuthService.CheckToken(refreshToken)
		if !exist {
			c.SetCookie("refresh_token", "", -1, "/", "", false, true)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Refresh token is not exist"})
			return
		}

		// Step 5: Generate a new access token
		expirationAccess := config.AppConfig.AccessTokenDuration
		newAccessToken, err := utils.GenerateJWT(claims.ID, claims.Subject, expirationAccess)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to renew access token"})
			return
		}

		// Optional: Set it in response header or cookie
		c.SetCookie("access_token", newAccessToken, int(expirationAccess.Seconds()), "/", "", false, true) // HttpOnly = true

		// Set userID in context before continuing
		c.Set("userID", claims.ID)

		// Proceed with request
		c.Next()
	}
}
