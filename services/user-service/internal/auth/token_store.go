package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	ErrTokenNotFound = errors.New("token not found")
)

type TokenStore struct {
	client          *redis.Client
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewTokenStore(client *redis.Client, accessTokenTTL, refreshTokenTTL time.Duration) *TokenStore {
	return &TokenStore{
		client:          client,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (s *TokenStore) StoreTokens(ctx context.Context, accessToken, refreshToken string, userID uuid.UUID) error {
	accessKey := fmt.Sprintf("access:%s", userID.String())
	refreshKey := fmt.Sprintf("refresh:%s", userID.String())

	pipe := s.client.Pipeline()
	pipe.Set(ctx, accessKey, accessToken, s.accessTokenTTL)
	pipe.Set(ctx, refreshKey, refreshToken, s.refreshTokenTTL)
	_, err := pipe.Exec(ctx)
	return err

}

func (s *TokenStore) GetToken(ctx context.Context, userID uuid.UUID, tokenType TokenType) (string, error) {
	var key string
	if tokenType == AccessToken {
		key = fmt.Sprintf("access:%s", userID.String())
	} else {
		key = fmt.Sprintf("refresh:%s", userID.String())
	}

	token, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrTokenNotFound
		}
		return "", err
	}

	return token, nil
}

// InvalidateTokens deletes the access and refresh tokens from redis, for log-outs
func (s *TokenStore) InvalidateTokens(ctx context.Context, userID uuid.UUID) error {
	accessKey := fmt.Sprintf("access:%s", userID.String())
	refreshKey := fmt.Sprintf("refresh:%s", userID.String())

	pipe := s.client.TxPipeline()
	pipe.Del(ctx, accessKey)
	pipe.Del(ctx, refreshKey)
	_, err := pipe.Exec(ctx)
	return err
}

// RotateTokens deletes the old access token and store the new access and refresh tokens in a trx to ensure atomicity
func (s *TokenStore) RotateTokens(ctx context.Context, userID uuid.UUID, newAccessToken, newRefreshToken string) error {
	accessKey := fmt.Sprintf("access:%s", userID.String())
	refreshKey := fmt.Sprintf("refresh:%s", userID.String())

	// Start transaction
	pipe := s.client.TxPipeline()
	pipe.Del(ctx, accessKey)
	pipe.Del(ctx, refreshKey)
	pipe.Set(ctx, accessKey, newAccessToken, s.accessTokenTTL)
	pipe.Set(ctx, refreshKey, newRefreshToken, s.refreshTokenTTL)
	_, err := pipe.Exec(ctx)
	return err
}
