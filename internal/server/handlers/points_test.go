package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Okenamay/gophermart/internal/config"
	"github.com/Okenamay/gophermart/internal/server/middleware"
	"github.com/Okenamay/gophermart/internal/storage/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestPointsBalance(t *testing.T) {
	cfg := &config.Cfg{}
	appLogger := zap.NewNop().Sugar()
	userID := 123

	t.Run("successful balance retrieval", func(t *testing.T) {
		mockStorage := new(MockStorage)
		handler := New(cfg, mockStorage, appLogger)

		req, err := http.NewRequest(http.MethodGet, "/api/user/balance", nil)
		require.NoError(t, err)

		ctx := context.WithValue(req.Context(), middleware.UserIDContextKey, userID)
		req = req.WithContext(ctx)

		mockBalance := &database.Balance{Current: 500.5, Withdrawn: 42}
		mockStorage.On("GetUserBalance", mock.Anything, userID).Return(mockBalance, nil)

		rr := httptest.NewRecorder()
		handler.PointsBalance(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var balance Balance
		err = json.Unmarshal(rr.Body.Bytes(), &balance)
		require.NoError(t, err)
		assert.Equal(t, mockBalance.Current, balance.Current)
		assert.Equal(t, mockBalance.Withdrawn, balance.Withdrawn)
		mockStorage.AssertExpectations(t)
	})
}
