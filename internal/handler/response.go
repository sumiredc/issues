package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sumire/issues/internal/domain"
)

// Envelope is the standard API response wrapper.
type Envelope struct {
	Data  any             `json:"data,omitempty"`
	Meta  *PaginationMeta `json:"meta,omitempty"`
	Error *APIError       `json:"error,omitempty"`
}

// PaginationMeta holds cursor-based pagination info.
type PaginationMeta struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasNext    bool   `json:"has_next"`
}

// APIError represents an error in the API response.
type APIError struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

// FieldError represents a field-level validation error.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// JSON writes a JSON response with the standard envelope.
func JSON(c echo.Context, status int, data any) error {
	return c.JSON(status, Envelope{Data: data})
}

// JSONList writes a paginated JSON list response.
func JSONList(c echo.Context, status int, data any, meta PaginationMeta) error {
	return c.JSON(status, Envelope{Data: data, Meta: &meta})
}

// HTTPErrorHandler is the global error handler for echo.
func HTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	status, apiErr := mapError(err)
	if jsonErr := c.JSON(status, Envelope{Error: &apiErr}); jsonErr != nil {
		slog.Error("failed to send error response", "error", jsonErr)
	}
}

func mapError(err error) (int, APIError) {
	// Handle echo's own HTTP errors (404, 405, etc.)
	var echoErr *echo.HTTPError
	if errors.As(err, &echoErr) {
		msg, _ := echoErr.Message.(string)
		if msg == "" {
			msg = http.StatusText(echoErr.Code)
		}
		return echoErr.Code, APIError{
			Code:    http.StatusText(echoErr.Code),
			Message: msg,
		}
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, APIError{
			Code:    "not_found",
			Message: "The requested resource was not found",
		}
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized, APIError{
			Code:    "unauthorized",
			Message: "Authentication is required",
		}
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, APIError{
			Code:    "forbidden",
			Message: "You do not have permission to perform this action",
		}
	case errors.Is(err, domain.ErrInvalidInput):
		return http.StatusBadRequest, APIError{
			Code:    "invalid_input",
			Message: "The request body is invalid",
		}
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, APIError{
			Code:    "conflict",
			Message: "The resource already exists or conflicts with current state",
		}
	default:
		var validationErr *domain.ValidationError
		if errors.As(err, &validationErr) {
			return http.StatusBadRequest, APIError{
				Code:    "validation_error",
				Message: "Validation failed",
				Details: []FieldError{
					{Field: validationErr.Field, Message: validationErr.Message},
				},
			}
		}

		slog.Error("unhandled error", "error", err)
		return http.StatusInternalServerError, APIError{
			Code:    "internal_error",
			Message: "An unexpected error occurred",
		}
	}
}
