package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Okenamay/gophermart/internal/config"
	logger "github.com/Okenamay/gophermart/internal/logger/zap"
	"github.com/Okenamay/gophermart/internal/luhn"
)

// PointsWithdraw обрабатывает запрос на списание баллов.
func PointsWithdraw(conf *config.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Zap.Info("PointsWithdraw handler started")

		var req WithdrawalRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}

		if !luhn.IsValidString(req.Order) {
			http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
			return
		}

		// TODO: Добавить логику списания.
		// 1. Получить userID из контекста.
		// 2. Проверить баланс пользователя.
		// 3. Если средств недостаточно, вернуть 402 Payment Required.
		// 4. Сохранить операцию списания в БД.

		w.WriteHeader(http.StatusOK)
	}
}

// ListWithdrawals возвращает историю списаний пользователя.
func ListWithdrawals(conf *config.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Zap.Info("ListWithdrawals handler started")

		// TODO: Добавить логику получения истории списаний из БД.
		// 1. Получить userID из контекста.
		// 2. Получить все списания пользователя, отсортированные по времени.
		// 3. Если списаний нет, вернуть 204 No Content.

		// Пример ответа
		withdrawals := []Withdrawal{
			{
				Order:       "2377225624",
				Sum:         500,
				ProcessedAt: time.Now().Add(-48 * time.Hour),
			},
		}

		if len(withdrawals) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(withdrawals)
	}
}
