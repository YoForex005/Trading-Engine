package httputil

import (
	"encoding/json"
	"net/http"

	"github.com/epic1st/rtx/backend/internal/logging"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error  string `json:"error"`
	Status int    `json:"status"`
}

// SuccessResponse represents a standard success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// RespondWithError sends a JSON error response
func RespondWithError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := ErrorResponse{
		Error:  message,
		Status: status,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logging.Default.Error("failed to encode error response",
			"status", status,
			"message", message,
			"error", err,
		)
	}
}

// RespondWithJSON sends a JSON response with the given data
func RespondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logging.Default.Error("failed to encode JSON response",
			"status", status,
			"error", err,
		)
		// Try to send error response as fallback
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:  "Internal server error",
			Status: http.StatusInternalServerError,
		})
	}
}

// RespondWithSuccess sends a JSON success response with optional data
func RespondWithSuccess(w http.ResponseWriter, message string, data interface{}) {
	response := SuccessResponse{
		Message: message,
		Data:    data,
	}
	RespondWithJSON(w, http.StatusOK, response)
}

// DecodeJSONBody decodes JSON request body into the provided target
func DecodeJSONBody(w http.ResponseWriter, r *http.Request, target interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body")
		return false
	}
	return true
}
