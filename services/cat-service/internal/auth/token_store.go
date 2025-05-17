package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type TokenStore struct {
	client *redis.Client
}

func NewTokenStore(client *redis.Client) *TokenStore {
	return &TokenStore{
		client: client,
	}
}

func (s *TokenStore) GetAccessToken(ctx context.Context, userID uuid.UUID) (string, error) {
	key := fmt.Sprintf("access:%s", userID.String())

	token, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", err
		}
	}

	return token, nil
}
