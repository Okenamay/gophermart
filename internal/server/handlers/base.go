package handlers

import (
	"context"

	"github.com/Okenamay/gophermart/internal/config"
	"github.com/Okenamay/gophermart/internal/storage/database"
)

// Storage определяет интерфейс для всех операций с БД, которые нужны хендлерам
type Storage interface {
	CreateUser(ctx context.Context, login, passwordHash string) (int, error)
	GetUserByLogin(ctx context.Context, login string) (*database.User, error)
	CreateOrder(ctx context.Context, userID int, orderNumber string) error
	GetOrderByNumber(ctx context.Context, number string) (*database.Order, error)
	GetOrdersByUser(ctx context.Context, userID int) ([]database.Order, error)
	GetUserBalance(ctx context.Context, userID int) (*database.Balance, error)
	CreateWithdrawal(ctx context.Context, userID int, orderNumber string, sum float64) error
	GetWithdrawalsByUser(ctx context.Context, userID int) ([]database.Withdrawal, error)
}

// Handler - структура для хранения зависимостей хендлеров
type Handler struct {
	Config *config.Cfg
	DB     Storage
}

// New создает новый экземпляр Handler
func New(cfg *config.Cfg, db Storage) *Handler {
	return &Handler{
		Config: cfg,
		DB:     db,
	}
}
