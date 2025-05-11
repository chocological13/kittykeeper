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
