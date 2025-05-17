package models

import (
	"github.com/google/uuid"
	"time"
)

type Cat struct {
	ID                  uuid.UUID
	OwnerID             uuid.UUID
	Name                string
	Breed               *string
	DateOfBirth         *time.Time
	Weight              *float64
	Color               *string
	Gender              *string
	PhotoUrl            *string
	MedicalNotes        *string
	DietaryRequirements *string
	DateOfDeath         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           *time.Time
}

type CreateCatRequest struct {
	Name                string   `json:"name" binding:"required"`
	Breed               *string  `json:"breed"`
	DateOfBirth         *string  `json:"date_of_birth"`
	Weight              *float64 `json:"weight" binding:"omitempty,numeric,min=0,max=100"`
	Color               *string  `json:"color"`
	Gender              *string  `json:"gender" binding:"omitempty,oneof=male female unknown"`
	PhotoUrl            *string  `json:"photo_url"`
	MedicalNotes        *string  `json:"medical_notes"`
	DietaryRequirements *string  `json:"dietary_requirements"`
}

type CreateCatRequestParams struct {
	Name                string
	Breed               *string
	DateOfBirth         *time.Time
	Weight              *float64
	Color               *string
	Gender              *string
	PhotoUrl            *string
	MedicalNotes        *string
	DietaryRequirements *string
}

type CatResponse struct {
	ID                  uuid.UUID  `json:"id"`
	OwnerID             uuid.UUID  `json:"owner_id"`
	Name                string     `json:"name"`
	Breed               *string    `json:"breed,omitempty"`
	DateOfBirth         *time.Time `json:"date_of_birth,omitempty"`
	Weight              *float64   `json:"weight,omitempty"`
	Color               *string    `json:"color,omitempty"`
	Gender              *string    `json:"gender,omitempty"`
	PhotoURL            *string    `json:"photo_url,omitempty"`
	DateOfDeath         *time.Time `json:"date_of_death,omitempty"`
	MedicalNotes        *string    `json:"medical_notes,omitempty"`
	DietaryRequirements *string    `json:"dietary_requirements,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}
