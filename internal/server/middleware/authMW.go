package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Okenamay/gophermart/internal/auth"
	"github.com/Okenamay/gophermart/internal/config"
	"go.uber.org/zap"
)

type contextKey string

// UserIDContextKey - ключ для доступа к ID пользователя в контексте запроса
const UserIDContextKey = contextKey("userID")

// Authenticator - middleware для проверки JWT-токена
func Authenticator(conf *config.Cfg, logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Warn("Authorization header is missing")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			headerParts := strings.Split(authHeader, " ")
			if len(headerParts) != 2 || headerParts[0] != "Bearer" {
				logger.Warn("Invalid Authorization header format")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tokenString := headerParts[1]
			userID, err := auth.GetUserIDFromToken(conf, tokenString)
			if err != nil {
				logger.Warnw("Invalid token", "error", err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Сохраняем ID пользователя в контексте для последующих хендлеров
			ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
