package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Okenamay/gophermart/internal/auth"
	logger "github.com/Okenamay/gophermart/internal/logger/zap"
	"github.com/Okenamay/gophermart/internal/storage/database"
)

// RegisterUser осуществляет регистрацию нового пользователя
func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var creds UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if creds.Login == "" || creds.Password == "" {
		http.Error(w, "Login and password must not be empty", http.StatusBadRequest)
		return
	}

	passwordHash, err := auth.HashPassword(creds.Password)
	if err != nil {
		logger.Zap.Errorw("Failed to hash password", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	userID, err := h.DB.CreateUser(r.Context(), creds.Login, passwordHash)
	if err != nil {
		if errors.Is(err, database.ErrLoginConflict) {
			http.Error(w, "Login already exists", http.StatusConflict)
			return
		}
		logger.Zap.Errorw("Failed to create user in DB", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	token, err := auth.BuildJWTString(h.Config, userID)
	if err != nil {
		logger.Zap.Errorw("Failed to build JWT", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}

// LoginUser осуществляет авторизацию пользователей
func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var creds UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	user, err := h.DB.GetUserByLogin(r.Context(), creds.Login)
	if err != nil {
		if errors.Is(err, database.ErrUserNotFound) {
			http.Error(w, "Invalid login or password", http.StatusUnauthorized)
			return
		}
		logger.Zap.Errorw("Failed to get user from DB", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !auth.CheckPasswordHash(creds.Password, user.PasswordHash) {
		http.Error(w, "Invalid login or password", http.StatusUnauthorized)
		return
	}

	token, err := auth.BuildJWTString(h.Config, user.ID)
	if err != nil {
		logger.Zap.Errorw("Failed to build JWT", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}
