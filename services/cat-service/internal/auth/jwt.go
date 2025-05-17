package auth

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type TokenClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

var (
	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("expired token")
	ErrInvalidTokenClaims = errors.New("invalid token claims")
)

type JWTService struct {
	accessTokenSecret string
}

func NewJWTService(accessTokenSecret string) *JWTService {
	return &JWTService{
		accessTokenSecret: accessTokenSecret,
	}
}

// VerifyToken verifies a JWT token and returns the claims
func (s *JWTService) VerifyToken(tokenString string) (*TokenClaims, error) {
	// * Parse the token string and get standard claims
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, keyFunc(s.accessTokenSecret))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	// * Parse the token string and get custom claims
	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidTokenClaims
	}

	return claims, nil
}

// ExtractUserID extracts the user ID from a JWT token string
func (s *JWTService) ExtractUserID(tokenString string) (uuid.UUID, error) {
	claims, err := s.VerifyToken(tokenString)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}

// ? Helpers
// keyFunc returns a jwt.Keyfunc for verifying tokens with the correct secret
func keyFunc(secret string) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	}
}
