package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	logger "github.com/Okenamay/gophermart/internal/logger/zap"
	"github.com/Okenamay/gophermart/internal/luhn"
	"github.com/Okenamay/gophermart/internal/server/middleware"
	"github.com/Okenamay/gophermart/internal/storage/database"
)

// PointsWithdraw обрабатывает запрос на списание баллов
func (h *Handler) PointsWithdraw(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int)
	if !ok {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var req WithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if !luhn.IsValidString(req.Order) {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	err := h.DB.CreateWithdrawal(r.Context(), userID, req.Order, req.Sum)
	if err != nil {
		if errors.Is(err, database.ErrInsufficientFunds) {
			http.Error(w, "Insufficient funds", http.StatusPaymentRequired)
			return
		}
		logger.Zap.Errorw("Failed to create withdrawal", "userID", userID, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ListWithdrawals возвращает историю списаний пользователя
func (h *Handler) ListWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int)
	if !ok {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	withdrawalsDB, err := h.DB.GetWithdrawalsByUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, database.ErrNoWithdrawalsFound) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		logger.Zap.Errorw("Failed to get withdrawals", "userID", userID, "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Преобразуем модель БД в модель API
	withdrawalsAPI := make([]Withdrawal, len(withdrawalsDB))
	for i, w := range withdrawalsDB {
		withdrawalsAPI[i] = Withdrawal{
			Order:       w.OrderNumber,
			Sum:         w.Sum,
			ProcessedAt: w.ProcessedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(withdrawalsAPI)
}
