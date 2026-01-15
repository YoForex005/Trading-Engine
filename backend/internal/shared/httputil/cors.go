package httputil

import "net/http"

// SetCORSHeaders sets standard CORS headers on the response writer
func SetCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// HandleOPTIONS handles CORS preflight OPTIONS requests
// Returns true if request was OPTIONS and was handled
func HandleOPTIONS(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == "OPTIONS" {
		SetCORSHeaders(w)
		w.WriteHeader(http.StatusOK)
		return true
	}
	return false
}

// WithCORS wraps a handler with CORS handling
func WithCORS(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		SetCORSHeaders(w)
		if HandleOPTIONS(w, r) {
			return
		}
		handler(w, r)
	}
}
