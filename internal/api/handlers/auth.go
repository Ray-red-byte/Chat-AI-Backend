package handlers

import (
	"chat-app/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Login handles user login and returns a JWT token
func (h *AuthHandler) Login(c *gin.Context) {
	// Bind JSON input to the struct
	var input struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Authenticate user and get the JWT token
	expirationAccess := config.AppConfig.AccessTokenDuration
	expirationRefresh := config.AppConfig.RefreshTokenDuration
	accessToken, refreshToken, err := h.AuthService.LoginUser(input.Email, input.Password, expirationAccess, expirationRefresh)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Set the token in an HttpOnly cookie
	c.SetCookie("access_token", accessToken, int(expirationAccess.Seconds()), "/", "", false, true)
	c.SetCookie("refresh_token", refreshToken, int(expirationRefresh.Seconds()), "/", "", false, true)

	// Respond with a success message
	c.JSON(http.StatusOK, gin.H{
		"message":      "Login successful",
		"access_token": accessToken, // Frontend stores this in localStorage
	})
}
