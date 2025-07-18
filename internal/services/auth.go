package services

import (
	"chat-ai-backend/internal/models"
	"chat-ai-backend/internal/repositories"
	"chat-ai-backend/utils"
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// services/auth_service.go
type AuthService struct {
	Repo *repositories.UserRepository
}

func NewAuthService(repo *repositories.UserRepository) *AuthService {
	return &AuthService{Repo: repo}
}

// RegisterUser handles user registration.
func (s *AuthService) RegisterUser(username, password, email string) error {
	utils.Logger.Info("Starting user registration process")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if the username or email already exists
	existingUser, err := s.Repo.CheckUserExists(ctx, username, email)
	if err != nil {
		utils.Logger.Error("Unexpected error occurred while checking user existence: %v", err)
		return err
	}

	if existingUser != nil {
		if existingUser.Username == username {
			utils.Logger.Warn("Username '%s' already exists", username)
			return errors.New("username already exists")
		}
		if existingUser.Email == email {
			utils.Logger.Warn("Email '%s' already exists", email)
			return errors.New("email already exists")
		}
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		utils.Logger.Error("Error hashing password: %v", err)
		return err
	}

	// Create a new user object
	user := models.User{
		Username:     username,
		Password:     string(hashedPassword),
		Email:        email,
		RegisteredAt: time.Now(),
	}

	// Insert the new user into the database
	err = s.Repo.InsertUser(ctx, user)
	if err != nil {
		utils.Logger.Error("Error inserting user into the database: %v", err)
		return err
	}

	utils.Logger.Info("User '%s' registered successfully", username)
	return nil
}

// LoginUser handles user authentication and updates the last login timestamp.
func (s *AuthService) LoginUser(email string, password string, expirationAccess time.Duration, expirationRefresh time.Duration) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find the user in the database
	user, err := s.Repo.FindUserByUserEmail(ctx, email)
	if err != nil {
		if err.Error() == "user not found" {
			utils.Logger.Error("Login failed: user '%s' not found", email)
			return "", "", errors.New("invalid email or password")
		}
		utils.Logger.Error("Error retrieving user '%s': %v", email, err)
		return "", "", err
	}

	// Compare the provided password with the stored hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		utils.Logger.Error("Login failed: incorrect password for user '%s'", email)
		return "", "", errors.New("invalid username or password")
	}

	// Generate a access JWT token
	accessToken, err := utils.GenerateJWT(user.ID, user.Email, expirationAccess)
	if err != nil {
		utils.Logger.Error("Failed to generate JWT token for user '%s': %v", email, err)
		return "", "", err
	}

	// Generate a refresh JWT token
	refreshToken, err := utils.GenerateJWT(user.ID, user.Email, expirationRefresh)
	if err != nil {
		utils.Logger.Error("Failed to generate refresh token for user '%s': %v", email, err)
		return "", "", err
	}

	// Update the last login timestamp
	err = s.Repo.UpdateLastLogin(ctx, email)
	if err != nil {
		utils.Logger.Error("Failed to update last login timestamp for user '%s': %v", email, err)
		return "", "", errors.New("failed to update last login timestamp")
	}

	// Store the refresh token in the database
	err = s.Repo.StoreTokenRedis(ctx, email, refreshToken, expirationRefresh)
	if err != nil {
		utils.Logger.Error("Failed to store refresh token for user '%s': %v", email, err)
		return "", "", errors.New("failed to store refresh token")
	}

	return accessToken, refreshToken, nil
}

// LogoutUser invalidates the user's refresh token
func (s *AuthService) LogoutUser(email string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Remove the refresh token from redis
	err := s.Repo.DeleteRefreshTokenRedis(ctx, email)
	if err != nil {
		utils.Logger.Error("Failed to invalidate refresh token for user '%s': %v", email, err)
		return errors.New("failed to invalidate refresh token")
	}

	return nil
}

func (s *AuthService) CheckToken(token string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Validate the JWT token
	claims, err := utils.ValidateJWT(token)
	if err != nil {
		utils.Logger.Error("Invalid token: %v", err)
		return false, errors.New("invalid token")
	}

	// Check if the token is blacklisted in Redis
	exist, err := s.Repo.CheckTokenInRedis(ctx, claims.Subject, token)
	if err != nil {
		utils.Logger.Error("Error checking token blacklist status: %v", err)
		return false, err
	}
	return exist, nil
}

// RefreshAccessToken validates refresh token & issues new tokens
func (s *AuthService) RefreshAccessToken(refreshToken string) (string, error) {
	// Validate the refresh token (extract user claims)
	claims, err := utils.ValidateJWT(refreshToken)
	if err != nil {
		return "", errors.New("invalid refresh token")
	}

	// Check if refresh token is still valid in DB
	storedToken, err := s.Repo.GetRefreshToken(context.Background(), claims.Subject)
	if err != nil {
		return "", errors.New("could not retrieve refresh token from DB")
	}

	// Ensure refresh token is not empty ("")
	if storedToken == "" {
		return "", errors.New("refresh token has been revoked")
	}

	// Ensure refresh token matches the one in DB
	if storedToken != refreshToken {
		return "", errors.New("refresh token mismatch")
	}

	// Generate new access token (valid for 15 min)
	newAccessToken, err := utils.GenerateJWT(claims.ID, claims.Subject, time.Minute*15)
	if err != nil {
		return "", err
	}

	// Keep the same refresh token (no refresh token rotation)
	return newAccessToken, nil
}
