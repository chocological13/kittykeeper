package utils

import (
	"errors"
	"fmt"
	"github.com/chocological13/kittykeeper/cat-service/internal/database/repository"
	"github.com/chocological13/kittykeeper/cat-service/internal/models"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"strings"
	"time"
)

func FromDBCat(dbCat repository.Cat) models.CatResponse {
	return models.CatResponse{
		ID:                  dbCat.ID,
		OwnerID:             dbCat.OwnerID,
		Name:                dbCat.Name,
		Breed:               dbCat.Breed,
		DateOfBirth:         pgDateToPtr(dbCat.DateOfBirth),
		Weight:              pgNumericToPtr(dbCat.Weight),
		Color:               dbCat.Color,
		Gender:              dbCat.Gender,
		PhotoURL:            dbCat.PhotoUrl,
		MedicalNotes:        dbCat.MedicalNotes,
		DietaryRequirements: dbCat.DietaryRequirements,
		DateOfDeath:         pgDateToPtr(dbCat.DateOfDeath),
		CreatedAt:           dbCat.CreatedAt,
		UpdatedAt:           dbCat.UpdatedAt,
	}
}

func PtrToPgDate(date *time.Time) pgtype.Date {
	if date != nil {
		return pgtype.Date{
			Time:  *date,
			Valid: true,
		}
	}
	return pgtype.Date{}
}

func PtrToPgNumeric(f *float64) (pgtype.Numeric, error) {
	if f == nil {
		return pgtype.Numeric{}, nil
	}

	var num pgtype.Numeric
	err := num.Scan(*f)
	if err != nil {
		return pgtype.Numeric{}, err
	}
	return num, nil
}

func FormatValidationError(err error) map[string]string {
	formatedErrors := make(map[string]string)

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		for _, err := range validationErrors {
			field := strings.ToLower(err.Field())
			param := strings.ToLower(err.Param())
			formatedErrors[field] = formatTagMessage(err.Tag(), field, param)
		}
	}

	return formatedErrors
}

// Helpers
func pgDateToPtr(t pgtype.Date) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}

func pgNumericToPtr(t pgtype.Numeric) *float64 {
	if t.Valid {
		f, _ := t.Float64Value()
		return &f.Float64
	}
	return nil
}

func formatTagMessage(tag, field, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("'%s' is required", field)
	case "oneof":
		return fmt.Sprintf("'%s' must be one of: %s", field, param)
	case "numeric":
		return fmt.Sprintf("'%s' must be numeric", field)
	case "time_format":
		return fmt.Sprintf("'%s' must be a valid time as so: %s", field, param)
	case "min":
		return fmt.Sprintf("'%s' must be at least %s", field, param)
	case "max":
		return fmt.Sprintf("'%s' must not be longer than %s characters", field, param)
	default:
		return fmt.Sprintf("'%s' is invalid", field)
	}
}
