package handlers

import (
	"chat-app/config"
	"chat-app/internal/models"
	"chat-app/internal/services"
	"chat-app/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	AuthService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{AuthService: authService}
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {

	var input models.UserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Call the service to create a new user
	err := h.AuthService.RegisterUser(input.Username, input.Password, input.Email)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// Respond with success
	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

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

// LogoutHandler invalidates the user's JWT token and clears the auth_token cookie
func (h *AuthHandler) LogoutHandler(c *gin.Context) {
	// Retrieve refresh token from cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
		return
	}

	// Validate the refresh token
	claims, err := utils.ValidateJWT(refreshToken)
	if err != nil {
		// Even if token is invalid, still delete cookies
		c.SetCookie("access_token", "", -1, "/", "", false, true)
		c.SetCookie("refresh_token", "", -1, "/", "", false, true)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	// Update refresh_tokoen field in mongo DB
	err = h.AuthService.LogoutUser(claims.Subject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not invalidate refresh token"})
		return
	}

	// Delete both cookies
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "successfully logged out",
		"email":   claims.Subject,
	})
}

// RefreshTokenHandler handles issuing new access tokens using a valid refresh token
func (h *AuthHandler) RefreshTokenHandler(c *gin.Context) {
	// Retrieve refresh token from cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh token"})
		return
	}

	// Validate refresh token
	newAccessToken, err := h.AuthService.RefreshAccessToken(refreshToken)
	if err != nil || newAccessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	// Respond with new access token
	c.JSON(http.StatusOK, gin.H{
		"message":      "Token refreshed successfully",
		"access_token": newAccessToken,
	})
}
