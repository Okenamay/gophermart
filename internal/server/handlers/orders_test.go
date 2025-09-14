package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Okenamay/gophermart/internal/config"
	"github.com/Okenamay/gophermart/internal/server/middleware"
	"github.com/Okenamay/gophermart/internal/storage/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAddOrder(t *testing.T) {
	cfg := &config.Cfg{}
	appLogger := zap.NewNop().Sugar()
	userID := 123

	t.Run("successful order upload", func(t *testing.T) {
		mockStorage := new(MockStorage)
		handler := New(cfg, mockStorage, appLogger)
		orderNumber := "9278923470"

		req, err := http.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString(orderNumber))
		require.NoError(t, err)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, userID)
		req = req.WithContext(ctx)

		mockStorage.On("GetOrderByNumber", mock.Anything, orderNumber).Return(nil, database.ErrOrderNotFound)
		mockStorage.On("CreateOrder", mock.Anything, userID, orderNumber).Return(nil)

		rr := httptest.NewRecorder()
		handler.AddOrder(rr, req)

		assert.Equal(t, http.StatusAccepted, rr.Code)
		mockStorage.AssertExpectations(t)
	})

	t.Run("order already uploaded by self", func(t *testing.T) {
		mockStorage := new(MockStorage)
		handler := New(cfg, mockStorage, appLogger)
		orderNumber := "9278923470"

		req, err := http.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString(orderNumber))
		require.NoError(t, err)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, userID)
		req = req.WithContext(ctx)

		existingOrder := &database.Order{UserID: userID}
		mockStorage.On("GetOrderByNumber", mock.Anything, orderNumber).Return(existingOrder, nil)

		rr := httptest.NewRecorder()
		handler.AddOrder(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockStorage.AssertExpectations(t)
	})

	t.Run("order uploaded by another user", func(t *testing.T) {
		mockStorage := new(MockStorage)
		handler := New(cfg, mockStorage, appLogger)
		orderNumber := "9278923470"

		req, err := http.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString(orderNumber))
		require.NoError(t, err)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, userID)
		req = req.WithContext(ctx)

		existingOrder := &database.Order{UserID: 456}
		mockStorage.On("GetOrderByNumber", mock.Anything, orderNumber).Return(existingOrder, nil)

		rr := httptest.NewRecorder()
		handler.AddOrder(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)
		mockStorage.AssertExpectations(t)
	})

	t.Run("invalid order number format", func(t *testing.T) {
		handler := New(cfg, nil, appLogger)
		orderNumber := "123"

		req, err := http.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBufferString(orderNumber))
		require.NoError(t, err)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, userID)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		handler.AddOrder(rr, req)

		assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)
	})
}

func TestListOrders(t *testing.T) {
	cfg := &config.Cfg{}
	appLogger := zap.NewNop().Sugar()
	userID := 123

	t.Run("user has orders", func(t *testing.T) {
		mockStorage := new(MockStorage)
		handler := New(cfg, mockStorage, appLogger)

		req, err := http.NewRequest(http.MethodGet, "/api/user/orders", nil)
		require.NoError(t, err)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, userID)
		req = req.WithContext(ctx)

		accrual := 500.0
		mockOrders := []database.Order{
			{Number: "123", Status: "PROCESSED", Accrual: &accrual, UploadedAt: time.Now()},
		}
		mockStorage.On("GetOrdersByUser", mock.Anything, userID).Return(mockOrders, nil)

		rr := httptest.NewRecorder()
		handler.ListOrders(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
		var orders []Order
		err = json.Unmarshal(rr.Body.Bytes(), &orders)
		require.NoError(t, err)
		assert.Len(t, orders, 1)
		assert.Equal(t, "123", orders[0].Number)
		mockStorage.AssertExpectations(t)
	})

	t.Run("user has no orders", func(t *testing.T) {
		mockStorage := new(MockStorage)
		handler := New(cfg, mockStorage, appLogger)

		req, err := http.NewRequest(http.MethodGet, "/api/user/orders", nil)
		require.NoError(t, err)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, userID)
		req = req.WithContext(ctx)

		mockStorage.On("GetOrdersByUser", mock.Anything, userID).Return(nil, database.ErrNoOrdersFound)

		rr := httptest.NewRecorder()
		handler.ListOrders(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		mockStorage.AssertExpectations(t)
	})
}
