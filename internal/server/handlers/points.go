package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Okenamay/gophermart/internal/server/middleware"
)

// PointsBalance возвращает текущий баланс пользователя
func (h *Handler) PointsBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int)
	if !ok {
		h.appLogger.Errorw("Failed to access User ID", "userID", userID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	balance, err := h.DB.GetUserBalance(r.Context(), userID)
	if err != nil {
		h.appLogger.Errorw("Failed to get user balance", "userID", userID, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Balance{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	})
}
