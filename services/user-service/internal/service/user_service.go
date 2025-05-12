package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/chocological13/kittykeeper/services/user-service/internal/auth"
	"github.com/chocological13/kittykeeper/services/user-service/internal/database/repository"
	"github.com/chocological13/kittykeeper/services/user-service/internal/models"
	"github.com/chocological13/kittykeeper/services/user-service/internal/utils"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrInvalidCredentials    = errors.New("invalid credentials")
)

type UserService struct {
	queries    repository.Querier
	auth       *auth.AuthService
	tokenStore *auth.TokenStore
}

func NewUserService(queries repository.Querier, auth *auth.AuthService, tokenStore *auth.TokenStore) *UserService {
	return &UserService{
		queries:    queries,
		auth:       auth,
		tokenStore: tokenStore,
	}
}

func (s *UserService) RegisterUser(ctx context.Context, params models.CreateUserParams) (models.User, error) {
	// * Check if email exists
	_, err := s.queries.GetUserByEmail(ctx, params.Email)
	if err == nil {
		return models.User{}, ErrEmailAlreadyExists
	} else if !errors.Is(err, sql.ErrNoRows) {
		return models.User{}, fmt.Errorf("failed to check if email exists: %w", err)
	}

	// * Check if username exists
	_, err = s.queries.GetUserByUsername(ctx, params.Username)
	if err == nil {
		return models.User{}, ErrUsernameAlreadyExists
	} else if !errors.Is(err, sql.ErrNoRows) {
		return models.User{}, fmt.Errorf("failed to check if username exists: %w", err)
	}

	// * Hash password to save in the db
	hashedPassword, err := s.auth.HashPassword(params.Password)
	if err != nil {
		return models.User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	// * Create the user
	firstName := pgtype.Text{String: params.FirstName, Valid: params.FirstName != ""}
	lastName := pgtype.Text{String: params.LastName, Valid: params.LastName != ""}

	dbUser, err := s.queries.CreateUser(ctx, repository.CreateUserParams{
		Username:     params.Username,
		Email:        params.Email,
		PasswordHash: hashedPassword,
		FirstName:    firstName,
		LastName:     lastName,
	})
	if err != nil {
		return models.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	return utils.FromDBUser(dbUser), nil
}

func (s *UserService) Login(ctx context.Context, params models.LoginParams) (models.User, models.TokenPair, error) {
	var dbUser repository.User
	var err error

	// * Check if the cred is an email or a username
	if utils.IsEmail(params.Credential) {
		dbUser, err = s.queries.GetUserByEmail(ctx, params.Credential)
	} else {
		dbUser, err = s.queries.GetUserByUsername(ctx, params.Credential)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, models.TokenPair{}, ErrInvalidCredentials
		}
		return models.User{}, models.TokenPair{}, fmt.Errorf("failed to check if user exists: %w", err)
	}

	// * Check password
	if err := s.auth.CheckPassword(dbUser.PasswordHash, params.Password); err != nil {
		return models.User{}, models.TokenPair{}, ErrInvalidCredentials
	}

	// * Generate Tokens
	accessToken, err := s.auth.GenerateAccessToken(dbUser.ID)
	if err != nil {
		return models.User{}, models.TokenPair{}, fmt.Errorf("failed to generate access token: %w", err)
	}
	refreshToken, err := s.auth.GenerateRefreshToken(dbUser.ID)
	if err != nil {
		return models.User{}, models.TokenPair{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// * Store tokens in redis
	if err := s.tokenStore.StoreAccessToken(ctx, accessToken, dbUser.ID); err != nil {
		return models.User{}, models.TokenPair{}, fmt.Errorf("failed to store access token in redis: %w", err)
	}
	if err := s.tokenStore.StoreRefreshToken(ctx, refreshToken, dbUser.ID); err != nil {
		return models.User{}, models.TokenPair{}, fmt.Errorf("failed to store refresh token in redis: %w", err)
	}

	return utils.FromDBUser(dbUser), models.TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
