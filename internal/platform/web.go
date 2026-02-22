package platform

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// APIError represents the standard JSON error response format.
type APIError struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains the specifics of an API error.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	// Details could be added here later for validation specifics if needed.
}

// RespondJSON writes a structured JSON payload to the response.
func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Default().Error("failed to encode json response", "error", err)
	}
}

// RespondError writes a standardized APIError JSON payload.
func RespondError(w http.ResponseWriter, status int, code string, message string) {
	errResp := APIError{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	}
	RespondJSON(w, status, errResp)
}

// RespondValidationError formats go-playground/validator errors
func RespondValidationError(w http.ResponseWriter, err error) {
	// A simple approach: grab the first error for the message, or format them all.
	// For this spike, returning a 400 with a generic validation failed message and the error string.
	RespondError(w, http.StatusBadRequest, "VALIDATION_FAILED", err.Error())
}
