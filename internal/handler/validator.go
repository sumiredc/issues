package handler

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/sumire/issues/internal/domain"
)

// AppValidator wraps go-playground/validator for echo.
type AppValidator struct {
	validator *validator.Validate
}

// NewAppValidator creates a new AppValidator.
func NewAppValidator() *AppValidator {
	return &AppValidator{validator: validator.New()}
}

// Validate validates a struct using go-playground/validator tags.
func (v *AppValidator) Validate(i any) error {
	if err := v.validator.Struct(i); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if ok && len(validationErrors) > 0 {
			fe := validationErrors[0]
			return &domain.ValidationError{
				Field:   fe.Field(),
				Message: fmt.Sprintf("failed on '%s' validation", fe.Tag()),
			}
		}
		return fmt.Errorf("%w: %v", domain.ErrInvalidInput, err)
	}
	return nil
}
