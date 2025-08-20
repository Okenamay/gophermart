package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var Zap *zap.SugaredLogger

// InitLogger initializes the global logger.
func InitLogger() error {
	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	Zap = logger.Sugar()
	return nil
}

type (
	// responseData captures details about the HTTP response.
	responseData struct {
		status int
		size   int
	}
	// loggingResponseWriter wraps http.ResponseWriter to capture status and size.
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// WithLogging is a middleware that logs details about each HTTP request.
func WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &responseData{status: 0, size: 0}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r)
		duration := time.Since(start)

		Zap.Infow("Request handled",
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	}
	return http.HandlerFunc(logFn)
}
