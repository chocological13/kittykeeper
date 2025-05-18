package models

import (
	"fmt"
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

type UpdateCatRequest struct {
	Name                string   `json:"name"`
	Breed               *string  `json:"breed"`
	DateOfBirth         *string  `json:"date_of_birth"`
	Weight              *float64 `json:"weight" binding:"omitempty,numeric,min=0,max=100"`
	Color               *string  `json:"color"`
	Gender              *string  `json:"gender" binding:"omitempty,oneof=male female unknown"`
	PhotoUrl            *string  `json:"photo_url"`
	MedicalNotes        *string  `json:"medical_notes"`
	DietaryRequirements *string  `json:"dietary_requirements"`
	DateOfDeath         *string  `json:"date_of_death"`
}

type CatRequestParams struct {
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

// ? Helper
func (c *CreateCatRequest) ToParams() (CatRequestParams, error) {
	var dob *time.Time
	if c.DateOfBirth != nil && *c.DateOfBirth != "" {
		d, err := time.Parse("2006-01-02", *c.DateOfBirth)
		if err != nil {
			return CatRequestParams{}, fmt.Errorf("invalid date of birth: %w", err)
		}
		dob = &d
	}

	if dob != nil && dob.After(time.Now()) {
		return CatRequestParams{}, fmt.Errorf("date of birth cannot be in the future")
	}

	return CatRequestParams{
		Name:                c.Name,
		Breed:               c.Breed,
		DateOfBirth:         dob,
		Weight:              c.Weight,
		Color:               c.Color,
		Gender:              c.Gender,
		PhotoUrl:            c.PhotoUrl,
		MedicalNotes:        c.MedicalNotes,
		DietaryRequirements: c.DietaryRequirements,
	}, nil
}

func (u *UpdateCatRequest) ToParams() (CatRequestParams, error) {
	var dob *time.Time
	if u.DateOfBirth != nil && *u.DateOfBirth != "" {
		d, err := time.Parse("2006-01-02", *u.DateOfBirth)
		if err != nil {
			return CatRequestParams{}, fmt.Errorf("invalid date of birth: %w", err)
		}
		dob = &d
	}

	var dod *time.Time
	if u.DateOfDeath != nil && *u.DateOfDeath != "" {
		d, err := time.Parse("2006-01-02", *u.DateOfDeath)
		if err != nil {
			return CatRequestParams{}, fmt.Errorf("invalid date of death: %w", err)
		}
		dod = &d
	}

	if dob != nil && dob.After(time.Now()) {
		return CatRequestParams{}, fmt.Errorf("date of birth cannot be in the future")
	}
	if dod != nil && dod.After(time.Now()) {
		return CatRequestParams{}, fmt.Errorf("date of death cannot be in the future")
	}

	return CatRequestParams{
		Name:                u.Name,
		Breed:               u.Breed,
		DateOfBirth:         dob,
		Weight:              u.Weight,
		Color:               u.Color,
		Gender:              u.Gender,
		PhotoUrl:            u.PhotoUrl,
		MedicalNotes:        u.MedicalNotes,
		DietaryRequirements: u.DietaryRequirements,
	}, nil
}
