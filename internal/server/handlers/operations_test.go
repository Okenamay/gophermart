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

func TestPointsWithdraw(t *testing.T) {
	cfg := &config.Cfg{}
	appLogger := zap.NewNop().Sugar()
	userID := 123

	t.Run("successful withdrawal", func(t *testing.T) {
		mockStorage := new(MockStorage)
		handler := New(cfg, mockStorage, appLogger)

		reqBody := WithdrawalRequest{Order: "2377225624", Sum: 100}
		body, _ := json.Marshal(reqBody)
		req, err := http.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBuffer(body))
		require.NoError(t, err)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, userID)
		req = req.WithContext(ctx)

		mockStorage.On("CreateWithdrawal", mock.Anything, userID, reqBody.Order, reqBody.Sum).Return(nil)

		rr := httptest.NewRecorder()
		handler.PointsWithdraw(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockStorage.AssertExpectations(t)
	})

	t.Run("insufficient funds", func(t *testing.T) {
		mockStorage := new(MockStorage)
		handler := New(cfg, mockStorage, appLogger)

		reqBody := WithdrawalRequest{Order: "2377225624", Sum: 1000}
		body, _ := json.Marshal(reqBody)
		req, err := http.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBuffer(body))
		require.NoError(t, err)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, userID)
		req = req.WithContext(ctx)

		mockStorage.On("CreateWithdrawal", mock.Anything, userID, reqBody.Order, reqBody.Sum).Return(database.ErrInsufficientFunds)

		rr := httptest.NewRecorder()
		handler.PointsWithdraw(rr, req)

		assert.Equal(t, http.StatusPaymentRequired, rr.Code)
		mockStorage.AssertExpectations(t)
	})
}

func TestListWithdrawals(t *testing.T) {
	cfg := &config.Cfg{}
	appLogger := zap.NewNop().Sugar()
	userID := 123

	t.Run("user has withdrawals", func(t *testing.T) {
		mockStorage := new(MockStorage)
		handler := New(cfg, mockStorage, appLogger)

		req, err := http.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
		require.NoError(t, err)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, userID)
		req = req.WithContext(ctx)

		mockData := []database.Withdrawal{
			{OrderNumber: "123", Sum: 100, ProcessedAt: time.Now()},
		}
		mockStorage.On("GetWithdrawalsByUser", mock.Anything, userID).Return(mockData, nil)

		rr := httptest.NewRecorder()
		handler.ListWithdrawals(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var withdrawals []Withdrawal
		err = json.Unmarshal(rr.Body.Bytes(), &withdrawals)
		require.NoError(t, err)
		assert.Len(t, withdrawals, 1)
		mockStorage.AssertExpectations(t)
	})

	t.Run("user has no withdrawals", func(t *testing.T) {
		mockStorage := new(MockStorage)
		handler := New(cfg, mockStorage, appLogger)

		req, err := http.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
		require.NoError(t, err)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, userID)
		req = req.WithContext(ctx)

		mockStorage.On("GetWithdrawalsByUser", mock.Anything, userID).Return(nil, database.ErrNoWithdrawalsFound)

		rr := httptest.NewRecorder()
		handler.ListWithdrawals(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		mockStorage.AssertExpectations(t)
	})
}
