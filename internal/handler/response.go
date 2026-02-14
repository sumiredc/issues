package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/sumire/issues/internal/domain"
)

// Envelope is the standard API response wrapper.
type Envelope struct {
	Data  any            `json:"data,omitempty"`
	Meta  *PaginationMeta `json:"meta,omitempty"`
	Error *APIError      `json:"error,omitempty"`
}

// PaginationMeta holds cursor-based pagination info.
type PaginationMeta struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasNext    bool   `json:"has_next"`
}

// APIError represents an error in the API response.
type APIError struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []FieldError  `json:"details,omitempty"`
}

// FieldError represents a field-level validation error.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(Envelope{Data: data}); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

// WriteJSONList writes a paginated JSON list response.
func WriteJSONList(w http.ResponseWriter, status int, data any, meta PaginationMeta) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(Envelope{Data: data, Meta: &meta}); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

// WriteError maps domain errors to HTTP responses and writes them.
func WriteError(w http.ResponseWriter, err error) {
	status, apiErr := mapError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if encErr := json.NewEncoder(w).Encode(Envelope{Error: &apiErr}); encErr != nil {
		slog.Error("failed to encode error response", "error", encErr)
	}
}

func mapError(err error) (int, APIError) {
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
