package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/Okenamay/gophermart/internal/config"
	logger "github.com/Okenamay/gophermart/internal/logger/zap"
	"github.com/Okenamay/gophermart/internal/luhn"
)

// AddOrder обрабатывает загрузку номера заказа.
func AddOrder(conf *config.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Zap.Info("AddOrder handler started")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Invalid request format", http.StatusBadRequest)
			return
		}
		orderNumber := string(body)

		if !luhn.IsValidString(orderNumber) {
			http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
			return
		}

		// TODO: Добавить логику обработки заказа.
		// 1. Получить userID из контекста.
		// 2. Проверить, не загружал ли этот пользователь заказ ранее (вернуть 200 OK).
		// 3. Проверить, не загружал ли другой пользователь этот заказ (вернуть 409 Conflict).
		// 4. Если всё в порядке, сохранить заказ со статусом NEW (вернуть 202 Accepted).

		w.WriteHeader(http.StatusAccepted)
	}
}

// ListOrders возвращает список заказов пользователя.
func ListOrders(conf *config.Cfg) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Zap.Info("ListOrders handler started")

		// TODO: Добавить логику получения заказов из БД.
		// 1. Получить userID из контекста.
		// 2. Получить все заказы пользователя из БД, отсортированные по времени загрузки.
		// 3. Если заказов нет, вернуть 204 No Content.

		// Пример ответа
		orders := []Order{
			{
				Number:     "9278923470",
				Status:     "PROCESSED",
				Accrual:    500,
				UploadedAt: time.Now().Add(-24 * time.Hour),
			},
			{
				Number:     "12345678903",
				Status:     "PROCESSING",
				UploadedAt: time.Now().Add(-12 * time.Hour),
			},
		}

		if len(orders) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(orders)
	}
}
