package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Okenamay/gophermart/internal/config"
	logger "github.com/Okenamay/gophermart/internal/logger/zap"
)

// PointsBalance возвращает текущий баланс пользователя.
func PointsBalance(conf *config.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Zap.Info("PointsBalance handler started")

		// TODO: Добавить логику получения баланса из БД.
		// 1. Получить userID из контекста.
		// 2. Рассчитать текущий баланс (сумма начислений).
		// 3. Рассчитать сумму списанных баллов.

		balance := Balance{
			Current:   500.5,
			Withdrawn: 42,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(balance)
	}
}
