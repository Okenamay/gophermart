package handlers

import (
	"context"

	"github.com/Okenamay/gophermart/internal/storage/database"
	"github.com/stretchr/testify/mock"
)

// MockStorage - мок хранилища БД для тестов хендлеров
type MockStorage struct {
	mock.Mock
}

// MockStorage внедряет интерфейс Storage
var _ Storage = (*MockStorage)(nil)

func (m *MockStorage) CreateUser(ctx context.Context, login, passwordHash string) (int, error) {
	args := m.Called(ctx, login, passwordHash)
	return args.Int(0), args.Error(1)
}

func (m *MockStorage) GetUserByLogin(ctx context.Context, login string) (*database.User, error) {
	args := m.Called(ctx, login)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*database.User), args.Error(1)
}

func (m *MockStorage) CreateOrder(ctx context.Context, userID int, orderNumber string) error {
	args := m.Called(ctx, userID, orderNumber)
	return args.Error(0)
}

func (m *MockStorage) GetOrderByNumber(ctx context.Context, number string) (*database.Order, error) {
	args := m.Called(ctx, number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*database.Order), args.Error(1)
}

func (m *MockStorage) GetOrdersByUser(ctx context.Context, userID int) ([]database.Order, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]database.Order), args.Error(1)
}

func (m *MockStorage) GetUserBalance(ctx context.Context, userID int) (*database.Balance, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*database.Balance), args.Error(1)
}

func (m *MockStorage) CreateWithdrawal(ctx context.Context, userID int, orderNumber string, sum float64) error {
	args := m.Called(ctx, userID, orderNumber, sum)
	return args.Error(0)
}

func (m *MockStorage) GetWithdrawalsByUser(ctx context.Context, userID int) ([]database.Withdrawal, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]database.Withdrawal), args.Error(1)
}
