package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Okenamay/gophermart/internal/luhn"
	"github.com/Okenamay/gophermart/internal/server/middleware"
	"github.com/Okenamay/gophermart/internal/storage/database"
)

// AddOrder обрабатывает загрузку номера заказа.
func (h *Handler) AddOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int)
	if !ok {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

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

	existingOrder, err := h.DB.GetOrderByNumber(r.Context(), orderNumber)
	if err != nil && !errors.Is(err, database.ErrOrderNotFound) {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if existingOrder != nil {
		if existingOrder.UserID == userID {
			w.WriteHeader(http.StatusOK) // Заказ уже загружен этим пользователем
		} else {
			http.Error(w, "Order already uploaded by another user", http.StatusConflict)
		}
		return
	}

	err = h.DB.CreateOrder(r.Context(), userID, orderNumber)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// ListOrders возвращает список заказов пользователя.
func (h *Handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int)
	if !ok {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	ordersDB, err := h.DB.GetOrdersByUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, database.ErrNoOrdersFound) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	ordersAPI := make([]Order, len(ordersDB))
	for i, o := range ordersDB {
		ordersAPI[i] = Order{
			Number:     o.Number,
			Status:     o.Status,
			UploadedAt: o.UploadedAt,
		}
		if o.Accrual != nil {
			ordersAPI[i].Accrual = *o.Accrual
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ordersAPI)
}
