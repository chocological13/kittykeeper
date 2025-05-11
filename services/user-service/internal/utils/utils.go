package utils

import (
	"fmt"
	"github.com/chocological13/kittykeeper/services/user-service/internal/database/repository"
	"github.com/chocological13/kittykeeper/services/user-service/internal/models"
	"github.com/go-playground/validator/v10"
	"strings"
)

// FromDBUser maps User response from db to the app's definition of User
func FromDBUser(dbUser repository.User) models.User {
	return models.User{
		ID:        dbUser.ID,
		Username:  dbUser.Username,
		Email:     dbUser.Email,
		FirstName: dbUser.FirstName.String,
		LastName:  dbUser.LastName.String,
		CreatedAt: dbUser.CreatedAt,
		EditedAt:  dbUser.EditedAt,
	}
}

// FormatValidationError formats the validation errors into a readable response
func FormatValidationError(err error) map[string]string {
	formatedErrors := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, err := range validationErrors {
			field := strings.ToLower(err.Field())
			formatedErrors[field] = formatTagMessage(err.Tag(), field)
		}
	}

	return formatedErrors
}

func formatTagMessage(tag string, field string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("'%s' is required", field)
	case "email":
		return fmt.Sprintf("'%s' is not a valid email", field)
	case "min":
		return fmt.Sprintf("'%s' is too short", field)
	case "max":
		return fmt.Sprintf("'%s' is too long", field)
	default:
		return fmt.Sprintf("'%s' is invalid", field)
	}
}
