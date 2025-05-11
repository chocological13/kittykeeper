package auth

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

// TokenType represents the type of token
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// TokenClaims represents the claims of a JWT token
type TokenClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Type   TokenType
	jwt.RegisteredClaims
}

// AuthConfig holds the authentication configuration
type AuthConfig struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
}

type AuthService struct {
	config AuthConfig
}

func NewAuthService(config AuthConfig) *AuthService {
	return &AuthService{
		config: config,
	}
}

// HashPassword hashes a password using bcrypt
func (s *AuthService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// CheckPassword compares a password with a hashed password
func (s *AuthService) CheckPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("failed to compare password: %w", err)
	}
	return nil
}

// GenerateAccessToken generates a new JWT access token
func (s *AuthService) GenerateAccessToken(userID uuid.UUID) (string, error) {
	return s.generateToken(userID, AccessToken, s.config.AccessTokenSecret, s.config.AccessTokenTTL)
}

// GenerateRefreshToken generates a new JWT refresh token
func (s *AuthService) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	return s.generateToken(userID, RefreshToken, s.config.RefreshTokenSecret, s.config.RefreshTokenTTL)
}

// VerifyToken verifies a JWT token and returns the claims
func (s *AuthService) VerifyToken(tokenString string, tokenType TokenType) (*TokenClaims, error) {
	var secret string
	if tokenType == AccessToken {
		secret = s.config.AccessTokenSecret
	} else {
		secret = s.config.RefreshTokenSecret
	}

	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, keyFunc(secret))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid || claims.Type != tokenType {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// generateToken is a helper function to generate a JWT token
func (s *AuthService) generateToken(userID uuid.UUID, tokenType TokenType, secret string, expiry time.Duration) (string,
	error) {
	claims := TokenClaims{
		UserID: userID,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "kittykeeper",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signedToken, nil
}

// keyFunc returns a jwt.Keyfunc for verifying tokens with the correct secret
func keyFunc(secret string) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	}
}
