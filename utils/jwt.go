// utils/jwt.go

package utils

import (
	"errors"
	"log"
	"time"

	"chat-app/config"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey []byte

// InitializeJWT initializes the JWT key from the configuration
func InitializeJWT() {
	if config.AppConfig.JWTSecretKey == "" {
		log.Fatal("JWT secret key is not configured")
	}
	jwtKey = []byte(config.AppConfig.JWTSecretKey)
	log.Println("JWT secret key initialized successfully")
}

// GenerateJWT generates a new JWT token
func GenerateJWT(userId string, email string, duration time.Duration) (string, error) {
	claims := &jwt.RegisteredClaims{
		ID:        userId,
		Subject:   email,                                        // Stores the username as the subject
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)), // Token expires customized duration from now
		IssuedAt:  jwt.NewNumericDate(time.Now()),               // Token issuance time
		Issuer:    "chatapp",                                    // Issuer identifier
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ValidateJWT validates a given JWT token
func ValidateJWT(tokenString string) (*jwt.RegisteredClaims, error) {
	// Parse the token with claims
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		log.Println("Parsing token claims...")
		return jwtKey, nil
	})

	if err != nil {
		log.Printf("Error parsing token: %v\n", err)
		if errors.Is(err, jwt.ErrSignatureInvalid) {
			log.Println("Invalid token signature")
			return nil, errors.New("invalid token signature")
		}
		return nil, err
	}

	// Extract the claims and verify the token is valid
	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		log.Println("Token is valid")
		log.Printf("Token Claims - Subject: %s, Issuer: %s, ExpiresAt: %v\n", claims.Subject, claims.Issuer, claims.ExpiresAt)
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
