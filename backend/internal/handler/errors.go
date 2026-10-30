package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/HeroicHorizon/user-management/backend/internal/apperror"
)

// writeError maps an error to an HTTP response. Messages are generic so no
// driver, host or stack details reach the client; the cause is only logged.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		status  int
		code    string
		message string
		details map[string]string
	)
	var ve *apperror.ValidationError
	switch {
	case errors.As(err, &ve):
		status, code, message, details = http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed.", ve.Fields
	case errors.Is(err, apperror.ErrInvalidID):
		status, code, message = http.StatusBadRequest, "INVALID_ID", "Invalid user ID."
	case errors.Is(err, errBadBody):
		status, code, message = http.StatusBadRequest, "BAD_REQUEST", "Malformed request body."
	case errors.Is(err, errBodyTooLarge):
		status, code, message = http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "Request body is too large."
	case errors.Is(err, apperror.ErrNotFound):
		status, code, message = http.StatusNotFound, "NOT_FOUND", "User not found."
	case errors.Is(err, apperror.ErrDuplicateEmail):
		status, code, message = http.StatusConflict, "DUPLICATE_EMAIL", "A user with this email already exists."
	case errors.Is(err, apperror.ErrUnavailable), errors.Is(err, context.DeadlineExceeded):
		status, code, message = http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Service temporarily unavailable."
	case errors.Is(err, context.Canceled):
		// client went away; nobody reads this response
		status, code, message = 499, "REQUEST_CANCELED", "Request canceled."
	default:
		slog.Error("internal error", "path", r.URL.Path, "error", err)
		status, code, message = http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error."
	}
	writeJSON(w, status, envelope{Error: &errorBody{Code: code, Message: message, Details: details}})
}

var (
	errBadBody      = errors.New("bad body")
	errBodyTooLarge = errors.New("body too large")
)
