package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Okenamay/gophermart/internal/config"
	logger "github.com/Okenamay/gophermart/internal/logger/zap"
)

// RegUser обрабатывает регистрацию нового пользователя.
func RegUser(conf *config.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Zap.Info("RegUser handler started")

		var creds UserCredentials
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		// TODO: Добавить логику сохранения пользователя в БД.
		// 1. Проверить, что логин не занят (вернуть 409 Conflict).
		// 2. Захешировать пароль.
		// 3. Сохранить пользователя.
		// 4. Сгенерировать JWT и установить его в заголовок Authorization или cookie.

		w.Header().Set("Content-Type", "application/json")
		// Пример успешного ответа
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"message": "user registered and authenticated"}`)
	}
}

// LoginUser обрабатывает аутентификацию пользователя.
func LoginUser(conf *config.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Zap.Info("LoginUser handler started")

		var creds UserCredentials
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		// TODO: Добавить логику аутентификации.
		// 1. Найти пользователя по логину.
		// 2. Сравнить хеш пароля.
		// 3. Если всё верно (вернуть 200 OK), сгенерировать JWT и установить его.
		// 4. Если пара логин/пароль неверна, вернуть 401 Unauthorized.

		w.Header().Set("Content-Type", "application/json")
		// Пример успешного ответа
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"message": "user authenticated"}`)
	}
}
