// internal/repositories/UserRepository.go

package repositories

import (
	"chat-app/internal/models"
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository struct {
	Collection  *mongo.Collection
	RedisUserDB *redis.Client
}

func NewUserRepository(collection *mongo.Collection, redisClient *redis.Client) *UserRepository {
	return &UserRepository{
		Collection:  collection,
		RedisUserDB: redisClient,
	}
}

// CheckUserExists checks if a user with the given username or email already exists.
func (r *UserRepository) CheckUserExists(ctx context.Context, username, email string) (*models.User, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"username": username},
			{"email": email},
		},
	}
	var user models.User
	err := r.Collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No user found
		}
		return nil, err // Other errors
	}
	return &user, nil
}

// InsertUser inserts a new user into the database.
func (r *UserRepository) InsertUser(ctx context.Context, user models.User) error {
	_, err := r.Collection.InsertOne(ctx, user)
	if mongo.IsDuplicateKeyError(err) {
		return errors.New("username or email already exists")
	}
	return err
}

// FindUserByUsername retrieves a user by username.
func (r *UserRepository) FindUserByUserEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.Collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user email not found")
		}
		return nil, err
	}
	return &user, nil
}

// UpdateLastLogin updates the last login timestamp for a user.
func (r *UserRepository) UpdateLastLogin(ctx context.Context, email string) error {
	_, err := r.Collection.UpdateOne(
		ctx,
		bson.M{"email": email},
		bson.M{"$set": bson.M{"last_login": time.Now()}},
	)
	if err != nil {
		return err
	}
	return nil
}

// Insert refresh token into the database
func (r *UserRepository) InsertRefreshToken(ctx context.Context, email, token string) error {
	filter := bson.M{"email": email}
	update := bson.M{"$set": bson.M{"refresh_token": token}}

	opts := options.Update().SetUpsert(true)

	_, err := r.Collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}

	return nil
}

// GetRefreshToken retrieves a refresh token from MongoDB
func (r *UserRepository) GetRefreshToken(ctx context.Context, email string) (string, error) {
	var result struct {
		RefreshToken string `bson:"refresh_token"`
	}

	filter := bson.M{"email": email}
	err := r.Collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", errors.New("refresh token not found")
		}
		return "", err
	}

	return result.RefreshToken, nil
}

func (r *UserRepository) StoreTokenRedis(ctx context.Context, email string, token string, expirationRefresh time.Duration) error {

	key := "refresh_token:" + email
	err := r.RedisUserDB.Set(ctx, key, token, expirationRefresh).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) DeleteRefreshTokenRedis(ctx context.Context, email string) error {
	key := "refresh_token:" + email
	return r.RedisUserDB.Del(ctx, key).Err()
}

func (r *UserRepository) CheckTokenInRedis(ctx context.Context, email, token string) (bool, error) {
	key := "refresh_token:" + email

	storedToken, err := r.RedisUserDB.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil // No token stored = treated as invalid
	}
	if err != nil {
		return false, err // Redis connection or command error
	}
	return storedToken == token, nil
}
