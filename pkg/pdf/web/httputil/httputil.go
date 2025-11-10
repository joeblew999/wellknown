package httputil

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ValidateMethod checks if the request method matches expected
// Returns false and sends error response if method doesn't match
func ValidateMethod(w http.ResponseWriter, r *http.Request, expected string) bool {
	if r.Method != expected {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

// GetRequiredFormValue gets a required form value
// Returns empty string and sends error response if field is missing
func GetRequiredFormValue(w http.ResponseWriter, r *http.Request, fieldName string) (string, bool) {
	value := r.FormValue(fieldName)
	if value == "" {
		http.Error(w, fmt.Sprintf("%s is required", fieldName), http.StatusBadRequest)
		return "", false
	}
	return value, true
}

// GetRequiredQueryParam gets a required query parameter
// Returns empty string and sends error response if param is missing
func GetRequiredQueryParam(w http.ResponseWriter, r *http.Request, paramName string) (string, bool) {
	value := r.URL.Query().Get(paramName)
	if value == "" {
		http.Error(w, fmt.Sprintf("%s is required", paramName), http.StatusBadRequest)
		return "", false
	}
	return value, true
}

// RespondJSON sends a JSON response with the given status code
func RespondJSON(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// RespondJSONOK sends a 200 OK JSON response
func RespondJSONOK(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}

// RespondError sends an error response with the given status code and message
func RespondError(w http.ResponseWriter, status int, message string) {
	http.Error(w, message, status)
}

// RespondInternalError sends a 500 Internal Server Error response
func RespondInternalError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// RespondBadRequest sends a 400 Bad Request response
func RespondBadRequest(w http.ResponseWriter, message string) {
	http.Error(w, message, http.StatusBadRequest)
}

// RespondNotFound sends a 404 Not Found response
func RespondNotFound(w http.ResponseWriter, message string) {
	http.Error(w, message, http.StatusNotFound)
}

// DecodeJSONBody decodes JSON request body into the provided struct
// Returns error and sends error response if decoding fails
func DecodeJSONBody(w http.ResponseWriter, r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		RespondBadRequest(w, fmt.Sprintf("Invalid JSON: %v", err))
		return err
	}
	return nil
}
