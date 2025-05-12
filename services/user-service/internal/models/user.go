package models

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID        uuid.UUID
	Username  string
	Email     string
	FirstName string
	LastName  string
	CreatedAt time.Time
	EditedAt  time.Time
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type CreateUserParams struct {
	Username  string
	Email     string
	Password  string
	FirstName string
	LastName  string
}

type LoginParams struct {
	Credential string // <-- this can be username or email
	Password   string
}
