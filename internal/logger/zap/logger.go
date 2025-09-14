package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

// InitLogger инициализирует и возвращает новый экземпляр логгера, глобальная
// переменная Zap больше не используется
func InitLogger() (*zap.SugaredLogger, error) {
	appLogger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return appLogger.Sugar(), nil
}

type (
	responseData struct {
		status int
		size   int
	}
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

// WithLogging - middleware для логиррования HTTP запросов
func WithLogging(appLogger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			responseData := &responseData{status: 0, size: 0}
			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   responseData,
			}
			h.ServeHTTP(&lw, r)
			duration := time.Since(start)

			appLogger.Infow("Request handled",
				"uri", r.RequestURI,
				"method", r.Method,
				"status", responseData.status,
				"duration", duration,
				"size", responseData.size,
			)
		})
	}
}
