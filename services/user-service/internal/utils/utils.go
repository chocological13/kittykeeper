package utils

import (
	"errors"
	"fmt"
	"github.com/chocological13/kittykeeper/services/user-service/internal/database/repository"
	"github.com/chocological13/kittykeeper/services/user-service/internal/models"
	"github.com/go-playground/validator/v10"
	"net/mail"
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

func formatTagMessage(tag, field, param string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("'%s' is required", field)
	case "email":
		return fmt.Sprintf("'%s' is not a valid email", field)
	case "min":
		return fmt.Sprintf("'%s' must be at least %s", field, param)
	case "max":
		return fmt.Sprintf("'%s' must not be longer than %s characters", field, param)
	default:
		return fmt.Sprintf("'%s' is invalid", field)
	}
}

// IsEmail to check if the provided credential is a username or an email
func IsEmail(credential string) bool {
	_, err := mail.ParseAddress(credential)
	return err == nil
}
