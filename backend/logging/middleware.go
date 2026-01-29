package logging

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// responseWriter wraps http.ResponseWriter to capture status code and size
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return h.Hijack()
}

// HTTPLoggingMiddleware logs all HTTP requests with structured logging
func HTTPLoggingMiddleware(logger *Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Generate or extract request ID
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Wrap response writer to capture status
			rw := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			// Set request ID in response header
			w.Header().Set("X-Request-ID", requestID)

			// Add request ID to context
			ctx := ContextWithRequestID(r.Context(), requestID)
			r = r.WithContext(ctx)

			// Extract user/account from headers or JWT (if available)
			userID := r.Header.Get("X-User-ID")
			accountID := r.Header.Get("X-Account-ID")

			if userID != "" {
				ctx = ContextWithUserID(ctx, userID)
				r = r.WithContext(ctx)
			}

			if accountID != "" {
				ctx = ContextWithAccountID(ctx, accountID)
				r = r.WithContext(ctx)
			}

			// Log request
			logger.Info("HTTP Request",
				RequestID(requestID),
				String("method", r.Method),
				String("path", r.URL.Path),
				String("remote_addr", r.RemoteAddr),
				String("user_agent", r.UserAgent()),
				String("proto", r.Proto),
			)

			// Process request
			next.ServeHTTP(rw, r)

			// Calculate duration
			duration := time.Since(start).Milliseconds()

			// Determine log level based on status code
			logLevel := INFO
			if rw.status >= 500 {
				logLevel = ERROR
			} else if rw.status >= 400 {
				logLevel = WARN
			}

			// Check for slow requests (>1s)
			if duration > 1000 {
				logger.Warn("Slow HTTP Request",
					RequestID(requestID),
					String("method", r.Method),
					String("path", r.URL.Path),
					Int("status", rw.status),
					Int64("duration_ms", duration),
					Int("size_bytes", rw.size),
					String("slow_request", "true"),
				)
			}

			// Log response
			fields := []Field{
				RequestID(requestID),
				String("method", r.Method),
				String("path", r.URL.Path),
				Int("status", rw.status),
				Int64("duration_ms", duration),
				Int("size_bytes", rw.size),
			}

			if userID != "" {
				fields = append(fields, UserID(userID))
			}
			if accountID != "" {
				fields = append(fields, AccountID(accountID))
			}

			switch logLevel {
			case ERROR:
				logger.Error("HTTP Response Error", nil, fields...)
			case WARN:
				logger.Warn("HTTP Response Warning", fields...)
			default:
				logger.Info("HTTP Response", fields...)
			}
		})
	}
}

// PanicRecoveryMiddleware recovers from panics and logs them
func PanicRecoveryMiddleware(logger *Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					requestID := r.Header.Get("X-Request-ID")
					if requestID == "" {
						requestID = uuid.New().String()
					}

					logger.Error("Panic recovered",
						nil,
						RequestID(requestID),
						String("method", r.Method),
						String("path", r.URL.Path),
						String("panic", fmt.Sprint(err)),
						String("stack_trace", getStackTrace()),
					)

					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// CORSLoggingMiddleware logs CORS-related issues
func CORSLoggingMiddleware(logger *Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && r.Method == "OPTIONS" {
				requestID := r.Header.Get("X-Request-ID")
				logger.Debug("CORS Preflight Request",
					RequestID(requestID),
					String("method", r.Method),
					String("path", r.URL.Path),
					String("origin", origin),
				)
			}

			next.ServeHTTP(w, r)
		})
	}
}
