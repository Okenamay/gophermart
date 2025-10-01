package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Okenamay/gophermart/internal/auth"
	"github.com/Okenamay/gophermart/internal/config"
	"github.com/Okenamay/gophermart/internal/storage/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRegisterUser(t *testing.T) {
	cfg := &config.Cfg{AuthorizationKey: "test-secret", TokenExpiry: 24}
	appLogger := zap.NewNop().Sugar()

	t.Run("successful registration", func(t *testing.T) {
		mockStorage := new(MockStorage)
		handler := New(cfg, mockStorage, appLogger)

		creds := UserCredentials{Login: "test", Password: "password"}
		body, _ := json.Marshal(creds)

		req, err := http.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBuffer(body))
		require.NoError(t, err)

		mockStorage.On("CreateUser", mock.Anything, creds.Login, mock.AnythingOfType("string")).Return(1, nil)

		rr := httptest.NewRecorder()
		handler.RegisterUser(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NotEmpty(t, rr.Header().Get("Authorization"))
		mockStorage.AssertExpectations(t)
	})

	t.Run("login conflict", func(t *testing.T) {
		mockStorage := new(MockStorage)
		handler := New(cfg, mockStorage, appLogger)

		creds := UserCredentials{Login: "exists", Password: "password"}
		body, _ := json.Marshal(creds)

		req, err := http.NewRequest(http.MethodPost, "/api/user/register", bytes.NewBuffer(body))
		require.NoError(t, err)

		mockStorage.On("CreateUser", mock.Anything, creds.Login, mock.AnythingOfType("string")).Return(0, database.ErrLoginConflict)

		rr := httptest.NewRecorder()
		handler.RegisterUser(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)
		mockStorage.AssertExpectations(t)
	})
}

func TestLoginUser(t *testing.T) {
	cfg := &config.Cfg{AuthorizationKey: "test-secret", TokenExpiry: 24}
	appLogger := zap.NewNop().Sugar()
	hashedPassword, _ := auth.HashPassword("password")

	t.Run("successful login", func(t *testing.T) {
		mockStorage := new(MockStorage)
		handler := New(cfg, mockStorage, appLogger)

		creds := UserCredentials{Login: "test", Password: "password"}
		body, _ := json.Marshal(creds)

		req, err := http.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBuffer(body))
		require.NoError(t, err)

		mockUser := &database.User{ID: 1, Login: "test", PasswordHash: hashedPassword}
		mockStorage.On("GetUserByLogin", mock.Anything, creds.Login).Return(mockUser, nil)

		rr := httptest.NewRecorder()
		handler.LoginUser(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NotEmpty(t, rr.Header().Get("Authorization"))
		mockStorage.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockStorage := new(MockStorage)
		handler := New(cfg, mockStorage, appLogger)

		creds := UserCredentials{Login: "notfound", Password: "password"}
		body, _ := json.Marshal(creds)

		req, err := http.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBuffer(body))
		require.NoError(t, err)

		mockStorage.On("GetUserByLogin", mock.Anything, creds.Login).Return((*database.User)(nil), database.ErrUserNotFound)

		rr := httptest.NewRecorder()
		handler.LoginUser(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		mockStorage.AssertExpectations(t)
	})

	t.Run("wrong password", func(t *testing.T) {
		mockStorage := new(MockStorage)
		handler := New(cfg, mockStorage, appLogger)

		creds := UserCredentials{Login: "test", Password: "wrongpassword"}
		body, _ := json.Marshal(creds)

		req, err := http.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBuffer(body))
		require.NoError(t, err)

		mockUser := &database.User{ID: 1, Login: "test", PasswordHash: hashedPassword}
		mockStorage.On("GetUserByLogin", mock.Anything, creds.Login).Return(mockUser, nil)

		rr := httptest.NewRecorder()
		handler.LoginUser(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		mockStorage.AssertExpectations(t)
	})
}
